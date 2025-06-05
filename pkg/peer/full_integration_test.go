package peer

import (
	"net"
	"sync"
	"testing"
	"time"

	"example.com/p2p/pkg/message"
)

// TestFullIntegration tests a complete P2P network scenario with multiple peers
// forming a network, exchanging messages, and handling peer failures
func TestFullIntegration(t *testing.T) {
	// Create 4 peers with listeners
	peers := make([]*Peer, 4)
	listeners := make([]net.Listener, 4)
	
	for i := 0; i < 4; i++ {
		peers[i] = New("localhost:0")
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen peer %d: %v", i, err)
		}
		listeners[i] = ln
		defer ln.Close()
		go peers[i].Serve(ln)
	}
	
	// Connect peers in a chain: 0 <-> 1 <-> 2 <-> 3
	// This creates a linear topology where messages need to propagate
	connections := [][2]int{{0, 1}, {1, 2}, {2, 3}}
	
	for _, conn := range connections {
		from, to := conn[0], conn[1]
		if _, err := peers[from].Connect(listeners[to].Addr().String()); err != nil {
			t.Fatalf("connect peer %d to %d: %v", from, to, err)
		}
	}
	
	// Wait for all connections to establish
	time.Sleep(200 * time.Millisecond)
	
	// Verify connectivity
	expectedConnections := []int{1, 2, 2, 1} // peer 0 has 1 conn, peer 1 has 2 conns, etc.
	for i, expected := range expectedConnections {
		if peers[i].Connections() != expected {
			t.Fatalf("peer %d expected %d connections, got %d", i, expected, peers[i].Connections())
		}
	}
	
	// Broadcast a message from peer 0 and verify it reaches all other peers
	msg := &message.Message{SenderID: peers[0].ID, SequenceNo: 1, Payload: "Hello P2P Network!"}
	if err := peers[0].Broadcast(msg); err != nil {
		t.Fatalf("broadcast from peer 0: %v", err)
	}
	
	// Collect messages received by each peer
	receivedMessages := make(map[int]*message.Message)
	var wg sync.WaitGroup
	wg.Add(3) // peers 1, 2, 3 should receive the message
	
	for i := 1; i < 4; i++ {
		go func(peerIdx int) {
			defer wg.Done()
			select {
			case received := <-peers[peerIdx].Messages:
				receivedMessages[peerIdx] = received
			case <-time.After(2 * time.Second):
				t.Errorf("peer %d did not receive message within timeout", peerIdx)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify all peers received the message
	for i := 1; i < 4; i++ {
		received, ok := receivedMessages[i]
		if !ok {
			t.Fatalf("peer %d did not receive any message", i)
		}
		if received.Payload != msg.Payload {
			t.Fatalf("peer %d received wrong payload: expected %s, got %s", i, msg.Payload, received.Payload)
		}
		if received.SenderID != msg.SenderID {
			t.Fatalf("peer %d received wrong sender: expected %s, got %s", i, msg.SenderID, received.SenderID)
		}
	}
	
	// Test message deduplication by broadcasting the same message again
	// No peer should receive it again since it's already been seen
	if err := peers[0].Broadcast(msg); err != nil {
		t.Fatalf("second broadcast from peer 0: %v", err)
	}
	
	// Wait briefly and verify no duplicate messages
	time.Sleep(100 * time.Millisecond)
	for i := 1; i < 4; i++ {
		select {
		case unexpected := <-peers[i].Messages:
			t.Fatalf("peer %d received duplicate message: %+v", i, unexpected)
		default:
			// Expected - no duplicate message should be received
		}
	}
	
	// Test bidirectional communication by having peer 3 send a message back
	responseMsg := &message.Message{SenderID: peers[3].ID, SequenceNo: 1, Payload: "Response from peer 3"}
	if err := peers[3].Broadcast(responseMsg); err != nil {
		t.Fatalf("broadcast from peer 3: %v", err)
	}
	
	// Verify the response reaches peers 0, 1, and 2
	responseReceived := make(map[int]*message.Message)
	var responseWg sync.WaitGroup
	responseWg.Add(3)
	
	for i := 0; i < 3; i++ {
		go func(peerIdx int) {
			defer responseWg.Done()
			select {
			case received := <-peers[peerIdx].Messages:
				responseReceived[peerIdx] = received
			case <-time.After(2 * time.Second):
				t.Errorf("peer %d did not receive response within timeout", peerIdx)
			}
		}(i)
	}
	
	responseWg.Wait()
	
	// Verify response message received correctly
	for i := 0; i < 3; i++ {
		received, ok := responseReceived[i]
		if !ok {
			t.Fatalf("peer %d did not receive response message", i)
		}
		if received.Payload != responseMsg.Payload {
			t.Fatalf("peer %d received wrong response payload: expected %s, got %s", i, responseMsg.Payload, received.Payload)
		}
	}
}

// TestMessagePropagationSpeed tests how quickly messages propagate through the network
func TestMessagePropagationSpeed(t *testing.T) {
	// Create 3 peers in a line: A <-> B <-> C
	peers := make([]*Peer, 3)
	listeners := make([]net.Listener, 3)
	
	for i := 0; i < 3; i++ {
		peers[i] = New("localhost:0")
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen peer %d: %v", i, err)
		}
		listeners[i] = ln
		defer ln.Close()
		go peers[i].Serve(ln)
	}
	
	// Connect A to B and B to C
	if _, err := peers[0].Connect(listeners[1].Addr().String()); err != nil {
		t.Fatalf("connect A to B: %v", err)
	}
	if _, err := peers[1].Connect(listeners[2].Addr().String()); err != nil {
		t.Fatalf("connect B to C: %v", err)
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// Verify connectivity
	if peers[0].Connections() != 1 || peers[1].Connections() != 2 || peers[2].Connections() != 1 {
		t.Fatalf("unexpected connections: A=%d B=%d C=%d", 
			peers[0].Connections(), peers[1].Connections(), peers[2].Connections())
	}
	
	// Send message from A and measure propagation time
	start := time.Now()
	msg := &message.Message{SenderID: peers[0].ID, SequenceNo: 1, Payload: "Speed test"}
	if err := peers[0].Broadcast(msg); err != nil {
		t.Fatalf("broadcast: %v", err)
	}
	
	// Both B and C should receive the message quickly
	receivedB := false
	receivedC := false
	
	for i := 0; i < 2; i++ {
		select {
		case <-peers[1].Messages:
			receivedB = true
		case <-peers[2].Messages:
			receivedC = true
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message propagation")
		}
	}
	
	propagationTime := time.Since(start)
	if propagationTime > 500*time.Millisecond {
		t.Fatalf("message propagation took too long: %v", propagationTime)
	}
	
	if !receivedB || !receivedC {
		t.Fatalf("not all peers received message: B=%v C=%v", receivedB, receivedC)
	}
}