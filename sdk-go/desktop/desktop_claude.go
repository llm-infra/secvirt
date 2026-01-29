package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/llm-infra/secvirt/sdk-go/desktop/claude"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
)

func (s *Sandbox) SetClaudeSettings(ctx context.Context, settings *claude.Settings,
	opts ...Option) error {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	settings.SetEnv("ANTHROPIC_BASE_URL", "http://192.168.134.142:8995")
	settings.SetEnv("ANTHROPIC_AUTH_TOKEN", "skip")
	settings.SetEnv("ANTHROPIC_CUSTOM_HEADERS", "EXT: x-uid:17;x-org:1802543778616180738;")
	settings.SetEnv("ANTHROPIC_MODEL", "glm-4.6")
	settings.SetEnv("CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC", "1")

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return s.Filesystem().Write(ctx,
		filepath.Join(opt.cwd, ".claude", "settings.json"),
		data,
	)
}

type Skill struct {
	Name  string      `json:"name"`
	Files []SkillFile `json:"files"`
}

type SkillFile struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

func (s *Sandbox) SetClaudeSkills(ctx context.Context, skills []Skill,
	opts ...Option) error {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	for _, skill := range skills {
		skillRoot := filepath.Join(opt.cwd, ".claude", ".skills", skill.Name)

		for _, f := range skill.Files {
			if err := s.Filesystem().Write(ctx, filepath.Join(skillRoot, f.Path), f.Content); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Sandbox) ClaudeChat(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[[]claude.Message], error) {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("claude -p %s --output-format stream-json --verbose",
			strconv.Quote(content)),
		opt.envs,
		opt.cwd,
		opt.stdin,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream(ctx, handle, claude.NewDecoder()), nil
}
