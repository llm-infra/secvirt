package desktop

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/llm-infra/acp/sdk/go/acp"
	"github.com/llm-infra/secvirt/sdk-go/desktop/codex"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
)

func (s *Sandbox) SetCodexConfig(ctx context.Context, config *codex.Config,
	opts ...Option) error {
	opt := NewOptions(s.HomeDir())
	for _, o := range opts {
		o(opt)
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	return s.Filesystem().Write(ctx,
		filepath.Join(opt.cwd, ".codex", "config.toml"),
		data,
	)
}

func (s *Sandbox) CodexChat(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[codex.Message], error) {
	opt := NewOptions(s.HomeDir())
	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("codex exec %s --skip-git-repo-check --full-auto --json",
			strconv.Quote(content)),
		opt.envs,
		opt.cwd,
		opt.stdin,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream(ctx, handle, codex.NewDecoder()), nil
}

func (s *Sandbox) CodexChatWithACPStream(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[[]acp.Event], error) {
	opt := NewOptions(s.HomeDir())
	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("codex exec %s --skip-git-repo-check --full-auto --json",
			strconv.Quote(content)),
		opt.envs,
		opt.cwd,
		opt.stdin,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream(ctx, handle, codex.NewACPDecoder()), nil
}
