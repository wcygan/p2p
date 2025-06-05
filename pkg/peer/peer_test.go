package peer

import (
	"net"
	"testing"
	"time"

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

func TestHandshake(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	p1 := New("localhost:0")
	p2 := New("localhost:0")

	var id1, id2 string
	var err1, err2 error
	done := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			err1 = err
			close(done)
			return
		}
		id1, err1 = Handshake(conn, p1.ID)
		conn.Close()
		close(done)
	}()
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	id2, err2 = Handshake(conn, p2.ID)
	conn.Close()
	<-done

	if err1 != nil || err2 != nil {
		t.Fatalf("Handshake errors: %v %v", err1, err2)
	}
	if id1 != p2.ID || id2 != p1.ID {
		t.Fatalf("unexpected ids %s %s", id1, id2)
	}
}

func TestConnectServe(t *testing.T) {
	p1 := New("localhost:0")
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go p1.Serve(ln)

	p2 := New("localhost:0")
	if _, err := p2.Connect(ln.Addr().String()); err != nil {
		t.Fatalf("connect: %v", err)
	}

	// give some time for connection to register
	time.Sleep(100 * time.Millisecond)
	if p1.Connections() != 1 || p2.Connections() != 1 {
		t.Fatalf("expected both peers to have 1 connection, got p1=%d p2=%d", p1.Connections(), p2.Connections())
	}

	msg := &message.Message{SenderID: p2.ID, SequenceNo: 1, Payload: "hi"}
	if err := p2.Broadcast(msg); err != nil {
		t.Fatalf("broadcast: %v", err)
	}

	select {
	case m := <-p1.Messages:
		if m.Payload != msg.Payload {
			t.Fatalf("expected payload %s got %s", msg.Payload, m.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestHandleConn(t *testing.T) {
	p := New("localhost:0")

	// connection that will receive broadcasts
	b1a, b1b := net.Pipe()
	defer b1a.Close()
	defer b1b.Close()
	p.AddConn("peer1", b1a)

	// incoming connection
	inA, inB := net.Pipe()
	defer inA.Close()
	defer inB.Close()
	p.HandleConn("peer2", inB)

	msg := &message.Message{SenderID: "peer2", SequenceNo: 1, Payload: "hello"}
	data, _ := msg.Marshal()
	if _, err := inA.Write(data); err != nil {
		t.Fatalf("write: %v", err)
	}

	// verify broadcast to existing conn
	buf := make([]byte, 512)
	n, err := b1b.Read(buf)
	if err != nil {
		t.Fatalf("read broadcast: %v", err)
	}
	got, err := message.Unmarshal(buf[:n])
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Payload != msg.Payload {
		t.Fatalf("expected %s, got %s", msg.Payload, got.Payload)
	}

	// verify message delivered on channel
	select {
	case m := <-p.Messages:
		if m.Payload != msg.Payload {
			t.Fatalf("expected payload %s, got %s", msg.Payload, m.Payload)
		}
	default:
		t.Fatal("expected message on channel")
	}
}
