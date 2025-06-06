package peer

import (
	"bufio"
	"net"
	"strings"
	"testing"
)

// TestHandshakeLongID verifies that Handshake returns an error when the remote
// peer sends an ID longer than the allowed limit.
func TestHandshakeLongID(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() {
		_, err := Handshake(c1, "local")
		errCh <- err
	}()

	// Remote side: read the local ID and respond with an overly long ID.
	r := bufio.NewReader(c2)
	if _, err := r.ReadString('\n'); err != nil {
		t.Fatalf("read id: %v", err)
	}
	longID := strings.Repeat("x", 65)
	// Perform the write in a goroutine so we don't block if the Handshake
	// stops reading after the error condition is triggered.
	go func() { _, _ = c2.Write([]byte(longID + "\n")) }()

	if err := <-errCh; err == nil {
		t.Fatal("expected error from long remote id, got nil")
	}
}
