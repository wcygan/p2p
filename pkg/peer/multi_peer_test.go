package peer

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"example.com/p2p/pkg/message"
)

// TestMultiPeerNetwork tests a network of multiple peers connecting and communicating
func TestMultiPeerNetwork(t *testing.T) {
	const numPeers = 5
	peers := make([]*Peer, numPeers)
	listeners := make([]net.Listener, numPeers)
	
	// Setup peers with listeners
	for i := 0; i < numPeers; i++ {
		peer := New("localhost:0")
		peers[i] = peer
		
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen peer %d: %v", i, err)
		}
		listeners[i] = ln
		
		// Start serving in background
		go func(p *Peer, listener net.Listener) {
			_ = p.Serve(listener)
		}(peer, ln)
	}
	
	// Cleanup
	defer func() {
		for i, ln := range listeners {
			ln.Close()
			// Give peers time to cleanup
			time.Sleep(10 * time.Millisecond)
			t.Logf("Peer %d final connections: %d", i, peers[i].Connections())
		}
	}()
	
	// Connect peers in a chain: 0 -> 1 -> 2 -> 3 -> 4
	for i := 0; i < numPeers-1; i++ {
		_, err := peers[i].Connect(listeners[i+1].Addr().String())
		if err != nil {
			t.Fatalf("connect peer %d to %d: %v", i, i+1, err)
		}
	}
	
	// Add some cross-connections for redundancy: 0 -> 2, 1 -> 3, 2 -> 4
	for i := 0; i < numPeers-2; i++ {
		_, err := peers[i].Connect(listeners[i+2].Addr().String())
		if err != nil {
			t.Fatalf("cross-connect peer %d to %d: %v", i, i+2, err)
		}
	}
	
	// Wait for connections to stabilize
	time.Sleep(100 * time.Millisecond)
	
	// Verify connections - accounting for bidirectional nature
	// 0 -> 1, 0 -> 2 = 2 connections for peer 0
	// 1 -> 2, plus incoming from 0 and 3 = 3 connections for peer 1  
	// 2 -> 3, 2 -> 4, plus incoming from 0, 1 = 4 connections for peer 2
	// 3 -> 4, plus incoming from 1, 2 = 3 connections for peer 3
	// 4 has incoming from 2, 3 = 2 connections for peer 4
	expectedConnections := []int{2, 3, 4, 3, 2} // Based on our connection pattern
	for i, expected := range expectedConnections {
		actual := peers[i].Connections()
		if actual != expected {
			t.Errorf("peer %d: expected %d connections, got %d", i, expected, actual)
		}
	}
	
	// Test message propagation from peer 0
	testMessage := &message.Message{
		SenderID:   peers[0].ID,
		SequenceNo: 1,
		Payload:    "Hello multi-peer network!",
	}
	
	// Broadcast from peer 0
	if err := peers[0].Broadcast(testMessage); err != nil {
		t.Fatalf("broadcast from peer 0: %v", err)
	}
	
	// Verify all other peers receive the message
	timeout := time.After(2 * time.Second)
	received := make([]bool, numPeers)
	received[0] = true // Sender doesn't receive own message
	
	for i := 1; i < numPeers; i++ {
		select {
		case msg := <-peers[i].Messages:
			if msg.Payload == testMessage.Payload && msg.SenderID == testMessage.SenderID {
				received[i] = true
				t.Logf("Peer %d received message: %s", i, msg.Payload)
			}
		case <-timeout:
			t.Fatalf("timeout waiting for message at peer %d", i)
		}
	}
	
	// Verify all peers received the message
	for i, wasReceived := range received {
		if !wasReceived {
			t.Errorf("peer %d did not receive the message", i)
		}
	}
}

