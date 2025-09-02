package sandbox

import (
	"fmt"
	"sync/atomic"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	defaultAPIPort   = 8994
	defaultProxyPort = 8993
)

type Option func(*Options)

type Options struct {
	host        string
	user        string
	template    TemplateType
	sandboxID   string
	apiPort     int
	proxyPort   int
	healthPorts []int

	useTelemetry   bool
	tracerProvider trace.TracerProvider
	propagators    propagation.TextMapPropagator
}

func newOptions() *Options {
	return &Options{
		host:      "localhost",
		user:      "default",
		template:  "codeide:latest",
		apiPort:   defaultAPIPort,
		proxyPort: defaultProxyPort,
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

func WithAPIPort(port int) Option {
	return func(o *Options) { o.apiPort = port }
}

func WithProxyPort(port int) Option {
	return func(o *Options) { o.proxyPort = port }
}

func WithHealthPorts(ports []int) Option {
	return func(o *Options) { o.healthPorts = ports }
}

func WithTelemetry(tracerProvider trace.TracerProvider,
	propagators propagation.TextMapPropagator) Option {
	return func(o *Options) {
		o.useTelemetry = true
		o.tracerProvider = tracerProvider
		o.propagators = propagators
	}
}

type TemplateType string

const (
	TemplateCodeIDE TemplateType = "codeide:latest"
	TemplateHostMCP TemplateType = "hostmcp:latest"
	TemplateDesktop TemplateType = "desktop:latest"
)

type SandboxDetail struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	IP          string   `json:"ip"`
	User        string   `json:"user"`
	CreateAt    string   `json:"create_at"`
	CpuLimit    int64    `json:"cpu_limit"`
	MemLimit    uint64   `json:"mem_limit"`
	Envs        []string `json:"envs"`
	Binds       []string `json:"binds"`
	Timeout     int64    `json:"timeout"`
	HealthPorts []int    `json:"health_ports"`
	State       string   `json:"state"`

	RunnerState     string       `json:"runner_state"`
	LastActionTime  int64        `json:"last_action_time"`
	CurrActionCount atomic.Int64 `json:"curr_action_count"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("Error[%d]: %s", e.Code, e.Message)
}
