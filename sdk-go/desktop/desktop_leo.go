package desktop

import (
	"context"
	"fmt"

	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
)

func (s *Sandbox) LeoChat(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[LeoEvent], error) {
	opt := &Options{
		cwd: s.HomeDir(),
	}

	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("leo -p '%s'", content),
		opt.envs,
		opt.cwd,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream[LeoEvent](ctx, handle, &LeoEvent{}), nil
}

type LeoEvent struct{}

func (e *LeoEvent) Decode(data []byte, evt any) error {
	return nil
}
