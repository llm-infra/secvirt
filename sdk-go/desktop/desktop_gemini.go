package desktop

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/llm-infra/acp/sdk/go/acp"
	"github.com/llm-infra/secvirt/sdk-go/desktop/gemini"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
)

func (s *Sandbox) SetGeminiConfig(ctx context.Context, config gemini.Config,
	opts ...Option) error {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	data, err := godotenv.Marshal(config)
	if err != nil {
		return err
	}

	return s.Filesystem().Write(ctx,
		filepath.Join(opt.cwd, ".env"),
		[]byte(data),
	)
}

func (s *Sandbox) GeminiChat(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[gemini.Message], error) {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("gemini -p %s --output-format stream-json --yolo",
			strconv.Quote(content)),
		opt.envs,
		opt.cwd,
		opt.stdin,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream(ctx, handle, gemini.NewDecoder()), nil
}

func (s *Sandbox) GeminiChatWithACPStream(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[[]acp.Event], error) {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("gemini -p %s --output-format stream-json --yolo",
			strconv.Quote(content)),
		opt.envs,
		opt.cwd,
		opt.stdin,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream(ctx, handle, gemini.NewACPDecoder()), nil
}
