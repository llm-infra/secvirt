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
		fmt.Sprintf("leo -p '%s' --output-format stream-json", content),
		opt.envs,
		opt.cwd,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream[LeoEvent](ctx, handle, &LeoEvent{}), nil
}

const (
	LeoEventTypeInit    = "init"
	LeoEventTypeMessage = "message"
	LeoEventTypeResult  = "result"
)

// {"type":"init","timestamp":"2025-12-08T05:39:21.144Z","session_id":"0bfe8398-a019-420a-b91c-a89aab1e3b30","model":"glm-4.6"}
// {"type":"message","timestamp":"2025-12-08T05:39:21.145Z","role":"user","content":"你好"}
// {"type":"message","timestamp":"2025-12-08T05:39:25.139Z","role":"assistant","content":"\n","delta":true}
// {"type":"message","timestamp":"2025-12-08T05:39:25.174Z","role":"assistant","content":"你好","delta":true}
// {"type":"message","timestamp":"2025-12-08T05:39:25.175Z","role":"assistant","content":"！","delta":true}
// {"type":"message","timestamp":"2025-12-08T05:39:25.175Z","role":"assistant","content":"有什么","delta":true}
// {"type":"message","timestamp":"2025-12-08T05:39:25.209Z","role":"assistant","content":"可以帮助","delta":true}
// {"type":"message","timestamp":"2025-12-08T05:39:25.260Z","role":"assistant","content":"你的","delta":true}
// {"type":"message","timestamp":"2025-12-08T05:39:25.260Z","role":"assistant","content":"吗","delta":true}
// {"type":"message","timestamp":"2025-12-08T05:39:25.260Z","role":"assistant","content":"？","delta":true}
// {"type":"result","timestamp":"2025-12-08T05:39:26.632Z","status":"success","stats":{"total_tokens":11232,"input_tokens":11122,"output_tokens":110,"duration_ms":5488,"tool_calls":0}}
type LeoEvent struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id,omitempty"`
	Role      string `json:"role,omitempty"`
	Content   string `json:"content,omitempty"`
	Delta     bool   `json:"delta,omitempty"`
	Status    string `json:"status,omitempty"`
	Stats     *Stats `json:"stats,omitempty"`
}

type Stats struct {
	TotalTokens  int64 `json:"total_tokens"`
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	DurationMs   int   `json:"duration_ms"`
	ToolCalls    int   `json:"tool_calls"`
}

func (e *LeoEvent) Decode(data []byte, evt any) error {
	return json.Unmarshal(data, evt)
}
