package allinone

import (
	"fmt"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/desktop"
	"github.com/llm-infra/secvirt/sdk-go/desktop/codex"
	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/mel2oo/go-dkit/json"
	"github.com/stretchr/testify/assert"
)

func TestCodexChat(t *testing.T) {
	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("192.168.134.142"),
		sandbox.WithUser("mel2oo"),
	)
	assert.NoError(t, err)
	defer sbx.DestroySandbox(t.Context())

	err = sbx.SetCodexConfig(t.Context(), codex.NewConfig(
		"glm-4.6", &codex.ModelProvider{
			Name:    "secwall",
			BaseURL: "http://192.168.134.142:8995/v1",
			HTTPHeaders: map[string]string{
				"EXT": "x-uid:17;x-org:1802543778616180738;",
			},
		},
	))
	assert.NoError(t, err)

	stream, err := sbx.CodexChat(t.Context(), "你好",
		desktop.WithEnvs(
			map[string]string{codex.APIKeyEnvVar: ""},
		),
	)
	if !assert.NoError(t, err) {
		return
	}
	defer stream.Close()

	for {
		res, err := stream.Recv()
		if err != nil {
			break
		}

		fmt.Println(json.MarshalString(res))
	}
}
