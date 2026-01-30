package opencode

import (
	"errors"

	"github.com/google/uuid"
	"github.com/llm-infra/acp/sdk/go/acp"
	"github.com/mel2oo/go-dkit/json"
	"github.com/sst/opencode-sdk-go"
)

var ErrEventFormatError = errors.New("event format error")

type ACPDecoder struct {
	blockID string

	input  float64
	output float64
}

func NewACPDecoder() *ACPDecoder {
	return &ACPDecoder{}
}

func (d *ACPDecoder) Decode(data []byte) ([]acp.Event, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	evts := make([]acp.Event, 0)
	switch msg.Part.Type {
	case opencode.PartTypeStepStart:
		if d.blockID == "" {
			d.blockID = msg.SessionID
			evts = append(evts, acp.NewBlockStartEvent(d.blockID))
		}

	case opencode.PartTypeStepFinish:
		tokens, ok := msg.Part.Tokens.(opencode.StepFinishPartTokens)
		if ok {
			d.input += tokens.Input
			d.output += tokens.Output
			d.output += tokens.Reasoning
		}
		if msg.Part.Reason == "stop" {
			evts = append(evts, acp.NewBlockEndEvent(d.blockID, &acp.Usage{
				PromptTokens:     int64(d.input),
				CompletionTokens: int64(d.output),
			}))
		}

	case opencode.PartTypeTool:
		state, ok := msg.Part.State.(opencode.ToolPartState)
		if !ok {
			return nil, ErrEventFormatError
		}

		if state.Status == opencode.ToolPartStateStatusCompleted {
			contentID := uuid.NewString()
			evts = append(evts,
				acp.NewContentStartEvent(contentID, d.blockID),
				acp.NewContentDeltaEvent(contentID, acp.NewStreamToolCallContent(msg.Part.Tool)),
				acp.NewContentDeltaEvent(contentID, acp.NewStreamToolArgsContent(json.MarshalString(state.Input))),
				acp.NewContentDeltaEvent(contentID, acp.NewStreamToolResultContent(state.Output)),
				acp.NewContentEndEvent(contentID),
			)
		}

	case opencode.PartTypeText:
		contentID := uuid.NewString()
		evts = append(evts,
			acp.NewContentStartEvent(contentID, d.blockID),
			acp.NewContentDeltaEvent(contentID, acp.NewStreamTextContent(msg.Part.Text)),
			acp.NewContentEndEvent(contentID),
		)
	}

	return evts, nil
}
