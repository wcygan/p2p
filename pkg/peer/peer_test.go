package peer

import (
	"net"
	"testing"
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
