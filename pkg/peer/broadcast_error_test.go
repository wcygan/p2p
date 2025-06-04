package peer

import (
	"net"
	"testing"

	"example.com/p2p/pkg/message"
)

// TestBroadcastRemovesBrokenConn ensures that a connection that fails during
// broadcast is removed and does not block sending to other peers.
func TestBroadcastRemovesBrokenConn(t *testing.T) {
	p := New("localhost:0")

	goodA, goodB := net.Pipe()
	badA, badB := net.Pipe()

	defer goodA.Close()
	defer goodB.Close()
	badA.Close() // simulate a broken connection
	defer badB.Close()

	p.AddConn("good", goodA)
	p.AddConn("bad", badA)

	msg := &message.Message{SenderID: p.ID, SequenceNo: 1, Payload: "hi"}

	readDone := make(chan *message.Message, 1)
	go func() {
		buf := make([]byte, 512)
		n, err := goodB.Read(buf)
		if err != nil {
			readDone <- nil
			return
		}
		m, _ := message.Unmarshal(buf[:n])
		readDone <- m
	}()

	// Expect an error due to the closed connection
	if err := p.Broadcast(msg); err == nil {
		t.Fatalf("expected error from broadcast")
	}

	if p.Connections() != 1 {
		t.Fatalf("expected 1 remaining connection, got %d", p.Connections())
	}

	m := <-readDone
	if m == nil {
		t.Fatal("failed to read message on good connection")
	}
	if m.Payload != msg.Payload {
		t.Fatalf("expected payload %s, got %s", msg.Payload, m.Payload)
	}
}
