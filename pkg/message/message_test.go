package message

import "testing"

func TestMarshalUnmarshal(t *testing.T) {
	original := NewChatMessage("id1", 1, "hello")
	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	decoded, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.SenderID != original.SenderID || decoded.SequenceNo != original.SequenceNo || decoded.Payload != original.Payload {
		t.Fatalf("decoded message does not match original")
	}
	if decoded.Type != TypeChat {
		t.Fatalf("expected type %s, got %s", TypeChat, decoded.Type)
	}
}

func TestUnmarshalInvalidJSON(t *testing.T) {
	invalidData := []byte("{invalid json")
	_, err := Unmarshal(invalidData)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestUnmarshalEmptyData(t *testing.T) {
	_, err := Unmarshal([]byte{})
	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}
}

func TestNewChatMessage(t *testing.T) {
	msg := NewChatMessage("peer1", 42, "hello world")
	
	if msg.SenderID != "peer1" {
		t.Errorf("expected sender ID 'peer1', got %s", msg.SenderID)
	}
	if msg.SequenceNo != 42 {
		t.Errorf("expected sequence number 42, got %d", msg.SequenceNo)
	}
	if msg.Type != TypeChat {
		t.Errorf("expected type %s, got %s", TypeChat, msg.Type)
	}
	if msg.Payload != "hello world" {
		t.Errorf("expected payload 'hello world', got %s", msg.Payload)
	}
	if msg.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
	if !msg.IsChatMessage() {
		t.Error("expected IsChatMessage to return true")
	}
	if msg.IsHeartbeat() {
		t.Error("expected IsHeartbeat to return false")
	}
}

func TestNewHeartbeatMessage(t *testing.T) {
	msg := NewHeartbeatMessage("peer2", 10)
	
	if msg.SenderID != "peer2" {
		t.Errorf("expected sender ID 'peer2', got %s", msg.SenderID)
	}
	if msg.SequenceNo != 10 {
		t.Errorf("expected sequence number 10, got %d", msg.SequenceNo)
	}
	if msg.Type != TypeHeartbeat {
		t.Errorf("expected type %s, got %s", TypeHeartbeat, msg.Type)
	}
	if msg.Payload != "ping" {
		t.Errorf("expected payload 'ping', got %s", msg.Payload)
	}
	if !msg.IsHeartbeat() {
		t.Error("expected IsHeartbeat to return true")
	}
	if msg.IsChatMessage() {
		t.Error("expected IsChatMessage to return false")
	}
}

func TestNewPeerListMessage(t *testing.T) {
	peers := []string{"peer1:8080", "peer2:8081", "peer3:8082"}
	msg := NewPeerListMessage("peer0", 5, peers)
	
	if msg.SenderID != "peer0" {
		t.Errorf("expected sender ID 'peer0', got %s", msg.SenderID)
	}
	if msg.Type != TypePeerList {
		t.Errorf("expected type %s, got %s", TypePeerList, msg.Type)
	}
	
	// Test extracting peer list
	extractedPeers, err := msg.GetPeerList()
	if err != nil {
		t.Fatalf("extract peer list: %v", err)
	}
	
	if len(extractedPeers) != len(peers) {
		t.Fatalf("expected %d peers, got %d", len(peers), len(extractedPeers))
	}
	
	for i, peer := range peers {
		if extractedPeers[i] != peer {
			t.Errorf("expected peer %s, got %s", peer, extractedPeers[i])
		}
	}
}

func TestGetPeerListFromNonPeerListMessage(t *testing.T) {
	msg := NewChatMessage("peer1", 1, "hello")
	peers, err := msg.GetPeerList()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if peers != nil {
		t.Errorf("expected nil peers for non-peer-list message, got %v", peers)
	}
}

func TestGetPeerListInvalidJSON(t *testing.T) {
	msg := &Message{
		SenderID:   "peer1",
		SequenceNo: 1,
		Type:       TypePeerList,
		Payload:    "{invalid json",
	}
	
	_, err := msg.GetPeerList()
	if err == nil {
		t.Fatal("expected error for invalid JSON in peer list")
	}
}
