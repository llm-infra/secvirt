package gemini

import "encoding/json"

const (
	MessageTypeInit       = "init"
	MessageTypeMessage    = "message"
	MessageTypeToolUse    = "tool_use"
	MessageTypeToolResult = "tool_result"
	MessageTypeResult     = "result"
	MessageTypeError      = "error"

	ToolShellCommand    = "run_shell_command"
	ToolGoogleWebSearch = "google_web_search"

	StatusSuccess = "success"
	StatusError   = "error"
)

type Message struct {
	// common
	Type      string `json:"type"`
	Timestamp string `json:"timestamp,omitempty"`

	// init only
	SessionID string `json:"session_id,omitempty"`
	Model     string `json:"model,omitempty"`

	// message only
	Role    string `json:"role,omitempty"` // user / assistant
	Content string `json:"content,omitempty"`
	Delta   bool   `json:"delta,omitempty"`

	// tool_use only
	ToolName   string         `json:"tool_name,omitempty"`
	ToolID     string         `json:"tool_id,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"`

	// tool_result only
	Output string `json:"output,omitempty"`

	// result only
	Stats *Stats `json:"stats,omitempty"`

	// tool_result and result
	Status string `json:"status,omitempty"`
	Error  *struct {
		Type    string `json:"type,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`

	// error only
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message,omitempty"`
}

type Stats struct {
	TotalTokens  int64 `json:"total_tokens"`
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	DurationMs   int   `json:"duration_ms"`
	ToolCalls    int   `json:"tool_calls"`
}

type Decoder struct{}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Decode(data []byte) (Message, error) {
	var evt Message
	err := json.Unmarshal(data, &evt)
	return evt, err
}
