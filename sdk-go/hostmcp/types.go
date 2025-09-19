package hostmcp

import (
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	EntryTypeStdio          = "stdio"
	EntryTypeSSE            = "sse"
	EntryTypeStreamableHTTP = "streamable-http"
)

type ServersFile struct {
	McpServers map[string]ServerEntry `json:"mcpServers,omitempty"`
}

type ServerEntry struct {
	Type string `json:"type,omitempty"`

	// STDIO
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	ENV     map[string]string `json:"env,omitempty"`

	// SSE/HTTP
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`
}

type MCPEndpoint struct {
	Name  string      `json:"name,omitempty"`
	Path  string      `json:"path,omitempty"`
	Entry ServerEntry `json:"entry,omitempty"`
	Tools []mcp.Tool  `json:"tools"`
}
