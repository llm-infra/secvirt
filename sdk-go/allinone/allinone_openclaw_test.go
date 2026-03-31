package allinone

import (
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

func TestParseOpenClawPIDs(t *testing.T) {
	pids, err := parseOpenClawPIDs("101\n202\n202\n303\n")
	require.NoError(t, err)
	assert.Equal(t, []uint32{101, 202, 303}, pids)
}

func TestParseOpenClawPIDsEmpty(t *testing.T) {
	pids, err := parseOpenClawPIDs("")
	require.NoError(t, err)
	assert.Empty(t, pids)
}

func TestParseOpenClawPIDsInvalid(t *testing.T) {
	_, err := parseOpenClawPIDs("101\nabc\n")
	require.Error(t, err)
}
