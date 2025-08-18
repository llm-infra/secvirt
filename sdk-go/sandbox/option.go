package sandbox

import "fmt"

type Option func(*Options)

type Options struct {
	host        string
	user        string
	template    TemplateType
	sandboxID   string
	healthPorts []int
}

func newOptions() *Options {
	return &Options{
		host:     "localhost",
		user:     "default",
		template: "codeide:latest",
	}
}

func WithHost(host string) Option {
	return func(o *Options) { o.host = host }
}

func WithUser(id string) Option {
	return func(o *Options) { o.user = id }
}

func WithTemplate(tmpl TemplateType) Option {
	return func(o *Options) { o.template = tmpl }
}

func WithSandboxID(id string) Option {
	return func(o *Options) { o.sandboxID = id }
}

func WithHealthPorts(ports []int) Option {
	return func(o *Options) { o.healthPorts = ports }
}

type TemplateType string

const (
	TemplateCodeIDE TemplateType = "codeide:latest"
	TemplateHostMCP TemplateType = "hostmcp:latest"
	TemplateDesktop TemplateType = "desktop:latest"
)

type SandboxDetail struct {
	ID             string   `json:"id"`
	IP             string   `json:"ip"`
	User           string   `json:"user"`
	CreateAt       string   `json:"create_at"`
	CpuLimit       int64    `json:"cpu_limit"`
	MemLimit       int64    `json:"mem_limit"`
	Envs           []string `json:"envs"`
	Binds          []string `json:"binds"`
	Timeout        int64    `json:"timeout"`
	HealthPorts    []int    `json:"health_ports"`
	State          string   `json:"state"`
	LastActionTime int64    `json:"last_action_time"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("Error[%d]: %s", e.Code, e.Message)
}
