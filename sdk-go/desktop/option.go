package desktop

type Option func(*Options)

type Options struct {
	cwd   string
	envs  map[string]string
	stdin bool
}

func NewOptions(s *Sandbox) *Options {
	return &Options{
		cwd: s.HomeDir(),
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
