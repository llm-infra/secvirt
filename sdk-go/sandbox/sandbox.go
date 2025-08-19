package sandbox

import (
	"context"
	"fmt"

	"github.com/dubonzi/otelresty"
	"github.com/go-resty/resty/v2"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/filesystem"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/mel2oo/go-dkit/ext"
	"github.com/mel2oo/go-dkit/otel"
)

var (
	defaultAPIPort   = 8994
	defaultProxyPort = 8993
)

type Sandbox struct {
	*SandboxDetail

	api   *resty.Client // API客户端
	proxy *resty.Client // 沙箱代理客户端
	fs    *filesystem.Filesystem
	cmd   *commands.Cmd
	pty   *commands.Pty
}

func NewSandbox(ctx context.Context, opts ...Option) (*Sandbox, error) {
	opt := newOptions()
	for _, o := range opts {
		o(opt)
	}

	apiBaseUrl := fmt.Sprintf("http://%s:%d", opt.host, defaultAPIPort)
	apiClient := resty.New()
	apiClient.SetBaseURL(apiBaseUrl)
	if opt.useTelemetry {
		otelresty.TraceClient(apiClient,
			otelresty.WithSpanNameFormatter(otel.RestySpanNameFormatter),
			otelresty.WithTracerProvider(opt.tracerProvider),
			otelresty.WithPropagators(opt.propagators),
		)
	}

	prxBaseUrl := fmt.Sprintf("http://%s:%d", opt.host, defaultProxyPort)
	prxClient := resty.New()
	prxClient.SetBaseURL(prxBaseUrl)
	if opt.useTelemetry {
		otelresty.TraceClient(prxClient,
			otelresty.WithSpanNameFormatter(otel.RestySpanNameFormatter),
			otelresty.WithTracerProvider(opt.tracerProvider),
			otelresty.WithPropagators(opt.propagators),
		)
	}

	sbx := &Sandbox{
		api:   apiClient,
		proxy: prxClient,
	}

	res, err := sbx._createSandbox(
		ctx,
		opt.user,
		opt.template,
		opt.healthPorts,
	)
	if err != nil {
		return nil, err
	}

	sbx.fs = filesystem.NewFileSystem(prxBaseUrl, res.ID, opt.user)
	sbx.cmd = commands.NewCmd(prxBaseUrl, res.ID, opt.user)
	sbx.pty = commands.NewPty(prxBaseUrl, res.ID, opt.user)
	sbx.SandboxDetail = res
	return sbx, nil
}

func (c *Sandbox) _createSandbox(ctx context.Context, userID string,
	template TemplateType, healthPorts []int) (*SandboxDetail, error) {
	resp, err := c.api.R().
		SetContext(ctx).
		SetBody(map[string]any{
			"user_id":      userID,
			"template":     template,
			"health_ports": healthPorts,
		}).
		SetResult(SandboxDetail{}).
		SetError(ErrorResponse{}).
		Post("/secvirt/v2/sandboxes")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(*ErrorResponse)
	}

	return resp.Result().(*SandboxDetail), nil
}

func (c *Sandbox) ProxyBaseURL() string {
	return c.proxy.BaseURL
}

func (c *Sandbox) ProxyRequest(ctx context.Context, port int) *resty.Request {
	req := c.proxy.R()
	req.SetContext(ctx)
	req.SetHeaders(spec.GenSandboxHeader(port, c.ID, c.User))
	ext.InjectHeader(ctx, req.Header)
	return req
}

func (c *Sandbox) ApiRequest(ctx context.Context) *resty.Request {
	req := c.api.R()
	req.SetContext(ctx)
	ext.InjectHeader(ctx, req.Header)
	return req
}

func (c *Sandbox) Filesystem() *filesystem.Filesystem {
	return c.fs
}

func (c *Sandbox) Cmd() *commands.Cmd {
	return c.cmd
}

func (c *Sandbox) Pty() *commands.Pty {
	return c.pty
}

func (c *Sandbox) GetSandbox(ctx context.Context, sandboxID string) (*SandboxDetail, error) {
	resp, err := c.api.R().
		SetContext(ctx).
		SetResult(SandboxDetail{}).
		SetError(ErrorResponse{}).
		Post("/secvirt/v2/sandboxes/" + sandboxID)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(*ErrorResponse)
	}

	return resp.Result().(*SandboxDetail), nil
}

func (c *Sandbox) StopSandbox(ctx context.Context, sandboxID string) error {
	resp, err := c.api.R().
		SetContext(ctx).
		SetError(ErrorResponse{}).
		Post("/secvirt/v2/sandboxes/" + sandboxID + "/stop")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.Error().(*ErrorResponse)
	}

	return nil
}

func (c *Sandbox) StartSandbox(ctx context.Context, sandboxID string) error {
	resp, err := c.api.R().
		SetContext(ctx).
		SetError(ErrorResponse{}).
		Post("/secvirt/v2/sandboxes/" + sandboxID + "/start")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.Error().(*ErrorResponse)
	}

	return nil
}

func (c *Sandbox) DestroySandbox(ctx context.Context, sandboxID string) error {
	resp, err := c.api.R().
		SetContext(ctx).
		SetError(ErrorResponse{}).
		Post("/secvirt/v2/sandboxes/" + sandboxID + "/destroy")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.Error().(*ErrorResponse)
	}

	return nil
}