// TestMessageDeduplicationAcrossNetwork tests that duplicate messages are properly filtered
func TestMessageDeduplicationAcrossNetwork(t *testing.T) {
	const numPeers = 4
	peers := make([]*Peer, numPeers)
	listeners := make([]net.Listener, numPeers)
	
	// Setup peers in a ring topology
	for i := 0; i < numPeers; i++ {
		peer := New("localhost:0")
		peers[i] = peer
		
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen peer %d: %v", i, err)
		}
		listeners[i] = ln
		
		go func(p *Peer, listener net.Listener) {
			_ = p.Serve(listener)
		}(peer, ln)
	}
	
	defer func() {
		for _, ln := range listeners {
			ln.Close()
		}
	}()
	
	// Create ring: 0 -> 1 -> 2 -> 3 -> 0
	for i := 0; i < numPeers; i++ {
		next := (i + 1) % numPeers
		_, err := peers[i].Connect(listeners[next].Addr().String())
		if err != nil {
			t.Fatalf("connect peer %d to %d: %v", i, next, err)
		}
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// Send same message from multiple peers simultaneously
	testMessage := &message.Message{
		SenderID:   "duplicate-test",
		SequenceNo: 123,
		Payload:    "This message should be deduplicated",
	}
	
	// Send from multiple peers at once
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ { // Send from first 3 peers
		wg.Add(1)
		go func(peerIndex int) {
			defer wg.Done()
			err := peers[peerIndex].Broadcast(testMessage)
			if err != nil {
				t.Errorf("broadcast from peer %d: %v", peerIndex, err)
			}
		}(i)
	}
	wg.Wait()
	
	// Count messages received by the last peer (which shouldn't send)
	receivedCount := 0
	timeout := time.After(1 * time.Second)
	
	for {
		select {
		case msg := <-peers[3].Messages:
			if msg.SenderID == testMessage.SenderID && msg.SequenceNo == testMessage.SequenceNo {
				receivedCount++
				t.Logf("Peer 3 received message #%d", receivedCount)
			}
		case <-timeout:
			goto done
		}
	}
	
done:
	// Should receive exactly one message despite multiple broadcasts
	if receivedCount != 1 {
		t.Errorf("expected to receive 1 deduplicated message, got %d", receivedCount)
	}
}

