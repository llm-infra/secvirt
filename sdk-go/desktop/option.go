package desktop

type Option func(*Options)

type Options struct {
	cwd  string
	envs map[string]string
}

func WithCwd(cwd string) Option {
	return func(o *Options) { o.cwd = cwd }
}

func WithEnvs(envs map[string]string) Option {
	return func(o *Options) { o.envs = envs }
}
