package desktop

import (
	"context"
	"net/http"
	"strconv"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/mel2oo/go-dkit/otel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"trpc.group/trpc-go/trpc-a2a-go/client"
)

type Sandbox struct {
	*sandbox.Sandbox

	leoHandle *commands.CommandHandle
}

func NewSandbox(ctx context.Context, opts ...sandbox.Option) (*Sandbox, error) {
	client, err := sandbox.NewSandbox(ctx,
		append(opts, sandbox.WithTemplate(sandbox.TemplateDesktop))...)
	if err != nil {
		return nil, err
	}

	return &Sandbox{Sandbox: client}, nil
}

const (
	LeoA2AServerBin     = "leo-a2a-server"
	LeoA2AServerEnvPort = "CODER_AGENT_PORT"
)

func (s *Sandbox) NewLeo(ctx context.Context, port int) (*client.A2AClient, error) {
	handle, err := s.Cmd().Start(ctx,
		LeoA2AServerBin,
		map[string]string{LeoA2AServerEnvPort: strconv.Itoa(port)},
		"",
	)
	if err != nil {
		return nil, err
	}

	s.leoHandle = handle
	return s.NewA2AClient(ctx, port)
}

func (s *Sandbox) CloseLeo(ctx context.Context) error {
	if s.leoHandle != nil {
		return s.leoHandle.Kill(ctx)
	}
	return nil
}

func (s *Sandbox) NewA2AClient(ctx context.Context, port int) (*client.A2AClient, error) {
	tr := otelhttp.NewTransport(
		spec.NewHeaderRoundTripper(
			spec.GenSandboxHeader(port, s.Name, ""),
			http.DefaultTransport,
		),
		otelhttp.WithTracerProvider(otel.Standard().TracerProvider),
		otelhttp.WithPropagators(otel.Standard().Propagators),
		otelhttp.WithSpanNameFormatter(otel.HttpSpanNameFormatter),
	)
	httpClient := &http.Client{Transport: tr}

	return client.NewA2AClient(s.ProxyBaseURL(),
		client.WithHTTPClient(httpClient))
}
