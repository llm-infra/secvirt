package hostmcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestGetMcpConfig(t *testing.T) {
	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)

	res, err := sbx.GetMcpConfig(
		t.Context(),
		"790b558b-188e-2b28-9064-53a51c187a3b",
	)
	assert.NoError(t, err)

	fmt.Println(res)
}

func TestMCPs(t *testing.T) {
	sbx, err := NewSandbox(
		context.TODO(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)

	res, err := sbx.GetLaunchMCPs(context.TODO())
	assert.NoError(t, err)

	fmt.Println(res)
}

func TestMCPLaunch(t *testing.T) {
	sbx, err := NewSandbox(
		context.TODO(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)

	mcps, err := sbx.Launch(context.TODO(), &ServersFile{
		McpServers: map[string]ServerEntry{
			"duck-mcp": {
				Type:    "stdio",
				Command: "duckduckgo-mcp-server",
			},
		},
	}, false)
	assert.NoError(t, err)

	cli, _, err := sbx.Connect(t.Context(), mcps[0])
	assert.NoError(t, err)

	tools, err := cli.ListTools(context.Background(), mcp.ListToolsRequest{})
	if assert.NoError(t, err) {
		for _, v := range tools.Tools {
			fmt.Println("tool:", v.Name)
		}
	}
}
