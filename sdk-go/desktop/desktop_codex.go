package desktop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
	"github.com/llm-infra/acp/sdk/go/acp"
	"github.com/llm-infra/secvirt/sdk-go/desktop/codex"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
)

func (s *Sandbox) SetCodexConfig(ctx context.Context, config *codex.Config,
	opts ...Option) error {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	return s.Filesystem().Write(ctx,
		filepath.Join(opt.cwd, ".codex", "config.toml"),
		data,
	)
}

func (s *Sandbox) CodexChat(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[CodexEvent], error) {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("codex exec %s --skip-git-repo-check --full-auto --json",
			strconv.Quote(content)),
		opt.envs,
		opt.cwd,
		opt.stdin,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream(ctx, handle, &CodexDecoder{}), nil
}

const (
	CodexEventTypeThreadStarted = "thread.started"
	CodexEventTypeTurnStarted   = "turn.started"
	CodexEventTypeTurnCompleted = "turn.completed"
	CodexEventTypeTurnFailed    = "turn.failed"
	CodexEventTypeItemStarted   = "item.started"
	CodexEventTypeItemUpdated   = "item.updated"
	CodexEventTypeItemCompleted = "item.completed"
	CodexEventTypeError         = "error"
)

type CodexDecoder struct{}

func (d *CodexDecoder) Decode(data []byte) (CodexEvent, error) {
	var evt CodexEvent
	err := json.Unmarshal(data, &evt)
	return evt, err
}

type CodexEvent struct {
	Type     string `json:"type"`
	ThreadID string `json:"thread_id,omitempty"` // only for thread.started
	Item     *Item  `json:"item,omitempty"`      // for item.* events
	Usage    *Usage `json:"usage,omitempty"`     // for turn.completed
	Message  string `json:"message,omitempty"`   // error message
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
}

type Usage struct {
	InputTokens       int64 `json:"input_tokens"`
	CachedInputTokens int64 `json:"cached_input_tokens"`
	OutputTokens      int64 `json:"output_tokens"`
}

func (s *Sandbox) CodexChatWithACPStream(ctx context.Context, content string,
	opts ...Option) (*commands.Stream[[]acp.Event], error) {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	handle, err := s.Cmd().Start(ctx,
		fmt.Sprintf("codex exec %s --skip-git-repo-check --full-auto --json",
			strconv.Quote(content)),
		opt.envs,
		opt.cwd,
		opt.stdin,
	)
	if err != nil {
		return nil, err
	}

	return commands.NewStream(ctx, handle, NewCodexAcpDecoder()), nil
}

type CodexAcpDecoder struct {
	blockID   string
	itemIDMap map[string]string
}

func NewCodexAcpDecoder() *CodexAcpDecoder {
	return &CodexAcpDecoder{
		itemIDMap: make(map[string]string),
	}
}

func (c *CodexAcpDecoder) Decode(data []byte) ([]acp.Event, error) {
	var codexEvt CodexEvent
	if err := json.Unmarshal(data, &codexEvt); err != nil {
		return nil, err
	}

	evts := make([]acp.Event, 0)

	switch codexEvt.Type {
	case CodexEventTypeThreadStarted:

	case CodexEventTypeTurnStarted:
		c.blockID = uuid.NewString()
		evts = append(evts, acp.NewBlockStartEvent(c.blockID))

	case CodexEventTypeTurnCompleted:
		if codexEvt.Usage != nil {
			evts = append(evts, acp.NewBlockEndEvent(c.blockID, &acp.Usage{
				PromptTokens:     codexEvt.Usage.InputTokens,
				CompletionTokens: codexEvt.Usage.OutputTokens,
			}))
		} else {
			evts = append(evts, acp.NewBlockEndEvent(c.blockID, nil))
		}

	// 失败返回
	case CodexEventTypeTurnFailed:
		evts = append(evts, acp.NewBlockEndEvent(c.blockID, nil))
		return nil, errors.New(codexEvt.Message)

	case CodexEventTypeItemStarted:
		evts = append(evts, c.ItemEvent(codexEvt)...)

	case CodexEventTypeItemUpdated:
		fmt.Println()

	case CodexEventTypeItemCompleted:
		evts = append(evts, c.ItemEvent(codexEvt)...)

	case CodexEventTypeError:
		fmt.Println()

	default:
		fmt.Println()
	}

	return evts, nil
}

func (c *CodexAcpDecoder) ItemEvent(e CodexEvent) []acp.Event {
	evts := make([]acp.Event, 0)

	contentID := uuid.NewString()
	switch e.Item.Type {
	case ItemTypeAgentMessage:
		evts = append(evts, acp.NewContentDeltaEvent(contentID,
			acp.NewStreamTextContent(e.Item.Text)))
	case ItemTypeReasoning:
		evts = append(evts, acp.NewContentDeltaEvent(contentID,
			acp.NewStreamThinkingContent(e.Item.Text)))
	case ItemTypeCommandExecution:
		if e.Type == CodexEventTypeItemStarted {
			evts = append(evts, acp.NewContentDeltaEvent(contentID,
				acp.NewStreamCommandContent(e.Item.Command)))
			c.itemIDMap[e.Item.ID] = contentID
		}
		if e.Type == CodexEventTypeItemCompleted {
			contentID := c.itemIDMap[e.Item.ID]
			evts = append(evts, acp.NewContentDeltaEvent(contentID,
				acp.NewStreamCommandResultContent(e.Item.AggregatedOutput, e.Item.ExitCode)))
		}

	case ItemTypeFileChange:
		// evts = append(evts, acp.NewCustomEvent(e.Item.Type, acp.WithValue(e.Item.Changes)))
	case ItemTypeMcpToolCall:
		fmt.Println()
	case ItemTypeWebSearch:
		fmt.Println()
	case ItemTypeTodoList:
		// evts = append(evts, acp.NewCustomEvent(e.Item.Type, acp.WithValue(e.Item.Items)))
	}

	return evts
}
