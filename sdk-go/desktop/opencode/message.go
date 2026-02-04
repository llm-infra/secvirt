package opencode

const (
	ToolBash  = "bash"
	ToolWrite = "write"
)

type BashArgs struct {
	Command     string `json:"command" mapstructure:"command"`
	Description string `json:"description" mapstructure:"description"`
}

type WriteArgs struct {
	FilePath string `json:"filePath" mapstructure:"filePath"`
	Content  string `json:"content" mapstructure:"content"`
}
