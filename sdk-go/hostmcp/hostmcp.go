package hostmcp

import (
	"context"
	"errors"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	defaultMcpServerPort = 8001
	defaultMcpRouterPort = 8002
)

type Sandbox struct {
	*sandbox.Sandbox
}

func NewSandbox(ctx context.Context, opts ...sandbox.Option) (*Sandbox, error) {
	client, err := sandbox.NewSandbox(ctx,
		append(opts,
			sandbox.WithTemplate(sandbox.TemplateHostMCP),
			sandbox.WithHealthPorts([]int{defaultMcpServerPort, defaultMcpRouterPort}))...)
	if err != nil {
		return nil, err
	}

	return &Sandbox{Sandbox: client}, nil
}

func (s *Sandbox) GetMcpConfig(ctx context.Context, id string) (map[string]ServerEntry, error) {
	type PaginatedResponse struct {
		Data map[string]ServerEntry `json:"clients"`
	}

	resp, err := s.ApiRequest(ctx).
		SetQueryParams(map[string]string{
			"cursor": id,
			"limit":  "1",
		}).
		SetResult(PaginatedResponse{}).
		SetError(sandbox.ErrorResponse{}).
		Get("/secvirt/v2/mcp/clients")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(*sandbox.ErrorResponse)
	}

	result := resp.Result().(*PaginatedResponse)
	if len(result.Data) == 0 {
		return nil, errors.New("mcp config not found")
	}

	return result.Data, nil
}

func (s *Sandbox) MCPs(ctx context.Context) ([]MCPEndpoint, error) {
	resp, err := s.ProxyRequest(ctx, defaultMcpServerPort).
		SetResult([]MCPEndpoint{}).
		SetError(sandbox.ErrorResponse{}).
		Get("/hostmcp/v1/mcps")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(*sandbox.ErrorResponse)
	}

	result := resp.Result().(*[]MCPEndpoint)
	return *result, nil
}

func (s *Sandbox) Launch(
	ctx context.Context,
	cfg *ServersFile,
	reload bool,
) ([]MCPEndpoint, error) {
	resp, err := s.ProxyRequest(ctx, defaultMcpServerPort).
		SetBody(map[string]any{
			"config": cfg,
			"reload": reload,
		}).
		SetResult([]MCPEndpoint{}).
		SetError(sandbox.ErrorResponse{}).
		Post("/hostmcp/v1/launch")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(*sandbox.ErrorResponse)
	}

	result := resp.Result().(*[]MCPEndpoint)
	if len(*result) == 0 {
		return nil, errors.New("failed lanuch mcp server")
	}

	return *result, err
}

func (s *Sandbox) Connect(
	ctx context.Context,
	endpoint MCPEndpoint,
) (*client.Client, *mcp.InitializeResult, error) {
	client, err := client.NewStreamableHttpClient(
		s.ProxyBaseURL()+endpoint.Path+"mcp",
		transport.WithHTTPHeaders(
			spec.GenSandboxHeader(defaultMcpRouterPort, s.ID, ""),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	if err := client.Start(ctx); err != nil {
		return nil, nil, err
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "sdk-test",
		Version: "v1.0.0",
	}

	initResult, err := client.Initialize(ctx, initRequest)
	return client, initResult, err
}
