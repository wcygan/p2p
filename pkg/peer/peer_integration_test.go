package peer

import (
	"net"
	"testing"
	"time"

	"example.com/p2p/pkg/message"
)

// TestThreePeerMessaging spins up three peers connected in a line and ensures
// a broadcast from one end reaches the other end.
func TestThreePeerMessaging(t *testing.T) {
	p1 := New("localhost:0")
	ln1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen1: %v", err)
	}
	defer ln1.Close()
	go p1.Serve(ln1)

	p2 := New("localhost:0")
	ln2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen2: %v", err)
	}
	defer ln2.Close()
	go p2.Serve(ln2)

	p3 := New("localhost:0")
	ln3, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen3: %v", err)
	}
	defer ln3.Close()
	go p3.Serve(ln3)

	// Connect peers in a chain: p2 <-> p1 and p3 <-> p2
	if _, err := p2.Connect(ln1.Addr().String()); err != nil {
		t.Fatalf("p2 connect p1: %v", err)
	}
	if _, err := p3.Connect(ln2.Addr().String()); err != nil {
		t.Fatalf("p3 connect p2: %v", err)
	}

	// Wait briefly for all connections to establish
	time.Sleep(100 * time.Millisecond)

	msg := &message.Message{SenderID: p1.ID, SequenceNo: 1, Payload: "hello"}
	if err := p1.Broadcast(msg); err != nil {
		t.Fatalf("broadcast: %v", err)
	}

	// Expect p2 and p3 to each receive the message once
	for _, p := range []*Peer{p2, p3} {
		select {
		case m := <-p.Messages:
			if m.Payload != msg.Payload {
				t.Fatalf("expected payload %s got %s", msg.Payload, m.Payload)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}
	}
}

// TestBroadcastNoLoop verifies that a message broadcast does not bounce back to the sender.
func TestBroadcastNoLoop(t *testing.T) {
	p1 := New("localhost:0")
	p2 := New("localhost:0")

	c1, c2 := net.Pipe()
	p1.HandleConn(p2.ID, c1)
	p2.HandleConn(p1.ID, c2)

	msg := &message.Message{SenderID: p1.ID, SequenceNo: 1, Payload: "hi"}
	if err := p1.Broadcast(msg); err != nil {
		t.Fatalf("broadcast: %v", err)
	}

	select {
	case m := <-p2.Messages:
		if m.Payload != msg.Payload {
			t.Fatalf("expected payload %s got %s", msg.Payload, m.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}

	select {
	case m := <-p1.Messages:
		t.Fatalf("unexpected looped message: %v", m)
	case <-time.After(100 * time.Millisecond):
		// ok - no message received
	}
}
