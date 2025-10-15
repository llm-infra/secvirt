package hostmcp

import (
	"context"
	"errors"
	"net/http"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/mel2oo/go-dkit/otel"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

func (s *Sandbox) GetLaunchMCPs(ctx context.Context) ([]MCPEndpoint, error) {
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

func (s *Sandbox) Launch(ctx context.Context, cfg *ServersFile, reload bool,
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

func (s *Sandbox) LaunchWithID(ctx context.Context, id string, reload bool,
) ([]MCPEndpoint, error) {
	cfg, err := s.GetMcpConfig(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.Launch(ctx, &ServersFile{McpServers: cfg}, false)
}

func (s *Sandbox) Connect(ctx context.Context, endpoint MCPEndpoint,
) (*mcp.ClientSession, error) {
	tr := otelhttp.NewTransport(
		&headerRoundTripper{
			headers: spec.GenSandboxHeader(defaultMcpRouterPort, s.Name, ""),
			rt:      http.DefaultTransport,
		},
		otelhttp.WithTracerProvider(otel.Standard().TracerProvider),
		otelhttp.WithPropagators(otel.Standard().Propagators),
		otelhttp.WithSpanNameFormatter(otel.HttpSpanNameFormatter),
	)
	httpClient := &http.Client{Transport: tr}

	transport := &mcp.StreamableClientTransport{
		Endpoint:   s.ProxyBaseURL() + endpoint.Path,
		HTTPClient: httpClient,
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "sdk-go",
		Version: "v1.0.0",
	}, nil)

	return client.Connect(ctx, transport, nil)
}

type headerRoundTripper struct {
	headers map[string]string
	rt      http.RoundTripper
}

func (h *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
	return h.rt.RoundTrip(req)
}
