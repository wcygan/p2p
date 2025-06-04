package peer

import (
	"net"
	"testing"
	"time"
)

// TestHandleConnRemoteClose ensures a closed connection is removed.
func TestHandleConnRemoteClose(t *testing.T) {
	p := New("localhost:0")
	a, b := net.Pipe()
	p.HandleConn("peer2", a)
	b.Close()
	time.Sleep(50 * time.Millisecond)
	if p.Connections() != 0 {
		t.Fatalf("expected connection to be removed, got %d", p.Connections())
	}
}
