package message

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of message
type MessageType string

const (
	TypeChat      MessageType = "chat"
	TypeHeartbeat MessageType = "heartbeat"
	TypePeerList  MessageType = "peer_list"
)

// Message represents a message exchanged between peers.
type Message struct {
	SenderID   string      `json:"sender_id"`
	SequenceNo int         `json:"sequence_no"`
	Type       MessageType `json:"type"`
	Payload    string      `json:"payload"`
	Timestamp  time.Time   `json:"timestamp"`
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

// NewChatMessage creates a new chat message
func NewChatMessage(senderID string, sequenceNo int, text string) *Message {
	return &Message{
		SenderID:   senderID,
		SequenceNo: sequenceNo,
		Type:       TypeChat,
		Payload:    text,
		Timestamp:  time.Now(),
	}
}

// NewHeartbeatMessage creates a new heartbeat message
func NewHeartbeatMessage(senderID string, sequenceNo int) *Message {
	return &Message{
		SenderID:   senderID,
		SequenceNo: sequenceNo,
		Type:       TypeHeartbeat,
		Payload:    "ping",
		Timestamp:  time.Now(),
	}
}

// NewPeerListMessage creates a new peer list message
func NewPeerListMessage(senderID string, sequenceNo int, peers []string) *Message {
	peerData, _ := json.Marshal(peers)
	return &Message{
		SenderID:   senderID,
		SequenceNo: sequenceNo,
		Type:       TypePeerList,
		Payload:    string(peerData),
		Timestamp:  time.Now(),
	}
}

// IsHeartbeat returns true if the message is a heartbeat
func (m *Message) IsHeartbeat() bool {
	return m.Type == TypeHeartbeat
}

// IsChatMessage returns true if the message is a chat message
func (m *Message) IsChatMessage() bool {
	return m.Type == TypeChat
}

// GetPeerList extracts peer list from a peer list message
func (m *Message) GetPeerList() ([]string, error) {
	if m.Type != TypePeerList {
		return nil, nil
	}
	
	var peers []string
	if err := json.Unmarshal([]byte(m.Payload), &peers); err != nil {
		return nil, err
	}
	return peers, nil
}
