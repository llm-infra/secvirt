package hostmcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestMCPs(t *testing.T) {
	sbx, err := NewSandbox(
		context.TODO(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)

	res, err := sbx.MCPs(context.TODO())
	assert.NoError(t, err)

	fmt.Println(res)
}

func TestMCPLaunch(t *testing.T) {
	sbx, err := NewSandbox(
		context.TODO(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)

	cli, err := sbx.Launch(context.TODO(), &ServersFile{
		McpServers: map[string]ServerEntry{
			"duck-mcp": {
				Type:    "stdio",
				Command: "duckduckgo-mcp-server",
			},
		},
	}, false)
	assert.NoError(t, err)

	tools, err := cli.ListTools(context.Background(), mcp.ListToolsRequest{})
	if assert.NoError(t, err) {
		for _, v := range tools.Tools {
			fmt.Println("tool:", v.Name)
		}
	}
}
