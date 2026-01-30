package opencode

type Config struct {
	Schema   string              `json:"$schema"`
	Model    string              `json:"model"`
	Provider map[string]Provider `json:"provider"`
	Mcp      map[string]Mcp      `json:"mcp"`
}

type Provider struct {
	Npm     string                   `json:"npm"`
	Name    string                   `json:"name"`
	Options ProviderOptions          `json:"options"`
	Models  map[string]ProviderModel `json:"models"`
}

func NewOpenAIProvider(name, baseUrl string,
	headers map[string]string, models []string) Provider {
	provider := Provider{
		Npm:  "@ai-sdk/openai-compatible",
		Name: name,
		Options: ProviderOptions{
			APIKey:  "",
			BaseURL: baseUrl,
			Headers: headers,
		},
		Models: map[string]ProviderModel{},
	}
	for _, model := range models {
		provider.Models[model] = ProviderModel{
			Name: model,
		}
	}
	return provider
}

type ProviderModel struct {
	Name string `json:"name"`
}

type ProviderOptions struct {
	APIKey  string            `json:"apiKey"`
	BaseURL string            `json:"baseURL"`
	Headers map[string]string `json:"headers,omitempty"`
}

const (
	McpTypeLocal  = "local"
	McpTypeRemote = "remote"
)

type Mcp struct {
	Type        string            `json:"type"`
	Command     []string          `json:"command,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	URL         string            `json:"url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Enabled     bool              `json:"enabled"`
}

func NewConfig(model string, opts ...Option) *Config {
	cfg := &Config{
		Schema: "https://opencode.ai/config.json",
		Model:  model,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

type Option func(*Config)

func WithProvider(provider Provider) Option {
	return func(c *Config) {
		if c.Provider == nil {
			c.Provider = map[string]Provider{}
		}

		c.Provider[provider.Name] = provider
	}
}

func WithMcp(name string, mcp Mcp) Option {
	return func(c *Config) {
		if c.Mcp == nil {
			c.Mcp = map[string]Mcp{}
		}

		c.Mcp[name] = mcp
	}
}
