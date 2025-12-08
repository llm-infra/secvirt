package desktop

import (
	"context"
	"fmt"

	"github.com/llm-infra/secvirt/sdk-go/desktop/codex"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
)

func (s *Sandbox) CodexChat(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[codex.Event], error) {
	opt := &Options{
		cwd: s.HomeDir(),
	}

	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("codex exec '%s' --skip-git-repo-check --full-auto --json", content),
		opt.envs,
		opt.cwd,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream[codex.Event](ctx, handle, &codex.Decoder{}), nil
}
