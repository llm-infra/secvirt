package allinone

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/desktop/opencode"
	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/mel2oo/go-dkit/json"
	"github.com/stretchr/testify/assert"
)

func TestOpenCodeChat(t *testing.T) {
	// ctx := ext.WithContextValue(t.Context(), ext.New("x-uid:17;x-org:1802543778616180738;"))

	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("192.168.134.142"),
		sandbox.WithUser("mel2oo"),
	)
	assert.NoError(t, err)
	defer sbx.DestroySandbox(t.Context())

	provider := opencode.NewOpenAIProvider("secwall", "http://192.168.134.142:8995/v1",
		map[string]string{"EXT": "x-uid:17;x-org:1802543778616180738;"},
		[]string{"glm-4.6"},
	)
	config := opencode.NewConfig("secwall/glm-4.6",
		opencode.WithProvider(provider),
		opencode.WithMcp("cvss_calculator", opencode.Mcp{
			Type:    opencode.McpTypeRemote,
			URL:     "http://10.20.152.105:18000/sse",
			Enabled: true,
		}),
	)

	err = sbx.SetOpenCodeConfig(t.Context(), config)
	assert.NoError(t, err)

	// skill
	{
		fi, err := os.Open("pdf.tar.gz")
		if assert.NoError(t, err) {
			err = sbx.SetOpenCodeSkills(t.Context(), map[string]io.Reader{"pdf.tar.gz": fi})
			assert.NoError(t, err)
			fi.Close()
		}
	}

	// agent
	{

	}

	stream, err := sbx.OpenCodeChat(t.Context(), "你好")
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
