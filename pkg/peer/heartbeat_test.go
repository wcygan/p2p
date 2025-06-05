package peer

import (
	"testing"
	"time"

	"example.com/p2p/pkg/config"
	"example.com/p2p/pkg/message"
)

func TestNewHeartbeatManager(t *testing.T) {
	cfg := config.Default()
	deadPeers := make([]string, 0)
	onPeerDead := func(peerID string) {
		deadPeers = append(deadPeers, peerID)
	}
	
	hm := NewHeartbeatManager(cfg, "test-peer", onPeerDead)
	
	if hm.peerID != "test-peer" {
		t.Errorf("expected peer ID 'test-peer', got %s", hm.peerID)
	}
	
	if hm.GetPeerCount() != 0 {
		t.Errorf("expected 0 peers initially, got %d", hm.GetPeerCount())
	}
	
	sent, received, active := hm.GetPeerStats()
	if sent != 0 || received != 0 || active != 0 {
		t.Errorf("expected all stats to be 0, got sent=%d received=%d active=%d", sent, received, active)
	}
}

func TestAddRemovePeer(t *testing.T) {
	cfg := config.Default()
	hm := NewHeartbeatManager(cfg, "test-peer", nil)
	
	// Add a peer
	hm.AddPeer("peer1", "localhost:8080", nil)
	
	if hm.GetPeerCount() != 1 {
		t.Errorf("expected 1 peer after adding, got %d", hm.GetPeerCount())
	}
	
	peerList := hm.GetPeerList()
	if len(peerList) != 1 || peerList[0] != "localhost:8080" {
		t.Errorf("expected peer list [localhost:8080], got %v", peerList)
	}
	
	// Remove the peer
	hm.RemovePeer("peer1")
	
	if hm.GetPeerCount() != 0 {
		t.Errorf("expected 0 peers after removing, got %d", hm.GetPeerCount())
	}
}

func TestUpdateLastSeen(t *testing.T) {
	cfg := config.Default()
	hm := NewHeartbeatManager(cfg, "test-peer", nil)
	
	// Add a peer
	hm.AddPeer("peer1", "localhost:8080", nil)
	
	// Get initial last seen time
	hm.mu.RLock()
	initialTime := hm.peers["peer1"].LastSeen
	hm.mu.RUnlock()
	
	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)
	hm.UpdateLastSeen("peer1")
	
	// Check that last seen was updated
	hm.mu.RLock()
	updatedTime := hm.peers["peer1"].LastSeen
	hm.mu.RUnlock()
	
	if !updatedTime.After(initialTime) {
		t.Error("expected last seen time to be updated")
	}
}

func TestProcessHeartbeat(t *testing.T) {
	cfg := config.Default()
	hm := NewHeartbeatManager(cfg, "test-peer", nil)
	
	// Add a peer
	hm.AddPeer("peer1", "localhost:8080", nil)
	
	// Process a heartbeat
	heartbeat := message.NewHeartbeatMessage("peer1", 1)
	hm.ProcessHeartbeat(heartbeat)
	
	// Check that heartbeat was recorded
	_, received, active := hm.GetPeerStats()
	if received != 1 {
		t.Errorf("expected 1 heartbeat received, got %d", received)
	}
	if active != 1 {
		t.Errorf("expected 1 active peer, got %d", active)
	}
	
	// Process heartbeat from unknown peer (should not crash)
	unknownHeartbeat := message.NewHeartbeatMessage("unknown", 1)
	hm.ProcessHeartbeat(unknownHeartbeat)
	
	// Stats should not change
	_, received, active = hm.GetPeerStats()
	if received != 1 {
		t.Errorf("expected 1 heartbeat received after unknown peer, got %d", received)
	}
}

func TestHeartbeatTimeout(t *testing.T) {
	// Use short timeout for testing
	cfg := config.Default()
	cfg.HeartbeatTimeout = config.JSONDuration(50 * time.Millisecond)
	cfg.HeartbeatInterval = config.JSONDuration(20 * time.Millisecond)
	
	deadPeersCh := make(chan string, 10)
	onPeerDead := func(peerID string) {
		deadPeersCh <- peerID
	}
	
	hm := NewHeartbeatManager(cfg, "test-peer", onPeerDead)
	
	// Add a peer
	hm.AddPeer("peer1", "localhost:8080", nil)
	
	// Start monitoring
	hm.Start()
	defer hm.Stop()
	
	// Wait for timeout to trigger
	select {
	case deadPeer := <-deadPeersCh:
		if deadPeer != "peer1" {
			t.Errorf("expected dead peer 'peer1', got %s", deadPeer)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("expected peer to be marked as dead within timeout")
	}
	
	// Verify peer was removed
	if hm.GetPeerCount() != 0 {
		t.Errorf("expected 0 peers after timeout, got %d", hm.GetPeerCount())
	}
}

func TestHeartbeatKeepsAlive(t *testing.T) {
	// Use short timeout for testing
	cfg := config.Default()
	cfg.HeartbeatTimeout = config.JSONDuration(100 * time.Millisecond)
	cfg.HeartbeatInterval = config.JSONDuration(30 * time.Millisecond)
	
	deadPeersCh := make(chan string, 10)
	onPeerDead := func(peerID string) {
		deadPeersCh <- peerID
	}
	
	hm := NewHeartbeatManager(cfg, "test-peer", onPeerDead)
	
	// Add a peer
	hm.AddPeer("peer1", "localhost:8080", nil)
	
	// Start monitoring
	hm.Start()
	defer hm.Stop()
	
	// Send periodic heartbeats to keep peer alive
	go func() {
		ticker := time.NewTicker(40 * time.Millisecond)
		defer ticker.Stop()
		
		for i := 0; i < 5; i++ {
			select {
			case <-ticker.C:
				heartbeat := message.NewHeartbeatMessage("peer1", i+1)
				hm.ProcessHeartbeat(heartbeat)
			}
		}
	}()
	
	// Wait and verify peer is not marked as dead
	select {
	case deadPeer := <-deadPeersCh:
		t.Errorf("peer %s should not be marked as dead when sending heartbeats", deadPeer)
	case <-time.After(250 * time.Millisecond):
		// Expected - peer should stay alive
	}
	
	// Verify peer is still there
	if hm.GetPeerCount() != 1 {
		t.Errorf("expected 1 peer to remain alive, got %d", hm.GetPeerCount())
	}
}

func TestStartStop(t *testing.T) {
	cfg := config.Default()
	hm := NewHeartbeatManager(cfg, "test-peer", nil)
	
	// Start should not block
	hm.Start()
	
	// Stop should not block
	hm.Stop()
	
	// Multiple stops should not panic
	hm.Stop()
}

func TestGetPeerListWithMultiplePeers(t *testing.T) {
	cfg := config.Default()
	hm := NewHeartbeatManager(cfg, "test-peer", nil)
	
	// Add multiple peers
	hm.AddPeer("peer1", "localhost:8080", nil)
	hm.AddPeer("peer2", "localhost:8081", nil)
	hm.AddPeer("peer3", "", nil) // Peer without address
	
	peerList := hm.GetPeerList()
	
	// Should only include peers with addresses
	if len(peerList) != 2 {
		t.Errorf("expected 2 peers with addresses, got %d", len(peerList))
	}
	
	// Check that both addresses are present
	found8080 := false
	found8081 := false
	for _, addr := range peerList {
		switch addr {
		case "localhost:8080":
			found8080 = true
		case "localhost:8081":
			found8081 = true
		}
	}
	
	if !found8080 || !found8081 {
		t.Errorf("expected to find both peer addresses, got %v", peerList)
	}
}