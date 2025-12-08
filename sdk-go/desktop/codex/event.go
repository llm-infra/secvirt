package codex

import "encoding/json"

const (
	EventTypeThreadStarted = "thread.started"
	EventTypeTurnStarted   = "turn.started"
	EventTypeTurnCompleted = "turn.completed"
	EventTypeTurnFailed    = "turn.failed"
	EventTypeItemStarted   = "item.started"
	EventTypeItemUpdated   = "item.updated"
	EventTypeItemCompleted = "item.completed"
	EventTypeError         = "error"
)

type Decoder struct{}

func (d *Decoder) Decode(data []byte, evt any) error {
	return json.Unmarshal(data, evt)
}

type Event struct {
	Type     string `json:"type"`
	ThreadID string `json:"thread_id,omitempty"` // only for thread.started
	Item     *Item  `json:"item,omitempty"`      // for item.* events
	Usage    *Usage `json:"usage,omitempty"`     // for turn.completed
	Message  string `json:"message,omitempty"`   // error message

	Extra map[string]any `json:"-"`
}

func (e *Event) UnmarshalJSON(b []byte) error {
	type Alias Event
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	delete(raw, "type")
	delete(raw, "thread_id")
	delete(raw, "item")
	delete(raw, "usage")
	delete(raw, "message")

	if len(raw) > 0 {
		e.Extra = raw
	}
	return nil
}

const (
	ItemTypeAgentMessage     = "agent_message"
	ItemTypeReasoning        = "reasoning"
	ItemTypeCommandExecution = "command_execution"
	ItemTypeFileChange       = "file_change"
	ItemTypeMcpToolCall      = "mcp_tool_call"
	ItemTypeWebSearch        = "web_search"
	ItemTypeTodoList         = "todo_list"
)

type Item struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	// agent_message/reasoning
	Text string `json:"text,omitempty"`

	// command_execution
	Command          string `json:"command,omitempty"`
	AggregatedOutput string `json:"aggregated_output,omitempty"`
	Status           string `json:"status,omitempty"`
	ExitCode         int    `json:"exit_code,omitempty"`

	// file_change
	Changes []struct {
		Path string `json:"path,omitempty"`
		Kind string `json:"kind,omitempty"`
	} `json:"changes,omitempty"`

	// todo_list
	Items []struct {
		Text      string `json:"text,omitempty"`
		Completed bool   `json:"completed,omitempty"`
	} `json:"items,omitempty"`

	Extra map[string]any `json:"-"`
}

func (it *Item) UnmarshalJSON(b []byte) error {
	type Alias Item
	aux := &struct{ *Alias }{Alias: (*Alias)(it)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	for _, k := range []string{
		"id", "type", "text",
		"command", "aggregated_output", "status", "exit_code",
		"changes",
		"items",
	} {
		delete(raw, k)
	}
	if len(raw) > 0 {
		it.Extra = raw
	}
	return nil
}

type Usage struct {
	InputTokens       int64 `json:"input_tokens"`
	CachedInputTokens int64 `json:"cached_input_tokens"`
	OutputTokens      int64 `json:"output_tokens"`
}
