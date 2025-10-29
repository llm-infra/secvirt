package desktop

import (
	"context"
	"net/http"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/mel2oo/go-dkit/otel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"trpc.group/trpc-go/trpc-a2a-go/client"
)

var defaultGeminiPort = 8003

type Sandbox struct {
	*sandbox.Sandbox
}

func NewSandbox(ctx context.Context, opts ...sandbox.Option) (*Sandbox, error) {
	client, err := sandbox.NewSandbox(ctx,
		append(opts,
			sandbox.WithTemplate(sandbox.TemplateDesktop),
			sandbox.WithHealthPorts([]int{defaultGeminiPort}))...)
	if err != nil {
		return nil, err
	}

	return &Sandbox{Sandbox: client}, nil
}

func (s *Sandbox) NewA2AClient(ctx context.Context) (*client.A2AClient, error) {
	tr := otelhttp.NewTransport(
		spec.NewHeaderRoundTripper(
			spec.GenSandboxHeader(defaultGeminiPort, s.Name, ""),
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
