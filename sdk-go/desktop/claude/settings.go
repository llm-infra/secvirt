package claude

// all supported environment variables
const (
	// API Configuration
	AnthropicAPIKey        = "ANTHROPIC_API_KEY"
	AnthropicAuthToken     = "ANTHROPIC_AUTH_TOKEN"
	AnthropicCustomHeaders = "ANTHROPIC_CUSTOM_HEADERS"

	// Model Configuration
	AnthropicDefaultHaikuModel       = "ANTHROPIC_DEFAULT_HAIKU_MODEL"
	AnthropicDefaultOpusModel        = "ANTHROPIC_DEFAULT_OPUS_MODEL"
	AnthropicDefaultSonnetModel      = "ANTHROPIC_DEFAULT_SONNET_MODEL"
	AnthropicModel                   = "ANTHROPIC_MODEL"
	AnthropicSmallFastModel          = "ANTHROPIC_SMALL_FAST_MODEL"
	AnthropicSmallFastModelAWSRegion = "ANTHROPIC_SMALL_FAST_MODEL_AWS_REGION"

	// Foundry
	AnthropicFoundryAPIKey = "ANTHROPIC_FOUNDRY_API_KEY"

	// AWS/Bedrock
	AWSBearerTokenBedrock = "AWS_BEARER_TOKEN_BEDROCK"

	// Bash Configuration
	BashDefaultTimeoutMS                = "BASH_DEFAULT_TIMEOUT_MS"
	BashMaxOutputLength                 = "BASH_MAX_OUTPUT_LENGTH"
	BashMaxTimeoutMS                    = "BASH_MAX_TIMEOUT_MS"
	ClaudeBashMaintainProjectWorkingDir = "CLAUDE_BASH_MAINTAIN_PROJECT_WORKING_DIR"

	// Claude Code Configuration
	ClaudeCodeAPIKeyHelperTTLMS          = "CLAUDE_CODE_API_KEY_HELPER_TTL_MS"
	ClaudeCodeClientCert                 = "CLAUDE_CODE_CLIENT_CERT"
	ClaudeCodeClientKey                  = "CLAUDE_CODE_CLIENT_KEY"
	ClaudeCodeClientKeyPassphrase        = "CLAUDE_CODE_CLIENT_KEY_PASSPHRASE"
	ClaudeCodeDisableExperimentalBetas   = "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"
	ClaudeCodeDisableNonessentialTraffic = "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"
	ClaudeCodeDisableTerminalTitle       = "CLAUDE_CODE_DISABLE_TERMINAL_TITLE"
	ClaudeCodeIDESkipAutoInstall         = "CLAUDE_CODE_IDE_SKIP_AUTO_INSTALL"
	ClaudeCodeMaxOutputTokens            = "CLAUDE_CODE_MAX_OUTPUT_TOKENS"
	ClaudeCodeShellPrefix                = "CLAUDE_CODE_SHELL_PREFIX"
	ClaudeCodeSkipBedrockAuth            = "CLAUDE_CODE_SKIP_BEDROCK_AUTH"
	ClaudeCodeSkipFoundryAuth            = "CLAUDE_CODE_SKIP_FOUNDRY_AUTH"
	ClaudeCodeSkipVertexAuth             = "CLAUDE_CODE_SKIP_VERTEX_AUTH"
	ClaudeCodeSubagentModel              = "CLAUDE_CODE_SUBAGENT_MODEL"
	ClaudeCodeUseBedrock                 = "CLAUDE_CODE_USE_BEDROCK"
	ClaudeCodeUseFoundry                 = "CLAUDE_CODE_USE_FOUNDRY"
	ClaudeCodeUseVertex                  = "CLAUDE_CODE_USE_VERTEX"
	ClaudeConfigDir                      = "CLAUDE_CONFIG_DIR"

	// Feature Toggles
	DisableAutoupdater            = "DISABLE_AUTOUPDATER"
	DisableBugCommand             = "DISABLE_BUG_COMMAND"
	DisableCostWarnings           = "DISABLE_COST_WARNINGS"
	DisableErrorReporting         = "DISABLE_ERROR_REPORTING"
	DisableNonEssentialModelCalls = "DISABLE_NON_ESSENTIAL_MODEL_CALLS"
	DisablePromptCaching          = "DISABLE_PROMPT_CACHING"
	DisablePromptCachingHaiku     = "DISABLE_PROMPT_CACHING_HAIKU"
	DisablePromptCachingOpus      = "DISABLE_PROMPT_CACHING_OPUS"
	DisablePromptCachingSonnet    = "DISABLE_PROMPT_CACHING_SONNET"
	DisableTelemetry              = "DISABLE_TELEMETRY"

	// Proxy Configuration
	HTTPProxy  = "HTTP_PROXY"
	HTTPSProxy = "HTTPS_PROXY"
	NoProxy    = "NO_PROXY"

	// MCP Configuration
	MaxMCPOutputTokens = "MAX_MCP_OUTPUT_TOKENS"
	MCPTimeout         = "MCP_TIMEOUT"
	MCPToolTimeout     = "MCP_TOOL_TIMEOUT"

	// Thinking Configuration
	MaxThinkingTokens = "MAX_THINKING_TOKENS"

	// Slash Commands
	SlashCommandToolCharBudget = "SLASH_COMMAND_TOOL_CHAR_BUDGET"

	// Tools
	UseBuiltinRipgrep = "USE_BUILTIN_RIPGREP"

	// Vertex AI Regions
	VertexRegionClaude35Haiku  = "VERTEX_REGION_CLAUDE_3_5_HAIKU"
	VertexRegionClaude37Sonnet = "VERTEX_REGION_CLAUDE_3_7_SONNET"
	VertexRegionClaude40Opus   = "VERTEX_REGION_CLAUDE_4_0_OPUS"
	VertexRegionClaude40Sonnet = "VERTEX_REGION_CLAUDE_4_0_SONNET"
	VertexRegionClaude41Opus   = "VERTEX_REGION_CLAUDE_4_1_OPUS"
)

