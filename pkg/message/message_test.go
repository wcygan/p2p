package message

import "testing"

func TestMarshalUnmarshal(t *testing.T) {
	original := &Message{SenderID: "id1", SequenceNo: 1, Payload: "hello"}
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
