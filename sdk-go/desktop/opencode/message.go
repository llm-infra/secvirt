package opencode

const (
	ToolBash  = "bash"
	ToolEdit  = "edit"
	ToolWrite = "write"
	ToolSkill = "skill"
)

type BashArgs struct {
	Command     string `json:"command" mapstructure:"command"`
	Description string `json:"description" mapstructure:"description"`
}

type EditArgs struct {
	FilePath  string `json:"filePath" mapstructure:"filePath"`
	NewString string `json:"newString" mapstructure:"newString"`
	OldString string `json:"oldString" mapstructure:"oldString"`
}

type WriteArgs struct {
	FilePath string `json:"filePath" mapstructure:"filePath"`
	Content  string `json:"content" mapstructure:"content"`
}

type SkillArgs struct {
	Name string `json:"name" mapstructure:"name"`
}
