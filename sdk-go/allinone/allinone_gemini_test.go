package allinone

import (
	"fmt"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/desktop/gemini"
	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/mel2oo/go-dkit/ext"
	"github.com/mel2oo/go-dkit/json"
	"github.com/stretchr/testify/assert"
)

func TestGeminiChat(t *testing.T) {
	ctx := ext.WithContextValue(t.Context(), ext.New("x-uid:17;x-org:1802543778616180738;"))

	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("192.168.134.142"),
		sandbox.WithUser("mel2oo"),
	)
	assert.NoError(t, err)
	defer sbx.DestroySandbox(t.Context())

	err = sbx.SetGeminiConfig(t.Context(),
		gemini.NewConfig(ctx, "glm-4.6", "http://192.168.134.142:8995"))
	assert.NoError(t, err)

	stream, err := sbx.GeminiChat(ctx, "你好")
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
