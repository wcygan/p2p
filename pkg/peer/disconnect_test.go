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
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if p.Connections() == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected connection to be removed, got %d", p.Connections())
}
