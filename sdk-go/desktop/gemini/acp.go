package gemini

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/llm-infra/acp/sdk/go/acp"
	"github.com/mel2oo/go-dkit/json"
)

var ErrInvalidToolParameters = errors.New("gemini tool_call event parameters error")

type ACPDecoder struct {
	sessionID string
	contentID string
	toolCalls map[string]toolCall
}

type toolCall struct {
	contentID string
	toolName  string
}

func NewACPDecoder() *ACPDecoder {
	return &ACPDecoder{
		toolCalls: make(map[string]toolCall),
	}
}

func (d *ACPDecoder) Decode(data []byte) ([]acp.Event, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	evts := make([]acp.Event, 0)

	switch msg.Type {
	case MessageTypeInit:
		d.sessionID = msg.SessionID
		evts = append(evts, acp.NewBlockStartEvent(d.sessionID))

	case MessageTypeMessage:
		if msg.Role == acp.RoleAssistant {
			if d.contentID == "" {
				d.contentID = uuid.NewString()
				evts = append(evts, acp.NewContentStartEvent(d.contentID, d.sessionID))
			}

			evts = append(evts, acp.NewContentDeltaEvent(d.contentID,
				acp.NewStreamTextContent(msg.Content)))
		}

	case MessageTypeResult:
		if d.contentID != "" {
			evts = append(evts, acp.NewContentEndEvent(d.contentID))
			d.contentID = ""
		}

		if msg.Stats != nil {
			evts = append(evts, acp.NewBlockEndEvent(d.sessionID, &acp.Usage{
				PromptTokens:     msg.Stats.InputTokens,
				CompletionTokens: msg.Stats.OutputTokens,
			}))
		} else {
			evts = append(evts, acp.NewBlockEndEvent(d.sessionID, nil))
		}

		if msg.Status == "error" {
			return evts, fmt.Errorf("%s", msg.Error.Message)
		}

	case MessageTypeToolUse:
		if d.contentID != "" {
			evts = append(evts, acp.NewContentEndEvent(d.contentID))
			d.contentID = ""
		}

		toolcall, ok := d.toolCalls[msg.ToolID]
		if !ok {
			toolcall.contentID = uuid.NewString()
			toolcall.toolName = msg.ToolName
			d.toolCalls[msg.ToolID] = toolcall
		}

		evts = append(evts, acp.NewContentStartEvent(toolcall.contentID, d.sessionID))

		switch toolcall.toolName {
		case ToolShellCommand:
			command, ok := msg.Parameters["command"].(string)
			if !ok {
				return nil, ErrInvalidToolParameters
			}

			evts = append(evts, acp.NewContentDeltaEvent(toolcall.contentID,
				acp.NewStreamCommandContent(command)))

		case ToolGoogleWebSearch:
			query, ok := msg.Parameters["query"].(string)
			if !ok {
				return nil, ErrInvalidToolParameters
			}

			evts = append(evts, acp.NewContentDeltaEvent(toolcall.contentID,
				acp.NewStreamWebSearchContent(query)))

		default:
			evts = append(evts, acp.NewContentDeltaEvent(toolcall.contentID,
				acp.NewStreamToolCallContent(msg.ToolName)))
			evts = append(evts, acp.NewContentDeltaEvent(toolcall.contentID,
				acp.NewStreamToolArgsContent(json.MarshalString(msg.Parameters))))
		}

	case MessageTypeToolResult:
		toolcall, ok := d.toolCalls[msg.ToolID]
		if ok {
			switch toolcall.toolName {
			case ToolShellCommand:
				if msg.Error != nil {
					evts = append(evts, acp.NewContentDeltaEvent(toolcall.contentID,
						acp.NewStreamCommandErrorContent(
							&acp.Error{
								Type:    msg.Error.Type,
								Message: msg.Error.Message,
							},
						),
					))
				} else {
					evts = append(evts, acp.NewContentDeltaEvent(toolcall.contentID,
						acp.NewStreamCommandResultContent(string(msg.Output), 0)))
				}

			case ToolGoogleWebSearch:
				if msg.Error != nil {
					evts = append(evts, acp.NewContentDeltaEvent(toolcall.contentID,
						acp.NewStreamWebSearchErrorContent(
							&acp.Error{
								Type:    msg.Error.Type,
								Message: msg.Error.Message,
							},
						),
					))
				} else {
					type webSearch struct {
						Query   string                `json:"query"`
						Answer  string                `json:"answer"`
						Results []acp.WebSearchResult `json:"results"`
					}

					var ws webSearch
					if err := json.Unmarshal([]byte(msg.Output), &ws); err != nil {
						return nil, err
					}

					evts = append(evts, acp.NewContentDeltaEvent(toolcall.contentID,
						acp.NewStreamWebSearchResultContent(ws.Answer, ws.Results)))
				}

			default:
				if msg.Error != nil {
					evts = append(evts, acp.NewContentDeltaEvent(toolcall.contentID,
						acp.NewStreamToolErrorContent(
							&acp.Error{
								Type:    msg.Error.Type,
								Message: msg.Error.Message,
							},
						),
					))
				} else {
					evts = append(evts, acp.NewContentDeltaEvent(toolcall.contentID,
						acp.NewStreamToolResultContent(string(msg.Output))))
				}
			}

			evts = append(evts, acp.NewContentEndEvent(toolcall.contentID))
			delete(d.toolCalls, msg.ToolID)
		}

	case MessageTypeError:
		return nil, fmt.Errorf(" call error, severity: %s, message: %s", msg.Severity, msg.Message)

	default:
		if d.contentID != "" {
			evts = append(evts, acp.NewContentEndEvent(d.contentID))
			d.contentID = ""
		}
	}

	return evts, nil
}
