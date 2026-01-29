package opencode

import (
	"encoding/json"

	"github.com/sst/opencode-sdk-go"
)

type Message struct {
	Type      string        `json:"type"`
	Timestamp int64         `json:"timestamp"`
	SessionID string        `json:"sessionID"`
	Part      opencode.Part `json:"part"`
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
