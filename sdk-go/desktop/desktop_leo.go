package desktop

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
)

func (s *Sandbox) LeoChat(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[LeoEvent], error) {
	opt := &Options{
		cwd: s.HomeDir(),
	}

	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("leo -p '%s' --output-format stream-json --yolo", content),
		opt.envs,
		opt.cwd,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream[LeoEvent](ctx, handle, &LeoEvent{}), nil
}

const (
	LeoEventTypeInit       = "init"
	LeoEventTypeMessage    = "message"
	LeoEventTypeToolUse    = "tool_use"
	LeoEventTypeToolResult = "tool_result"
	LeoEventTypeResult     = "result"

	LeoToolShellCommand = "run_shell_command"
)

type LeoEvent struct {
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
	ToolName   string          `json:"tool_name,omitempty"`
	ToolID     string          `json:"tool_id,omitempty"`
	Parameters json.RawMessage `json:"parameters,omitempty"` // 不同工具参数不同，用 RawMessage 最稳

	// tool_result only
	// ToolID 复用上面字段
	Status string          `json:"status,omitempty"`
	Output json.RawMessage `json:"output,omitempty"` // 可能是 string 或 object，用 RawMessage 兜底

	// result only
	Stats *LeoStats `json:"stats,omitempty"`
}

type LeoStats struct {
	TotalTokens  int64 `json:"total_tokens"`
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	DurationMs   int   `json:"duration_ms"`
	ToolCalls    int   `json:"tool_calls"`
}

func (e *LeoEvent) Decode(data []byte, evt any) error {
	return json.Unmarshal(data, evt)
}

func (e LeoEvent) IsInit() bool       { return e.Type == LeoEventTypeInit }
func (e LeoEvent) IsMessage() bool    { return e.Type == LeoEventTypeMessage }
func (e LeoEvent) IsToolUse() bool    { return e.Type == LeoEventTypeToolUse }
func (e LeoEvent) IsToolResult() bool { return e.Type == LeoEventTypeToolResult }
func (e LeoEvent) IsResult() bool     { return e.Type == LeoEventTypeResult }
