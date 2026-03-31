package allinone

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/a3tai/openclaw-go/chatcompletions"
	"github.com/llm-infra/secvirt/sdk-go/allinone/openclaw"
	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenClawSessionMethodSignatures(t *testing.T) {
	getSessionsType := reflect.TypeOf((*Sandbox).GetSessions)
	if assert.Equal(t, 2, getSessionsType.NumOut()) {
		assert.Equal(t, "*allinone.SessionsListResult", getSessionsType.Out(0).String())
		assert.Equal(t, "error", getSessionsType.Out(1).String())
	}

	getSessionPreviewType := reflect.TypeOf((*Sandbox).GetSession)
	if assert.Equal(t, 2, getSessionPreviewType.NumOut()) {
		assert.Equal(t, "*allinone.SessionPreviewResult", getSessionPreviewType.Out(0).String())
		assert.Equal(t, "error", getSessionPreviewType.Out(1).String())
	}
}

func initSandboxWithClaw(t *testing.T) (*Sandbox, error) {
	logrus.SetLevel(logrus.DebugLevel)

	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("192.168.134.142"),
		sandbox.WithUser("mel2oo"),
	)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	if !assert.NoError(t, sbx.InitOpenClaw(t.Context())) {
		t.Fatal(err)
	}

	return sbx, nil
}

func TestInitOpenClaw(t *testing.T) {
	sbx, err := initSandboxWithClaw(t)
	if !assert.NoError(t, err) {
		return
	}

	res, err := sbx.ChatClient().Create(t.Context(), chatcompletions.Request{
		Model: "openclaw:main",
		Messages: []chatcompletions.Message{
			{Role: "user", Content: "Hello!"},
		},
	})
	fmt.Println(res)

	_, err = sbx.GetModels(t.Context())
	assert.NoError(t, err)

	assert.NoError(t, sbx.SetModel(t.Context(), "model2", openclaw.ModelProvider{
		BaseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
		APIKey:  "29727a7304484c06946b10f0113d0185.fKUVlr04yfQIcgNd",
		API:     openclaw.APIOpenAICompletions,
		Models: []openclaw.ModelConfig{
			{
				ID:   "glm-5",
				Name: "glm-5",
				API:  openclaw.APIOpenAICompletions,
			},
		},
	}))

	assert.NoError(t, sbx.SetDefaultModel(t.Context(), openclaw.ModelRef{
		Primary: "model2/glm-5",
	}))

	res, err = sbx.ChatClient().Create(t.Context(), chatcompletions.Request{
		Model: "openclaw:main",
		Messages: []chatcompletions.Message{
			{Role: "user", Content: "Hello!"},
		},
	})
	fmt.Println(res)
}

func TestRetryOpenClawReadRetriesTransientEOF(t *testing.T) {
	attempts := 0

	data, err := retryOpenClawRead(t.Context(), func() ([]byte, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("protocol error: incomplete envelope: unexpected EOF")
		}
		return []byte("ok"), nil
	})

	require.NoError(t, err)
	assert.Equal(t, []byte("ok"), data)
	assert.Equal(t, 3, attempts)
}

func TestRetryOpenClawReadStopsOnNonRetryableError(t *testing.T) {
	attempts := 0
	wantErr := errors.New("permission denied")

	_, err := retryOpenClawRead(t.Context(), func() ([]byte, error) {
		attempts++
		return nil, wantErr
	})

	require.ErrorIs(t, err, wantErr)
	assert.Equal(t, 1, attempts)
}

func TestRetryFindOpenClawPIDRetriesNotRunning(t *testing.T) {
	attempts := 0

	pid, err := retryFindOpenClawPID(t.Context(), func() (int, error) {
		attempts++
		if attempts < 3 {
			return 0, ErrOpenClawNotRunning
		}
		return 1234, nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1234, pid)
	assert.Equal(t, 3, attempts)
}

func TestRetryFindOpenClawPIDStopsOnNonRetryableError(t *testing.T) {
	attempts := 0
	wantErr := errors.New("command failed")

	_, err := retryFindOpenClawPID(t.Context(), func() (int, error) {
		attempts++
		return 0, wantErr
	})

	require.ErrorIs(t, err, wantErr)
	assert.Equal(t, 1, attempts)
}
