package allinone

import (
	"context"

	"github.com/llm-infra/secvirt/sdk-go/codeide"
	"github.com/llm-infra/secvirt/sdk-go/desktop"
	"github.com/llm-infra/secvirt/sdk-go/desktop/claude"
	"github.com/llm-infra/secvirt/sdk-go/desktop/codex"
	"github.com/llm-infra/secvirt/sdk-go/hostmcp"
	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Sandbox struct {
	*sandbox.Sandbox
	codeide *codeide.Sandbox
	hostmcp *hostmcp.Sandbox
	desktop *desktop.Sandbox
}

func NewSandbox(ctx context.Context, opts ...sandbox.Option) (*Sandbox, error) {
	client, err := sandbox.NewSandbox(ctx,
		append(opts,
			sandbox.WithTemplate(sandbox.TemplateAllInOne),
			sandbox.WithHealthPorts([]int{
				codeide.DefaultCodeIDEPort,
				hostmcp.DefaultMcpServerPort,
				hostmcp.DefaultMcpRouterPort,
			}),
		)...)
	if err != nil {
		return nil, err
	}
	return &Sandbox{
		Sandbox: client,
		codeide: &codeide.Sandbox{Sandbox: client},
		hostmcp: &hostmcp.Sandbox{Sandbox: client},
		desktop: &desktop.Sandbox{Sandbox: client},
	}, nil
}

// codeide
func (s *Sandbox) Packages(ctx context.Context,
	lang string) ([]codeide.PackagesResponse, error) {
	return s.codeide.Packages(ctx, lang)
}

func (s *Sandbox) RunCodeV1(ctx context.Context, lang, code string,
	inputs map[string]any) (*codeide.RunCodeResponseV1, error) {
	return s.codeide.RunCodeV1(ctx, lang, code, inputs)
}

func (s *Sandbox) RunCode(ctx context.Context, lang, code string,
	inputs map[string]any) (*codeide.RunCodeResponse, error) {
	return s.codeide.RunCode(ctx, lang, code, inputs)
}

// hostmcp
func (s *Sandbox) GetLaunchMCPs(ctx context.Context) ([]hostmcp.MCPEndpoint, error) {
	return s.hostmcp.GetLaunchMCPs(ctx)
}

func (s *Sandbox) Launch(ctx context.Context, preloads []hostmcp.Preload,
	config *hostmcp.ServersFile, reload bool) ([]hostmcp.MCPEndpoint, error) {
	return s.hostmcp.Launch(ctx, preloads, config, reload)
}

func (s *Sandbox) Connect(ctx context.Context, endpoint hostmcp.MCPEndpoint,
) (*mcp.ClientSession, error) {
	return s.hostmcp.Connect(ctx, endpoint)
}

// desktop
func (s *Sandbox) SetCodexConfig(ctx context.Context, config *codex.Config,
	opts ...desktop.Option) error {
	return s.desktop.SetCodexConfig(ctx, config, opts...)
}

func (s *Sandbox) CodexChat(ctx context.Context, content string,
	opts ...desktop.Option) (*commands.Stream[desktop.CodexEvent], error) {
	return s.desktop.CodexChat(ctx, content, opts...)
}

func (s *Sandbox) SetClaudeSettings(ctx context.Context, settings *claude.Settings,
	opts ...desktop.Option) error {
	return s.desktop.SetClaudeSettings(ctx, settings, opts...)
}

func (s *Sandbox) SetSkills(ctx context.Context, skills []desktop.Skill,
	opts ...desktop.Option) error {
	return s.desktop.SetSkills(ctx, skills, opts...)
}

func (s *Sandbox) ClaudeChat(ctx context.Context, content string,
	opts ...desktop.Option) (*commands.Stream[[]claude.Message], error) {
	return s.desktop.ClaudeChat(ctx, content, opts...)
}
