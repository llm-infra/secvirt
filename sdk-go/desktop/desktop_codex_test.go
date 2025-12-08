package desktop

import (
	"fmt"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/mel2oo/go-dkit/json"
	"github.com/stretchr/testify/assert"
)

func TestCodexChat(t *testing.T) {
	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("10.50.10.18"),
	)
	assert.NoError(t, err)
	defer sbx.DestroySandbox(t.Context())

	stream, err := sbx.CodexChat(t.Context(), "你好",
		WithEnvs(
			map[string]string{
				"HTTP_PROXY":  "",
				"HTTPS_PROXY": "",
			},
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
