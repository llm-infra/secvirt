package commands

type ProcessInfo struct {
	pid  uint32
	tag  string
	cmd  string
	args []string
	envs map[string]string
	cwd  string
}

type PtySize struct {
	Rows uint32
	Cols uint32
}
