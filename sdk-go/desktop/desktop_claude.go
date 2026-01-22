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

func (s *Sandbox) SetSkills(ctx context.Context, skills []Skill,
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

	return commands.NewStream(ctx, handle, newClaudeDecoder()), nil
}

type ClaudeDecoder struct {
	parser *claude.Parser
}

func newClaudeDecoder() *ClaudeDecoder {
	return &ClaudeDecoder{
		parser: claude.NewParser(),
	}
}

func (d *ClaudeDecoder) Decode(data []byte) ([]claude.Message, error) {
	return d.parser.ProcessLine(string(data))
}
