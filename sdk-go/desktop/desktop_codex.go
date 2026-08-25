package desktop

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/BurntSushi/toml"
	"github.com/llm-infra/acp/sdk/go/acp"
	"github.com/llm-infra/secvirt/sdk-go/desktop/codex"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
	"github.com/sirupsen/logrus"
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

type SkillPackage struct {
	Name   string
	Hash   string
	Reader io.Reader
}

func (s *Sandbox) SetCodexSkills(ctx context.Context, skills []SkillPackage,
	opts ...Option) error {
	opt := NewOptions(s.HomeDir())
	for _, o := range opts {
		o(opt)
	}

	skillPath := filepath.Join(opt.cwd, ".codex", "skills")
	if _, err := s.Filesystem().Mkdir(ctx, skillPath); err != nil {
		return err
	}

	existingSkills, err := s.Filesystem().List(ctx, skillPath, 1)
	if err != nil {
		return err
	}

	desiredSkills := make(map[string]struct{}, len(skills)*2)
	for _, skill := range skills {
		desiredSkills[skill.Name] = struct{}{}
		desiredSkills[archiveBaseName(skill.Name)] = struct{}{}
	}

	for _, existingSkill := range existingSkills {
		if _, ok := desiredSkills[existingSkill.Name]; ok {
			continue
		}
		if opt.shouldLog {
			logrus.WithContext(ctx).Infof("skill should removed: %s", existingSkill.Name)
		}
		if err := s.Filesystem().Remove(ctx, existingSkill.Path); err != nil {
			return err
		}
	}

	for _, skill := range skills {
		temp := filepath.Join(skillPath, skill.Name)
		exist, err := s.Filesystem().Exist(ctx, temp)
		if err != nil && connect.CodeOf(err) != connect.CodeNotFound {
			return err
		}
		if exist {
			installedPackage, err := s.Filesystem().Read(ctx, temp)
			if err != nil {
				return err
			}
			installedHash := sha256.Sum256(installedPackage)
			if strings.EqualFold(hex.EncodeToString(installedHash[:]), skill.Hash) {
				if opt.shouldLog {
					logrus.WithContext(ctx).Infof("skill has installed, skip: %s", skill.Name)
				}
				continue
			}
		}

		data, err := io.ReadAll(skill.Reader)
		if err != nil {
			return err
		}

		skillName, has, err := checkZipRootDir(ctx, skill.Name, data)
		if err != nil {
			return err
		}
		newDir := filepath.Join(skillPath, skillName)

		if exist {
			if err = s.Filesystem().Remove(ctx, temp); err != nil {
				return err
			}
		}
		if exist, err := s.Filesystem().Exist(ctx, newDir); err == nil && exist {
			if err = s.Filesystem().Remove(ctx, newDir); err != nil {
				return err
			}
		}
		if err = s.Filesystem().Write(ctx, temp, data); err != nil {
			return err
		}

		if has {
			_, err = s.Cmd().Run(ctx,
				fmt.Sprintf("unzip -o %s", temp),
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
				fmt.Sprintf("unzip -o %s -d %s", temp, newDir),
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

func (s *Sandbox) SetCodexHooks(ctx context.Context, hooks io.Reader,
	opts ...Option) error {
	opt := NewOptions(s.HomeDir())
	for _, o := range opts {
		o(opt)
	}

	hooksPath := filepath.Join(opt.cwd, ".codex", "hooks")
	temp := filepath.Join(hooksPath, "hooks.zip")
	exist, err := s.Filesystem().Exist(ctx, temp)
	if err != nil && connect.CodeOf(err) != connect.CodeNotFound {
		return err
	}
	if exist {
		if opt.shouldLog {
			logrus.WithContext(ctx).Infof("hooks has installed, skip")
		}

		return nil
	}

	data, err := io.ReadAll(hooks)
	if err != nil {
		return err
	}
	if err = s.Filesystem().Write(ctx, temp, data); err != nil {
		return err
	}

	_, err = s.Cmd().Run(ctx,
		fmt.Sprintf("unzip -o %s", temp),
		nil,
		hooksPath,
		false,
	)
	if err != nil {
		return err
	}

	_, err = s.Cmd().Run(ctx,
		fmt.Sprintf("mv %s %s",
			filepath.Join(hooksPath, "hooks.json"),
			filepath.Join(opt.cwd, ".codex", "hooks.json"),
		),
		nil,
		opt.cwd,
		false,
	)
	return err
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
