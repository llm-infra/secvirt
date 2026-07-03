package desktop

import (
	"context"
	"fmt"
	"io"
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

func (s *Sandbox) SetCodexSkills(ctx context.Context, skills map[string]io.Reader,
	opts ...Option) error {
	opt := NewOptions(s.HomeDir())
	for _, o := range opts {
		o(opt)
	}

	skillPath := filepath.Join(opt.cwd, ".codex", "skills")
	for name, skill := range skills {
		data, err := io.ReadAll(skill)
		if err != nil {
			return err
		}

		temp := filepath.Join(skillPath, name)
		if exist, err := s.Filesystem().Exist(ctx, temp); err == nil && exist {
			s.Filesystem().Remove(ctx, temp)
		}
		if err = s.Filesystem().Write(ctx, temp, data); err != nil {
			return err
		}

		skillName, has, err := checkZipRootDir(ctx, name, data)
		if err != nil {
			return err
		}
		newDir := filepath.Join(skillPath, skillName)
		if exist, err := s.Filesystem().Exist(ctx, newDir); err == nil && exist {
			s.Filesystem().Remove(ctx, newDir)
		}

		if has {
			_, err = s.Cmd().Run(ctx,
				fmt.Sprintf("unzip -o %s && rm -rf %s", temp, temp),
				nil,
				skillPath,
				false,
			)
			if err != nil {
				return err
			}
		} else {
			if _, err = s.Filesystem().Mkdir(ctx, newDir); err != nil {
				return err
			}
			_, err = s.Cmd().Run(ctx,
				fmt.Sprintf("unzip -o %s -d %s && rm -rf %s",
					temp, newDir, temp),
				nil,
				skillPath,
				false,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
