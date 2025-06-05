package peer

import (
	"testing"
	"time"

	"example.com/p2p/pkg/config"
)

func TestNewReconnectManager(t *testing.T) {
	cfg := config.Default()
	peer := New("localhost:0")
	heartbeat := NewHeartbeatManager(cfg, peer.ID, nil)
	
	rm := NewReconnectManager(cfg, peer, heartbeat)
	
	if rm.config != cfg {
		t.Error("config not set correctly")
	}
	
	if rm.peer != peer {
		t.Error("peer not set correctly")
	}
	
	if rm.heartbeat != heartbeat {
		t.Error("heartbeat not set correctly")
	}
	
	// Clean up
	rm.Stop()
}

func TestReconnectAddRemovePeer(t *testing.T) {
	cfg := config.Default()
	peer := New("localhost:0")
	heartbeat := NewHeartbeatManager(cfg, peer.ID, nil)
	rm := NewReconnectManager(cfg, peer, heartbeat)
	defer rm.Stop()
	
	// Add peer
	rm.AddPeer("localhost:8080")
	
	stats := rm.GetReconnectionStats()
	if len(stats) != 1 {
		t.Errorf("expected 1 peer, got %d", len(stats))
	}
	
	if _, exists := stats["localhost:8080"]; !exists {
		t.Error("peer not found in stats")
	}
	
	// Remove peer
	rm.RemovePeer("localhost:8080")
	
	stats = rm.GetReconnectionStats()
	if len(stats) != 0 {
		t.Errorf("expected 0 peers after removal, got %d", len(stats))
	}
}

func TestTriggerReconnect(t *testing.T) {
	cfg := config.Default()
	peer := New("localhost:0")
	heartbeat := NewHeartbeatManager(cfg, peer.ID, nil)
	rm := NewReconnectManager(cfg, peer, heartbeat)
	defer rm.Stop()
	
	// Add peer
	rm.AddPeer("localhost:8080")
	
	// Trigger reconnect
	rm.TriggerReconnect("localhost:8080")
	
	stats := rm.GetReconnectionStats()
	peerStats := stats["localhost:8080"]
	
	if !peerStats.Active {
		t.Error("peer should be marked as active for reconnection")
	}
}

func TestCalculateBackoff(t *testing.T) {
	baseDelay := 1 * time.Second
	maxDelay := 1 * time.Minute
	
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, baseDelay},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
	}
	
	for _, tt := range tests {
		t.Run(string(rune(tt.attempt)), func(t *testing.T) {
			backoff := CalculateBackoff(tt.attempt, baseDelay, maxDelay)
			
			// Allow for jitter (±25%)
			minExpected := float64(tt.expected) * 0.75
			maxExpected := float64(tt.expected) * 1.25
			
			if float64(backoff) < minExpected || float64(backoff) > maxExpected {
				t.Errorf("backoff %v outside expected range [%v, %v]", 
					backoff, time.Duration(minExpected), time.Duration(maxExpected))
			}
		})
	}
}

func TestCalculateBackoffMaximum(t *testing.T) {
	baseDelay := 1 * time.Second
	maxDelay := 10 * time.Second
	
	// High attempt count should be capped at maxDelay
	backoff := CalculateBackoff(10, baseDelay, maxDelay)
	
	if backoff > maxDelay {
		t.Errorf("backoff %v exceeds maximum %v", backoff, maxDelay)
	}
}

func TestReconnectStats(t *testing.T) {
	cfg := config.Default()
	peer := New("localhost:0")
	heartbeat := NewHeartbeatManager(cfg, peer.ID, nil)
	rm := NewReconnectManager(cfg, peer, heartbeat)
	defer rm.Stop()
	
	// Add peer
	rm.AddPeer("localhost:8080")
	
	stats := rm.GetReconnectionStats()
	peerStats, exists := stats["localhost:8080"]
	
	if !exists {
		t.Fatal("peer stats not found")
	}
	
	if peerStats.Address != "localhost:8080" {
		t.Errorf("expected address localhost:8080, got %s", peerStats.Address)
	}
	
	if peerStats.Attempts != 0 {
		t.Errorf("expected 0 attempts initially, got %d", peerStats.Attempts)
	}
	
	if peerStats.Active {
		t.Error("peer should not be active initially")
	}
}

func TestReconnectStartStop(t *testing.T) {
	cfg := config.Default()
	peer := New("localhost:0")
	heartbeat := NewHeartbeatManager(cfg, peer.ID, nil)
	rm := NewReconnectManager(cfg, peer, heartbeat)
	
	// Start should not block
	rm.Start()
	
	// Stop should not block
	rm.Stop()
	
	// Multiple stops should not panic
	rm.Stop()
}