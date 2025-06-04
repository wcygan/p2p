package message

import (
	"encoding/json"
)

// Message represents a chat message exchanged between peers.
type Message struct {
	SenderID   string `json:"sender_id"`
	SequenceNo int    `json:"sequence_no"`
	Payload    string `json:"payload"`
}

// Marshal encodes the message as JSON bytes.
func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal decodes JSON bytes into the message.
func Unmarshal(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
