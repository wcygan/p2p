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
