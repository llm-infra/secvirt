package desktop

import "time"

type Option func(*Options)

type Options struct {
	cwd               string
	envs              map[string]string
	stdin             bool
	agent             string
	sessionID         string
	ocServerRetries   int
	ocServerRetryWait time.Duration
}

func NewOptions(s *Sandbox) *Options {
	return &Options{
		cwd:               s.HomeDir(),
		envs:              make(map[string]string),
		ocServerRetries:   3,
		ocServerRetryWait: time.Second,
	}
}

func WithCwd(cwd string) Option {
	return func(o *Options) { o.cwd = cwd }
}

func WithEnvs(envs map[string]string) Option {
	return func(o *Options) { o.envs = envs }
}

func WithStdin(stdin bool) Option {
	return func(o *Options) { o.stdin = stdin }
}

func WithAgent(agent string) Option {
	return func(o *Options) { o.agent = agent }
}

func WithSessionID(sessionID string) Option {
	return func(o *Options) { o.sessionID = sessionID }
}

func WithOcServerRetry(retries int, wait time.Duration) Option {
	return func(o *Options) {
		o.ocServerRetries = retries
		o.ocServerRetryWait = wait
	}
}
