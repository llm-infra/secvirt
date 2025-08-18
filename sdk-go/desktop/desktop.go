package desktop

import (
	"context"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
)

type Sandbox struct {
	*sandbox.Sandbox
}

func NewSandbox(ctx context.Context, opts ...sandbox.Option) (*Sandbox, error) {
	client, err := sandbox.NewSandbox(ctx,
		append(opts,
			sandbox.WithTemplate(sandbox.TemplateDesktop))...)
	if err != nil {
		return nil, err
	}

	return &Sandbox{Sandbox: client}, nil
}
