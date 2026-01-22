package codex

const (
	APIKeyEnvVar = "SECWALL_API_KEY"
)

type Config struct {
	// Core Model Selection
	Model                      string `toml:"model,omitempty"`
	ReviewModel                string `toml:"review_model,omitempty"`
	ModelProvider              string `toml:"model_provider,omitempty"`
	ModelContextWindow         *int   `toml:"model_context_window,omitempty"`
	ModelAutoCompactTokenLimit *int   `toml:"model_auto_compact_token_limit,omitempty"`
	ToolOutputTokenLimit       *int   `toml:"tool_output_token_limit,omitempty"`

	// Reasoning & Verbosity
	ModelReasoningEffort            string `toml:"model_reasoning_effort,omitempty"`
	ModelReasoningSummary           string `toml:"model_reasoning_summary,omitempty"`
	ModelVerbosity                  string `toml:"model_verbosity,omitempty"`
	ModelSupportsReasoningSummaries bool   `toml:"model_supports_reasoning_summaries,omitempty"`
	ModelReasoningSummaryFormat     string `toml:"model_reasoning_summary_format,omitempty"`

	// Instruction Overrides
	DeveloperInstructions         string `toml:"developer_instructions,omitempty"`
	Instructions                  string `toml:"instructions,omitempty"`
	CompactPrompt                 string `toml:"compact_prompt,omitempty"`
	ExperimentalInstructionsFile  string `toml:"experimental_instructions_file,omitempty"`
	ExperimentalCompactPromptFile string `toml:"experimental_compact_prompt_file,omitempty"`

	// Notifications
	Notify []string `toml:"notify,omitempty"`

	// Approval & Sandbox
	ApprovalPolicy        string                 `toml:"approval_policy,omitempty"`
	SandboxMode           string                 `toml:"sandbox_mode,omitempty"`
	SandboxWorkspaceWrite *SandboxWorkspaceWrite `toml:"sandbox_workspace_write,omitempty"`

	// Shell Environment
	ShellEnvironmentPolicy *ShellEnvironmentPolicy `toml:"shell_environment_policy,omitempty"`

	// History & File Opener
	History    *History `toml:"history,omitempty"`
	FileOpener string   `toml:"file_opener,omitempty"`

	// UI & Misc
	TUI                         *TUIConfig `toml:"tui,omitempty"`
	HideAgentReasoning          bool       `toml:"hide_agent_reasoning,omitempty"`
	ShowRawAgentReasoning       bool       `toml:"show_raw_agent_reasoning,omitempty"`
	DisablePasteBurst           bool       `toml:"disable_paste_burst,omitempty"`
	WindowsWSLSetupAcknowledged bool       `toml:"windows_wsl_setup_acknowledged,omitempty"`
	Notice                      *Notice    `toml:"notice,omitempty"`

	// Authentication
	CLIAuthCredentialsStore  string `toml:"cli_auth_credentials_store,omitempty"`
	ChatGPTBaseURL           string `toml:"chatgpt_base_url,omitempty"`
	ForcedChatGPTWorkspaceID string `toml:"forced_chatgpt_workspace_id,omitempty"`
	ForcedLoginMethod        string `toml:"forced_login_method,omitempty"`
	MCPOAuthCredentialsStore string `toml:"mcp_oauth_credentials_store,omitempty"`

	// Project Documentation
	ProjectDocMaxBytes          *int     `toml:"project_doc_max_bytes,omitempty"`
	ProjectDocFallbackFilenames []string `toml:"project_doc_fallback_filenames,omitempty"`

	// Tools (legacy)
	Tools *Tools `toml:"tools,omitempty"`

	// Features
	Features *Features `toml:"features,omitempty"`

	// Experimental
	ExperimentalUseFreeformApplyPatch bool `toml:"experimental_use_freeform_apply_patch,omitempty"`

	// MCP Servers
	MCPServers map[string]MCPServer `toml:"mcp_servers,omitempty"`

	// Model Providers
	ModelProviders map[string]ModelProvider `toml:"model_providers,omitempty"`

	// Profiles
	Profile  string                 `toml:"profile,omitempty"`
	Profiles map[string]interface{} `toml:"profiles,omitempty"`

	// Projects (trust levels)
	Projects map[string]Project `toml:"projects,omitempty"`

	// OpenTelemetry
	OTEL *OTEL `toml:"otel,omitempty"`
}

// SandboxWorkspaceWrite 沙箱工作区写配置
type SandboxWorkspaceWrite struct {
	WritableRoots       []string `toml:"writable_roots"`
	NetworkAccess       bool     `toml:"network_access"`
	ExcludeTMPDirEnvVar bool     `toml:"exclude_tmpdir_env_var"`
	ExcludeSlashTmp     bool     `toml:"exclude_slash_tmp"`
}

// ShellEnvironmentPolicy Shell 环境策略
type ShellEnvironmentPolicy struct {
	Inherit                string            `toml:"inherit"`
	IgnoreDefaultExcludes  bool              `toml:"ignore_default_excludes"`
	Exclude                []string          `toml:"exclude"`
	Set                    map[string]string `toml:"set"`
	IncludeOnly            []string          `toml:"include_only"`
	ExperimentalUseProfile bool              `toml:"experimental_use_profile"`
}

// History 历史记录配置
type History struct {
	Persistence string `toml:"persistence"`
	MaxBytes    *int   `toml:"max_bytes,omitempty"`
}

// TUIConfig TUI 配置
type TUIConfig struct {
	Notifications interface{} `toml:"notifications"` // bool or []string
	Animations    bool        `toml:"animations"`
}

