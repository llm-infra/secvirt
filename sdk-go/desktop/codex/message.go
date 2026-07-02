package codex

import "encoding/json"

const (
	MessageTypeThreadStarted = "thread.started"
	MessageTypeTurnStarted   = "turn.started"
	MessageTypeTurnCompleted = "turn.completed"
	MessageTypeTurnFailed    = "turn.failed"
	MessageTypeItemStarted   = "item.started"
	MessageTypeItemUpdated   = "item.updated"
	MessageTypeItemCompleted = "item.completed"
	MessageTypeError         = "error"
)

const (
	ItemTypeAgentMessage     = "agent_message"
	ItemTypeReasoning        = "reasoning"
	ItemTypeCommandExecution = "command_execution"
	ItemTypeFileChange       = "file_change"
	ItemTypeMcpToolCall      = "mcp_tool_call"
	ItemTypeWebSearch        = "web_search"
	ItemTypeTodoList         = "todo_list"
)

type Message struct {
	Type     string `json:"type"`
	ThreadID string `json:"thread_id,omitempty"` // only for thread.started
	Item     *Item  `json:"item,omitempty"`      // for item.* events
	Usage    *Usage `json:"usage,omitempty"`     // for turn.completed
	Message  string `json:"message,omitempty"`   // error message
}

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

	// mcp_tool_call
	ToolName  string `json:"tool_name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Result    string `json:"result,omitempty"`

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
}

func (i *Item) UnmarshalJSON(data []byte) error {
	type itemAlias Item
	var raw struct {
		*itemAlias
		ToolName  json.RawMessage `json:"tool_name"`
		Name      json.RawMessage `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
		Args      json.RawMessage `json:"args"`
		Input     json.RawMessage `json:"input"`
		Result    json.RawMessage `json:"result"`
		Output    json.RawMessage `json:"output"`
	}
	raw.itemAlias = (*itemAlias)(i)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	i.ToolName = firstRawString(raw.ToolName, raw.Name)
	i.Arguments = firstRawValue(raw.Arguments, raw.Args, raw.Input)
	i.Result = firstRawValue(raw.Result, raw.Output)
	return nil
}

func firstRawString(values ...json.RawMessage) string {
	for _, value := range values {
		if len(value) == 0 || string(value) == "null" {
			continue
		}
		var s string
		if err := json.Unmarshal(value, &s); err == nil {
			return s
		}
		return string(value)
	}
	return ""
}

func firstRawValue(values ...json.RawMessage) string {
	for _, value := range values {
		if len(value) == 0 || string(value) == "null" {
			continue
		}
		var s string
		if err := json.Unmarshal(value, &s); err == nil {
			return s
		}
		return string(value)
	}
	return ""
}

type Usage struct {
	InputTokens       int64 `json:"input_tokens"`
	CachedInputTokens int64 `json:"cached_input_tokens"`
	OutputTokens      int64 `json:"output_tokens"`
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
