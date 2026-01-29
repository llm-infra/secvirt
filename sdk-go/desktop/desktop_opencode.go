package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"

	"github.com/llm-infra/secvirt/sdk-go/desktop/opencode"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
)

// model provider mcp
func (s *Sandbox) SetOpenCodeConfig(ctx context.Context, config *opencode.Config,
	opts ...Option) error {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return s.Filesystem().Write(ctx,
		filepath.Join(opt.cwd, ".opencode", "opencode.json"),
		data,
	)
}

func (s *Sandbox) SetOpenCodeSkills(ctx context.Context, skills map[string]io.Reader,
	opts ...Option) error {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	skillPath := filepath.Join(opt.cwd, ".opencode", "skills")
	for name, skill := range skills {
		data, err := io.ReadAll(skill)
		if err != nil {
			return err
		}

		temp := filepath.Join(skillPath, name)
		if err = s.Filesystem().Write(ctx, temp, data); err != nil {
			return err
		}

		_, err = s.Cmd().Run(ctx,
			fmt.Sprintf("tar -zxvf %s && rm -rf %s", temp, temp),
			nil,
			skillPath,
			false,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Sandbox) SetOpenCodeAgents(ctx context.Context, agents map[string]io.Reader,
	opts ...Option) error {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	for name, agent := range agents {
		path := filepath.Join(opt.cwd, ".opencode", name)

		data, err := io.ReadAll(agent)
		if err != nil {
			return err
		}

		if err = s.Filesystem().Write(ctx, path, data); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sandbox) OpenCodeChat(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[opencode.Message], error) {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("opencode run %s --format json",
			strconv.Quote(content)),
		opt.envs,
		opt.cwd,
		opt.stdin,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream(ctx, handle, opencode.NewDecoder()), nil
}