// Notice 通知配置
type Notice struct {
	HideFullAccessWarning   *bool `toml:"hide_full_access_warning,omitempty"`
	HideRateLimitModelNudge *bool `toml:"hide_rate_limit_model_nudge,omitempty"`
}

// Tools 工具配置
type Tools struct {
	WebSearch bool `toml:"web_search"`
	ViewImage bool `toml:"view_image"`
}

// Features 特性标志
type Features struct {
	ApplyPatchFreeform         *bool `toml:"apply_patch_freeform,omitempty"`
	ElevatedWindowsSandbox     *bool `toml:"elevated_windows_sandbox,omitempty"`
	ExecPolicy                 *bool `toml:"exec_policy,omitempty"`
	ExperimentalWindowsSandbox *bool `toml:"experimental_windows_sandbox,omitempty"`
	Parallel                   *bool `toml:"parallel,omitempty"`
	RemoteCompaction           *bool `toml:"remote_compaction,omitempty"`
	RemoteModels               *bool `toml:"remote_models,omitempty"`
	ShellSnapshot              *bool `toml:"shell_snapshot,omitempty"`
	ShellTool                  *bool `toml:"shell_tool,omitempty"`
	Skills                     *bool `toml:"skills,omitempty"`
	UnifiedExec                *bool `toml:"unified_exec,omitempty"`
	ViewImageTool              *bool `toml:"view_image_tool,omitempty"`
	Warnings                   *bool `toml:"warnings,omitempty"`
	WebSearchRequest           *bool `toml:"web_search_request,omitempty"`
	Undo                       *bool `toml:"undo,omitempty"`
}

// MCPServer MCP 服务器配置
type MCPServer struct {
	// STDIO transport
	Command *string           `toml:"command,omitempty"`
	Args    []string          `toml:"args,omitempty"`
	Env     map[string]string `toml:"env,omitempty"`
	EnvVars []string          `toml:"env_vars,omitempty"`
	CWD     *string           `toml:"cwd,omitempty"`

	// HTTP transport
	URL               *string           `toml:"url,omitempty"`
	BearerTokenEnvVar *string           `toml:"bearer_token_env_var,omitempty"`
	HTTPHeaders       map[string]string `toml:"http_headers,omitempty"`
	EnvHTTPHeaders    map[string]string `toml:"env_http_headers,omitempty"`

	// Common
	StartupTimeoutSec *float64 `toml:"startup_timeout_sec,omitempty"`
	StartupTimeoutMS  *int     `toml:"startup_timeout_ms,omitempty"`
	ToolTimeoutSec    *float64 `toml:"tool_timeout_sec,omitempty"`
	EnabledTools      []string `toml:"enabled_tools,omitempty"`
	DisabledTools     []string `toml:"disabled_tools,omitempty"`
	Enabled           *bool    `toml:"enabled,omitempty"`
}

// ModelProvider 模型提供商配置
type ModelProvider struct {
	Name                    string            `toml:"name"`
	BaseURL                 string            `toml:"base_url"`
	WireAPI                 string            `toml:"wire_api,omitempty"`
	EnvKey                  string            `toml:"env_key,omitempty"`
	EnvKeyInstructions      string            `toml:"env_key_instructions,omitempty"`
	QueryParams             map[string]string `toml:"query_params,omitempty"`
	HTTPHeaders             map[string]string `toml:"http_headers,omitempty"`
	EnvHTTPHeaders          map[string]string `toml:"env_http_headers,omitempty"`
	RequiresOpenAIAuth      *bool             `toml:"requires_openai_auth,omitempty"`
	RequestMaxRetries       *int              `toml:"request_max_retries,omitempty"`
	StreamMaxRetries        *int              `toml:"stream_max_retries,omitempty"`
	StreamIdleTimeoutMS     *int              `toml:"stream_idle_timeout_ms,omitempty"`
	ExperimentalBearerToken *string           `toml:"experimental_bearer_token,omitempty"`
}

// Project 项目配置
type Project struct {
	TrustLevel string `toml:"trust_level"`
}

// OTEL OpenTelemetry 配置
type OTEL struct {
	LogUserPrompt bool        `toml:"log_user_prompt"`
	Environment   string      `toml:"environment"`
	Exporter      interface{} `toml:"exporter"` // "none" or OTELExporter
}

// OTELExporter OTLP 导出器配置
type OTELExporter struct {
	OTLPHTTP *OTLPHTTPExporter `toml:"otlp-http,omitempty"`
	OTLPGRPC *OTLPGRPCExporter `toml:"otlp-grpc,omitempty"`
}

// OTLPHTTPExporter OTLP HTTP 导出器
type OTLPHTTPExporter struct {
	Endpoint string            `toml:"endpoint"`
	Protocol string            `toml:"protocol"` // "binary" or "json"
	Headers  map[string]string `toml:"headers,omitempty"`
	TLS      *TLSConfig        `toml:"tls,omitempty"`
}

// OTLPGRPCExporter OTLP gRPC 导出器
type OTLPGRPCExporter struct {
	Endpoint string            `toml:"endpoint"`
	Headers  map[string]string `toml:"headers,omitempty"`
	TLS      *TLSConfig        `toml:"tls,omitempty"`
}

// TLSConfig TLS 配置
type TLSConfig struct {
	CACertificate     string `toml:"ca-certificate"`
	ClientCertificate string `toml:"client-certificate"`
	ClientPrivateKey  string `toml:"client-private-key"`
}

// NewConfig 创建默认配置
func NewConfig(model string, provider *ModelProvider) *Config {
	return &Config{
		Model:         model,
		ModelProvider: provider.Name,

		ModelProviders: map[string]ModelProvider{
			provider.Name: *provider,
		},
	}
}
