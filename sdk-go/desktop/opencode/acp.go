package opencode

import (
	"encoding/json"
	"errors"

	"github.com/llm-infra/acp/sdk/go/acp"
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
	defer func() {
		if msg.Part.Reason == "stop" {
			evts = append(evts, acp.NewBlockEndEvent(d.blockID, &acp.Usage{
				PromptTokens:     int64(d.input),
				CompletionTokens: int64(d.output),
			}))
		}
	}()

	switch msg.Part.Type {
	case opencode.PartTypeStepStart:
		if d.blockID == "" {
			d.blockID = msg.SessionID
			evts = append(evts, acp.NewBlockStartEvent(d.blockID))
		}

		evts = append(evts, acp.NewContentStartEvent(msg.Part.MessageID, d.blockID))

	case opencode.PartTypeStepFinish:
		tokens, ok := msg.Part.Tokens.(opencode.StepFinishPartTokens)
		if ok {
			d.input += tokens.Input
			d.output += tokens.Output
			d.output += tokens.Reasoning
		}

		evts = append(evts, acp.NewContentEndEvent(msg.Part.MessageID))

	case opencode.PartTypeTool:
		state, ok := msg.Part.State.(opencode.ToolPartState)
		if !ok {
			return nil, ErrEventFormatError
		}

		if state.Status == opencode.ToolPartStateStatusCompleted {
			evts = append(evts,
				acp.NewContentDeltaEvent(msg.Part.MessageID, acp.NewStreamToolCallContent(msg.Part.Tool)),
				acp.NewContentDeltaEvent(msg.Part.MessageID, acp.NewStreamToolArgsContent(state.JSON.Input.Raw())),
				acp.NewContentDeltaEvent(msg.Part.MessageID, acp.NewStreamToolResultContent(state.JSON.Output.Raw())),
			)
		}

	case opencode.PartTypeText:
		evts = append(evts,
			acp.NewContentDeltaEvent(msg.Part.MessageID, acp.NewStreamTextContent(msg.Part.Text)),
		)
	}

	return evts, nil
}
