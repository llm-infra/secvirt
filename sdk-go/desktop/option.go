package desktop

type Option func(*Options)

type Options struct {
	cwd       string
	envs      map[string]string
	stdin     bool
	agent     string
	sessionID string
}

func NewOptions(s *Sandbox) *Options {
	return &Options{
		cwd:  s.HomeDir(),
		envs: make(map[string]string),
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
