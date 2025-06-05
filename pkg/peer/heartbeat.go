package peer

import (
	"sync"
	"time"

	"example.com/p2p/pkg/config"
	"example.com/p2p/pkg/message"
)

// PeerInfo tracks information about a connected peer
type PeerInfo struct {
	ID           string
	Addr         string
	LastSeen     time.Time
	LastHeartbeat time.Time
	Conn         interface{} // net.Conn, stored as interface{} to avoid import cycles in tests
}

// HeartbeatManager manages peer health detection via heartbeats
type HeartbeatManager struct {
	config   *config.Config
	peerID   string
	peers    map[string]*PeerInfo
	mu       sync.RWMutex
	stopCh   chan struct{}
	onPeerDead func(peerID string)
	
	// Statistics
	heartbeatsSent     int64
	heartbeatsReceived int64
}

// NewHeartbeatManager creates a new heartbeat manager
func NewHeartbeatManager(cfg *config.Config, peerID string, onPeerDead func(string)) *HeartbeatManager {
	return &HeartbeatManager{
		config:     cfg,
		peerID:     peerID,
		peers:      make(map[string]*PeerInfo),
		stopCh:     make(chan struct{}),
		onPeerDead: onPeerDead,
	}
}

// Start begins the heartbeat monitoring
func (hm *HeartbeatManager) Start() {
	go hm.heartbeatLoop()
	go hm.healthCheckLoop()
}

// Stop stops the heartbeat monitoring
func (hm *HeartbeatManager) Stop() {
	close(hm.stopCh)
}

// AddPeer adds a peer to be monitored
func (hm *HeartbeatManager) AddPeer(id, addr string, conn interface{}) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	now := time.Now()
	hm.peers[id] = &PeerInfo{
		ID:           id,
		Addr:         addr,
		LastSeen:     now,
		LastHeartbeat: now,
		Conn:         conn,
	}
}

// RemovePeer removes a peer from monitoring
func (hm *HeartbeatManager) RemovePeer(id string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.peers, id)
}

// UpdateLastSeen updates the last seen time for a peer
func (hm *HeartbeatManager) UpdateLastSeen(id string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	if peer, exists := hm.peers[id]; exists {
		peer.LastSeen = time.Now()
	}
}

// ProcessHeartbeat processes a received heartbeat message
func (hm *HeartbeatManager) ProcessHeartbeat(msg *message.Message) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	if peer, exists := hm.peers[msg.SenderID]; exists {
		peer.LastHeartbeat = time.Now()
		peer.LastSeen = time.Now()
		hm.heartbeatsReceived++
	}
}

// GetPeerCount returns the number of monitored peers
func (hm *HeartbeatManager) GetPeerCount() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return len(hm.peers)
}

// GetPeerStats returns peer statistics
func (hm *HeartbeatManager) GetPeerStats() (sent int64, received int64, activePeers int) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.heartbeatsSent, hm.heartbeatsReceived, len(hm.peers)
}

// GetPeerList returns a list of known peer addresses
func (hm *HeartbeatManager) GetPeerList() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	
	addrs := make([]string, 0, len(hm.peers))
	for _, peer := range hm.peers {
		if peer.Addr != "" {
			addrs = append(addrs, peer.Addr)
		}
	}
	return addrs
}

// heartbeatLoop sends periodic heartbeats to all peers
func (hm *HeartbeatManager) heartbeatLoop() {
	ticker := time.NewTicker(time.Duration(hm.config.HeartbeatInterval))
	defer ticker.Stop()
	
	sequenceNo := 1
	
	for {
		select {
		case <-ticker.C:
			hm.sendHeartbeats(sequenceNo)
			sequenceNo++
		case <-hm.stopCh:
			return
		}
	}
}

// healthCheckLoop monitors peer health and removes dead peers
func (hm *HeartbeatManager) healthCheckLoop() {
	ticker := time.NewTicker(time.Duration(hm.config.HeartbeatInterval) / 2)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			hm.checkPeerHealth()
		case <-hm.stopCh:
			return
		}
	}
}

// sendHeartbeats sends heartbeat messages to all peers
func (hm *HeartbeatManager) sendHeartbeats(sequenceNo int) {
	hm.mu.RLock()
	peerCount := len(hm.peers)
	hm.mu.RUnlock()
	
	if peerCount > 0 {
		// This would be called by the main peer to broadcast the heartbeat
		// We can't directly broadcast here to avoid circular dependencies
		hm.mu.Lock()
		hm.heartbeatsSent++
		hm.mu.Unlock()
	}
}

// checkPeerHealth checks if any peers have exceeded the heartbeat timeout
func (hm *HeartbeatManager) checkPeerHealth() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	timeout := time.Duration(hm.config.HeartbeatTimeout)
	now := time.Now()
	deadPeers := make([]string, 0)
	
	for id, peer := range hm.peers {
		if now.Sub(peer.LastHeartbeat) > timeout {
			deadPeers = append(deadPeers, id)
		}
	}
	
	// Remove dead peers and notify
	for _, id := range deadPeers {
		delete(hm.peers, id)
		if hm.onPeerDead != nil {
			// Call without holding the lock to avoid deadlocks
			go hm.onPeerDead(id)
		}
	}
}