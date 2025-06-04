package peer

import (
	"net"
	"testing"

	"example.com/p2p/pkg/message"
)

func TestNewPeer(t *testing.T) {
	p := New("localhost:0")
	if p.ID == "" {
		t.Fatal("expected peer ID to be set")
	}
}

func TestAddRemoveConn(t *testing.T) {
	p := New("localhost:0")
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	p.AddConn("peer2", c1)
	if p.Connections() != 1 {
		t.Fatalf("expected 1 connection, got %d", p.Connections())
	}
	p.RemoveConn("peer2")
	if p.Connections() != 0 {
		t.Fatalf("expected 0 connections, got %d", p.Connections())
	}
}

func TestBroadcast(t *testing.T) {
	p := New("localhost:0")
	c1a, c1b := net.Pipe()
	c2a, c2b := net.Pipe()
	defer c1a.Close()
	defer c1b.Close()
	defer c2a.Close()
	defer c2b.Close()

	p.AddConn("peer1", c1a)
	p.AddConn("peer2", c2a)

	msg := &message.Message{SenderID: p.ID, SequenceNo: 1, Payload: "hi"}

	done := make(chan *message.Message, 2)
	go func() {
		buf := make([]byte, 512)
		n, _ := c1b.Read(buf)
		m, _ := message.Unmarshal(buf[:n])
		done <- m
	}()
	go func() {
		buf := make([]byte, 512)
		n, _ := c2b.Read(buf)
		m, _ := message.Unmarshal(buf[:n])
		done <- m
	}()

	if err := p.Broadcast(msg); err != nil {
		t.Fatalf("broadcast: %v", err)
	}

	m1 := <-done
	m2 := <-done

	if m1.Payload != msg.Payload || m2.Payload != msg.Payload {
		t.Fatalf("expected payload %s, got %s and %s", msg.Payload, m1.Payload, m2.Payload)
	}
}

func TestSeen(t *testing.T) {
	p := New("localhost:0")
	msg := &message.Message{SenderID: "p1", SequenceNo: 1, Payload: "hi"}
	if p.Seen(msg) {
		t.Fatalf("first time message should be unseen")
	}
	if !p.Seen(msg) {
		t.Fatalf("second time message should be seen")
	}
}