// TestNetworkPartitionRecovery tests network resilience when connections are lost
func TestNetworkPartitionRecovery(t *testing.T) {
	const numPeers = 4
	peers := make([]*Peer, numPeers)
	listeners := make([]net.Listener, numPeers)
	
	// Setup linear network: 0 - 1 - 2 - 3
	for i := 0; i < numPeers; i++ {
		peer := New("localhost:0")
		peers[i] = peer
		
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen peer %d: %v", i, err)
		}
		listeners[i] = ln
		
		go func(p *Peer, listener net.Listener) {
			_ = p.Serve(listener)
		}(peer, ln)
	}
	
	defer func() {
		for _, ln := range listeners {
			ln.Close()
		}
	}()
	
	// Connect in linear chain
	for i := 0; i < numPeers-1; i++ {
		_, err := peers[i].Connect(listeners[i+1].Addr().String())
		if err != nil {
			t.Fatalf("connect peer %d to %d: %v", i, i+1, err)
		}
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// Verify initial connectivity
	for i := 0; i < numPeers; i++ {
		expected := 1
		if i == 0 || i == numPeers-1 {
			expected = 1 // End peers have 1 connection
		} else {
			expected = 2 // Middle peers have 2 connections
		}
		
		if peers[i].Connections() != expected {
			t.Errorf("peer %d: expected %d connections, got %d", i, expected, peers[i].Connections())
		}
	}
	
	// Simulate partition by closing middle connection (1-2)
	t.Log("Simulating network partition...")
	
	// Disconnect peer 1 from peer 2 by removing the connection
	// This simulates a network partition
	peers[1].RemoveConn(peers[2].ID)
	peers[2].RemoveConn(peers[1].ID)
	
	time.Sleep(50 * time.Millisecond)
	
	// Verify partition
	if peers[1].Connections() != 1 {
		t.Errorf("peer 1 should have 1 connection after partition, got %d", peers[1].Connections())
	}
	if peers[2].Connections() != 1 {
		t.Errorf("peer 2 should have 1 connection after partition, got %d", peers[2].Connections())
	}
	
	// Test that messages don't cross the partition
	testMessage := &message.Message{
		SenderID:   peers[0].ID,
		SequenceNo: 999,
		Payload:    "Partition test message",
	}
	
	// Send from peer 0 (should reach peer 1 but not 2 or 3)
	if err := peers[0].Broadcast(testMessage); err != nil {
		t.Fatalf("broadcast from peer 0: %v", err)
	}
	
	// Peer 1 should receive it
	select {
	case msg := <-peers[1].Messages:
		if msg.Payload != testMessage.Payload {
			t.Errorf("peer 1 got wrong message: %s", msg.Payload)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("peer 1 should have received the message")
	}
	
	// Peer 2 and 3 should NOT receive it (due to partition)
	select {
	case <-peers[2].Messages:
		t.Error("peer 2 should not receive message across partition")
	case <-time.After(200 * time.Millisecond):
		// Expected - no message should arrive
	}
	
	select {
	case <-peers[3].Messages:
		t.Error("peer 3 should not receive message across partition")  
	case <-time.After(200 * time.Millisecond):
		// Expected - no message should arrive
	}
}

// TestConcurrentConnections tests many peers connecting simultaneously
func TestConcurrentConnections(t *testing.T) {
	// Create a central hub peer
	hub := New("localhost:0")
	hubLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("hub listen: %v", err)
	}
	defer hubLn.Close()
	
	go func() {
		_ = hub.Serve(hubLn)
	}()
	
	const numClients = 10
	clients := make([]*Peer, numClients)
	
	// Create client peers
	for i := 0; i < numClients; i++ {
		clients[i] = New("localhost:0")
	}
	
	// Connect all clients to hub concurrently
	var wg sync.WaitGroup
	errors := make(chan error, numClients)
	
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientIndex int) {
			defer wg.Done()
			_, err := clients[clientIndex].Connect(hubLn.Addr().String())
			if err != nil {
				errors <- fmt.Errorf("client %d connect: %w", clientIndex, err)
				return
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for connection errors
	for err := range errors {
		t.Error(err)
	}
	
	// Wait for connections to stabilize
	time.Sleep(200 * time.Millisecond)
	
	// Verify hub has all connections
	if hub.Connections() != numClients {
		t.Errorf("hub: expected %d connections, got %d", numClients, hub.Connections())
	}
	
	// Verify each client has one connection to hub
	for i, client := range clients {
		if client.Connections() != 1 {
			t.Errorf("client %d: expected 1 connection, got %d", i, client.Connections())
		}
	}
	
	// Test broadcast from hub reaches all clients
	testMessage := &message.Message{
		SenderID:   hub.ID,
		SequenceNo: 555,
		Payload:    "Broadcast to all clients",
	}
	
	if err := hub.Broadcast(testMessage); err != nil {
		t.Fatalf("hub broadcast: %v", err)
	}
	
	// Verify all clients receive the message
	timeout := time.After(2 * time.Second)
	received := 0
	
	for i := 0; i < numClients; i++ {
		select {
		case msg := <-clients[i].Messages:
			if msg.Payload == testMessage.Payload {
				received++
			}
		case <-timeout:
			t.Errorf("timeout waiting for message at client %d", i)
		}
	}
	
	if received != numClients {
		t.Errorf("expected %d clients to receive message, got %d", numClients, received)
	}
}

// TestLoadTesting tests system behavior under high message load
func TestLoadTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}
	
	const (
		numPeers    = 3
		numMessages = 100
		concurrent  = 5
	)
	
	peers := make([]*Peer, numPeers)
	listeners := make([]net.Listener, numPeers)
	
	// Setup fully connected network
	for i := 0; i < numPeers; i++ {
		peer := New("localhost:0")
		peers[i] = peer
		
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen peer %d: %v", i, err)
		}
		listeners[i] = ln
		
		go func(p *Peer, listener net.Listener) {
			_ = p.Serve(listener)
		}(peer, ln)
	}
	
	defer func() {
		for _, ln := range listeners {
			ln.Close()
		}
	}()
	
	// Connect all peers to each other
	for i := 0; i < numPeers; i++ {
		for j := i + 1; j < numPeers; j++ {
			_, err := peers[i].Connect(listeners[j].Addr().String())
			if err != nil {
				t.Fatalf("connect peer %d to %d: %v", i, j, err)
			}
		}
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// Send messages concurrently
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	var wg sync.WaitGroup
	messagesSent := int64(0)
	sendErrors := int64(0)
	
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for msgID := 0; msgID < numMessages; msgID++ {
				select {
				case <-ctx.Done():
					return
				default:
				}
				
				peerIndex := (goroutineID + msgID) % numPeers
				msg := &message.Message{
					SenderID:   peers[peerIndex].ID,
					SequenceNo: msgID + (goroutineID * numMessages),
					Payload:    fmt.Sprintf("Load test message %d from goroutine %d", msgID, goroutineID),
				}
				
				if err := peers[peerIndex].Broadcast(msg); err != nil {
					sendErrors++
					t.Logf("broadcast error: %v", err)
				} else {
					messagesSent++
				}
				
				// Small delay to prevent overwhelming
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Drain remaining messages with timeout
	drainTimeout := time.After(5 * time.Second)
	totalReceived := 0
	
drainLoop:
	for {
		select {
		case <-drainTimeout:
			break drainLoop
		default:
			// Check all peers for messages
			messageFound := false
			for i := 0; i < numPeers; i++ {
				select {
				case <-peers[i].Messages:
					totalReceived++
					messageFound = true
				default:
				}
			}
			if !messageFound {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
	
	t.Logf("Load test results:")
	t.Logf("  Messages sent: %d", messagesSent)
	t.Logf("  Send errors: %d", sendErrors)
	t.Logf("  Messages received: %d", totalReceived)
	t.Logf("  Success rate: %.2f%%", float64(messagesSent-sendErrors)/float64(messagesSent)*100)
	
	// Basic sanity checks
	if messagesSent == 0 {
		t.Error("no messages were sent")
	}
	
	if sendErrors > messagesSent/10 { // Allow up to 10% error rate
		t.Errorf("too many send errors: %d/%d", sendErrors, messagesSent)
	}
}