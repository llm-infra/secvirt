package allinone

import (
	"fmt"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/desktop"
	"github.com/llm-infra/secvirt/sdk-go/desktop/claude"
	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/mel2oo/go-dkit/json"
	"github.com/stretchr/testify/assert"
)

func TestCodeIDE(t *testing.T) {
	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("192.168.134.142"),
		sandbox.WithUser("mel2oo"),
	)
	assert.NoError(t, err)

	res, err := sbx.Packages(t.Context(), "python")
	if assert.NoError(t, err) {
		fmt.Println(res)
	}

	res, err = sbx.Packages(t.Context(), "javascript")
	if assert.NoError(t, err) {
		fmt.Println(res)
	}
}

func TestClaudeChat(t *testing.T) {
	settings := claude.NewSettings()
	settings.SetEnv("ANTHROPIC_BASE_URL", "http://192.168.134.142:8995")
	settings.SetEnv("ANTHROPIC_AUTH_TOKEN", "")
	settings.SetEnv("ANTHROPIC_CUSTOM_HEADERS", "EXT: x-uid:17;x-org:1802543778616180738;")
	settings.SetEnv("ANTHROPIC_MODEL", "glm-4.6")
	settings.SetEnv("CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC", "1")

	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("192.168.134.142"),
		sandbox.WithUser("mel2oo"),
	)
	assert.NoError(t, err)
	defer sbx.DestroySandbox(t.Context())

	err = sbx.SetClaudeSettings(t.Context(), settings)
	assert.NoError(t, err)

	stream, err := sbx.ClaudeChat(t.Context(), "hi",
		desktop.WithStdin(false))
	if !assert.NoError(t, err) {
		return
	}
	defer stream.Close()

	for {
		res, err := stream.Recv()
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Println(json.MarshalString(res))
	}
}
