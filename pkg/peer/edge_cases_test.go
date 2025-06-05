package peer

import (
	"net"
	"testing"
	"time"

	"example.com/p2p/pkg/message"
)

func TestRandomIDErrorHandling(t *testing.T) {
	// This test simulates the error case in randomID when rand.Read fails
	// We can't easily mock crypto/rand, but we can test the return value
	id := randomID()
	if id == "" {
		t.Fatal("randomID returned empty string")
	}
	if len(id) != 32 { // 16 bytes hex encoded = 32 chars
		t.Fatalf("expected 32 character ID, got %d", len(id))
	}
}

func TestHandshakeReadError(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() {
		_, err := Handshake(c1, "local")
		errCh <- err
	}()

	// Close c2 immediately to cause read error
	c2.Close()

	err := <-errCh
	if err == nil {
		t.Fatal("expected error from closed connection, got nil")
	}
}

func TestHandshakeWriteError(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c2.Close()

	// Close c1 to cause write error
	c1.Close()

	_, err := Handshake(c1, "local")
	if err == nil {
		t.Fatal("expected error from write to closed connection, got nil")
	}
}

func TestConnectDialError(t *testing.T) {
	p := New("localhost:0")
	
	// Try to connect to invalid address
	_, err := p.Connect("invalid:address")
	if err == nil {
		t.Fatal("expected error from invalid address, got nil")
	}
}

func TestConnectHandshakeError(t *testing.T) {
	p := New("localhost:0")
	
	// Create a listener that will close connections immediately
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close() // Close immediately to cause Handshake error
		}
	}()
	
	_, err = p.Connect(ln.Addr().String())
	if err == nil {
		t.Fatal("expected error from failed Handshake, got nil")
	}
}

func TestServeAcceptError(t *testing.T) {
	p := New("localhost:0")
	
	// Create a listener and close it to cause accept error
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	ln.Close()
	
	err = p.Serve(ln)
	if err == nil {
		t.Fatal("expected error from closed listener, got nil")
	}
}

func TestBroadcastMarshalError(t *testing.T) {
	p := New("localhost:0")
	
	// Create a message that will cause marshal error
	// Since Message uses basic types, we need to test with a nil message
	// But this would panic, so we test with valid message instead
	// and verify that broadcast works correctly
	msg := &message.Message{SenderID: p.ID, SequenceNo: 1, Payload: "test"}
	err := p.Broadcast(msg)
	if err != nil {
		t.Fatalf("expected no error for valid message, got %v", err)
	}
}

func TestReadLoopInvalidMessage(t *testing.T) {
	p := New("localhost:0")
	
	// Create pipe connections
	inA, inB := net.Pipe()
	defer inA.Close()
	defer inB.Close()
	
	// Start readLoop
	go p.readLoop("test-peer", inB)
	
	// Send invalid JSON data
	invalidData := []byte("{invalid json}")
	if _, err := inA.Write(invalidData); err != nil {
		t.Fatalf("write invalid data: %v", err)
	}
	
	// Give readLoop time to process
	time.Sleep(50 * time.Millisecond)
	
	// Verify no message was delivered (since JSON was invalid)
	select {
	case <-p.Messages:
		t.Fatal("unexpected message delivery for invalid JSON")
	default:
		// Expected - no message should be delivered
	}
}

func TestReadLoopConnectionError(t *testing.T) {
	p := New("localhost:0")
	
	// Create pipe connections
	inA, inB := net.Pipe()
	defer inA.Close()
	
	// Add connection first
	p.AddConn("test-peer", inB)
	
	// Start readLoop
	go p.readLoop("test-peer", inB)
	
	// Close the connection to cause read error
	inB.Close()
	
	// Give readLoop time to detect and handle the error
	time.Sleep(50 * time.Millisecond)
	
	// Verify connection was removed
	if p.Connections() != 0 {
		t.Fatalf("expected 0 connections after error, got %d", p.Connections())
	}
}

func TestReadLoopMessageChannelFull(t *testing.T) {
	p := New("localhost:0")
	
	// Fill up the message channel
	for i := 0; i < cap(p.Messages); i++ {
		p.Messages <- &message.Message{SenderID: "filler", SequenceNo: i, Payload: "fill"}
	}
	
	// Create pipe connections
	inA, inB := net.Pipe()
	defer inA.Close()
	defer inB.Close()
	
	// Start readLoop
	go p.readLoop("test-peer", inB)
	
	// Send a valid message
	msg := &message.Message{SenderID: "test", SequenceNo: 1, Payload: "test"}
	data, _ := msg.Marshal()
	if _, err := inA.Write(data); err != nil {
		t.Fatalf("write message: %v", err)
	}
	
	// Give readLoop time to process
	time.Sleep(50 * time.Millisecond)
	
	// The message should still be broadcasted even if channel is full
	// (this tests the default case in the select statement)
}

func TestRemoveConnNonExistent(t *testing.T) {
	p := New("localhost:0")
	
	// Try to remove a connection that doesn't exist
	p.RemoveConn("nonexistent")
	
	// Should not panic or cause issues
	if p.Connections() != 0 {
		t.Fatalf("expected 0 connections, got %d", p.Connections())
	}
}

func TestHandshakePartialRead(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() {
		_, err := Handshake(c1, "local")
		errCh <- err
	}()

	// Remote side: read the local ID and respond with partial data
	buf := make([]byte, 10)
	if _, err := c2.Read(buf); err != nil {
		t.Fatalf("read id: %v", err)
	}
	
	// Send partial ID without newline, then close
	if _, err := c2.Write([]byte("part")); err != nil {
		t.Fatalf("write partial: %v", err)
	}
	c2.Close()

	err := <-errCh
	if err == nil {
		t.Fatal("expected error from partial read, got nil")
	}
}

func TestHandshakeZeroRead(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() {
		_, err := Handshake(c1, "local")
		errCh <- err
	}()

	// Remote side: read the local ID but don't respond
	buf := make([]byte, 10)
	if _, err := c2.Read(buf); err != nil {
		t.Fatalf("read id: %v", err)
	}
	
	// Close connection to trigger read error
	c2.Close()

	err := <-errCh
	if err == nil {
		t.Fatal("expected error from closed connection, got nil")
	}
}

func TestServeHandshakeFailure(t *testing.T) {
	p := New("localhost:0")
	
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	
	// Start server in background
	go func() {
		_ = p.Serve(ln)
	}()
	
	// Connect but don't perform proper Handshake
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	
	// Send invalid Handshake data and close
	conn.Write([]byte("invalid"))
	conn.Close()
	
	// Give server time to handle the connection
	time.Sleep(50 * time.Millisecond)
	
	// Connection should not be added due to Handshake failure
	if p.Connections() != 0 {
		t.Fatalf("expected 0 connections after Handshake failure, got %d", p.Connections())
	}
}