// Settings represents the complete Claude Code settings.json structure
type Settings struct {
	// API and Authentication
	APIKeyHelper *string `json:"apiKeyHelper,omitempty"`

	// Session Management
	CleanupPeriodDays *int `json:"cleanupPeriodDays,omitempty"`

	// Company Settings
	CompanyAnnouncements []string `json:"companyAnnouncements,omitempty"`

	// Environment Variables
	Env map[string]string `json:"env,omitempty"`

	// Git Settings
	IncludeCoAuthoredBy *bool `json:"includeCoAuthoredBy,omitempty"`

	// Permissions
	Permissions *Permissions `json:"permissions,omitempty"`

	// Hooks
	Hooks           map[string]interface{} `json:"hooks,omitempty"`
	DisableAllHooks *bool                  `json:"disableAllHooks,omitempty"`

	// Model Configuration
	Model       *string `json:"model,omitempty"`
	OutputStyle *string `json:"outputStyle,omitempty"`

	// Login Configuration
	ForceLoginMethod  *string `json:"forceLoginMethod,omitempty"`
	ForceLoginOrgUUID *string `json:"forceLoginOrgUUID,omitempty"`

	// MCP Configuration
	EnableAllProjectMcpServers *bool           `json:"enableAllProjectMcpServers,omitempty"`
	EnabledMcpjsonServers      []string        `json:"enabledMcpjsonServers,omitempty"`
	DisabledMcpjsonServers     []string        `json:"disabledMcpjsonServers,omitempty"`
	AllowedMcpServers          []MCPServerRule `json:"allowedMcpServers,omitempty"`
	DeniedMcpServers           []MCPServerRule `json:"deniedMcpServers,omitempty"`

	// AWS Configuration
	AWSAuthRefresh      *string `json:"awsAuthRefresh,omitempty"`
	AWSCredentialExport *string `json:"awsCredentialExport,omitempty"`

	// Sandbox Configuration
	Sandbox *SandboxConfig `json:"sandbox,omitempty"`

	// Status Line
	StatusLine *StatusLineConfig `json:"statusLine,omitempty"`

	// Plugin Configuration
	EnabledPlugins         map[string]bool              `json:"enabledPlugins,omitempty"`
	ExtraKnownMarketplaces map[string]MarketplaceConfig `json:"extraKnownMarketplaces,omitempty"`
}

// Permissions defines permission rules for tools
type Permissions struct {
	Allow                        []string `json:"allow,omitempty"`
	Ask                          []string `json:"ask,omitempty"`
	Deny                         []string `json:"deny,omitempty"`
	AdditionalDirectories        []string `json:"additionalDirectories,omitempty"`
	DefaultMode                  *string  `json:"defaultMode,omitempty"`
	DisableBypassPermissionsMode *string  `json:"disableBypassPermissionsMode,omitempty"`
}

// SandboxConfig configures bash sandboxing behavior
type SandboxConfig struct {
	Enabled                   *bool          `json:"enabled,omitempty"`
	AutoAllowBashIfSandboxed  *bool          `json:"autoAllowBashIfSandboxed,omitempty"`
	ExcludedCommands          []string       `json:"excludedCommands,omitempty"`
	AllowUnsandboxedCommands  *bool          `json:"allowUnsandboxedCommands,omitempty"`
	Network                   *NetworkConfig `json:"network,omitempty"`
	EnableWeakerNestedSandbox *bool          `json:"enableWeakerNestedSandbox,omitempty"`
}

// NetworkConfig configures sandbox network settings
type NetworkConfig struct {
	AllowUnixSockets  []string `json:"allowUnixSockets,omitempty"`
	AllowLocalBinding *bool    `json:"allowLocalBinding,omitempty"`
	HTTPProxyPort     *int     `json:"httpProxyPort,omitempty"`
	SocksProxyPort    *int     `json:"socksProxyPort,omitempty"`
}

// MCPServerRule defines MCP server allow/deny rules
type MCPServerRule struct {
	ServerName string `json:"serverName"`
}

// MarketplaceConfig defines a plugin marketplace
type MarketplaceConfig struct {
	Source MarketplaceSource `json:"source"`
}

// MarketplaceSource defines the source of a marketplace
type MarketplaceSource struct {
	Source string  `json:"source"` // "github", "git", or "directory"
	Repo   *string `json:"repo,omitempty"`
	URL    *string `json:"url,omitempty"`
	Path   *string `json:"path,omitempty"`
}

// StatusLineConfig configures the status line display
type StatusLineConfig struct {
	Type    string  `json:"type"` // "command" or other types
	Command *string `json:"command,omitempty"`
}

// NewSettings creates a new Settings with default values
func NewSettings() *Settings {
	defaultTrue := true
	defaultCleanup := 30

	return &Settings{
		CleanupPeriodDays:   &defaultCleanup,
		IncludeCoAuthoredBy: &defaultTrue,
		Env:                 make(map[string]string),
		Permissions:         &Permissions{},
	}
}

// SetEnv adds an environment variable to settings
func (s *Settings) SetEnv(key, value string) {
	if s.Env == nil {
		s.Env = make(map[string]string)
	}
	s.Env[key] = value
}

// AddPermission adds a permission rule
func (p *Permissions) AddPermission(ruleType string, rule string) {
	switch ruleType {
	case "allow":
		p.Allow = append(p.Allow, rule)
	case "ask":
		p.Ask = append(p.Ask, rule)
	case "deny":
		p.Deny = append(p.Deny, rule)
	}
}
