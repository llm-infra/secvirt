package codex

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/llm-infra/acp/sdk/go/acp"
)

type ACPDecoder struct {
	blockID   string
	itemIDMap map[string]string
}

func NewACPDecoder() *ACPDecoder {
	return &ACPDecoder{
		itemIDMap: make(map[string]string),
	}
}

func (c *ACPDecoder) Decode(data []byte) ([]acp.Event, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	evts := make([]acp.Event, 0)

	switch msg.Type {
	case MessageTypeThreadStarted:

	case MessageTypeTurnStarted:
		c.blockID = uuid.NewString()
		evts = append(evts, acp.NewBlockStartEvent(c.blockID))

	case MessageTypeTurnCompleted:
		if msg.Usage != nil {
			evts = append(evts, acp.NewBlockEndEvent(c.blockID, &acp.Usage{
				PromptTokens:     msg.Usage.InputTokens,
				CompletionTokens: msg.Usage.OutputTokens,
			}))
		} else {
			evts = append(evts, acp.NewBlockEndEvent(c.blockID, nil))
		}

	// 失败返回
	case MessageTypeTurnFailed:
		evts = append(evts, acp.NewBlockEndEvent(c.blockID, nil))
		return nil, errors.New(msg.Message)

	case MessageTypeItemStarted:
		evts = append(evts, c.ItemEvent(msg)...)

	case MessageTypeItemUpdated:
		fmt.Println()

	case MessageTypeItemCompleted:
		evts = append(evts, c.ItemEvent(msg)...)

	case MessageTypeError:
		fmt.Println()

	default:
		fmt.Println()
	}

	return evts, nil
}

func (c *ACPDecoder) ItemEvent(e Message) []acp.Event {
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
		if e.Type == MessageTypeItemStarted {
			evts = append(evts, acp.NewContentDeltaEvent(contentID,
				acp.NewStreamCommandContent(e.Item.Command)))
			c.itemIDMap[e.Item.ID] = contentID
		}
		if e.Type == MessageTypeItemCompleted {
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
