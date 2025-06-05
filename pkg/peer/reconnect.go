package peer

import (
	"context"
	"math"
	"sync"
	"time"

	"example.com/p2p/pkg/config"
)

// ReconnectManager handles automatic reconnection to peers with exponential backoff
type ReconnectManager struct {
	config     *config.Config
	peer       *Peer
	heartbeat  *HeartbeatManager
	logger     interface{} // Logger interface to avoid circular dependency
	
	mu         sync.RWMutex
	reconnects map[string]*reconnectState
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// reconnectState tracks the reconnection state for a specific peer
type reconnectState struct {
	addr        string
	attempts    int
	lastAttempt time.Time
	backoff     time.Duration
	maxBackoff  time.Duration
	active      bool
}

// NewReconnectManager creates a new reconnection manager
func NewReconnectManager(cfg *config.Config, peer *Peer, heartbeat *HeartbeatManager) *ReconnectManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ReconnectManager{
		config:     cfg,
		peer:       peer,
		heartbeat:  heartbeat,
		reconnects: make(map[string]*reconnectState),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins the reconnection monitoring
func (rm *ReconnectManager) Start() {
	rm.wg.Add(1)
	go rm.reconnectionLoop()
}

// Stop stops the reconnection monitoring
func (rm *ReconnectManager) Stop() {
	rm.cancel()
	rm.wg.Wait()
}

// AddPeer adds a peer address for automatic reconnection
func (rm *ReconnectManager) AddPeer(addr string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if _, exists := rm.reconnects[addr]; !exists {
		rm.reconnects[addr] = &reconnectState{
			addr:       addr,
			attempts:   0,
			backoff:    1 * time.Second, // Start with 1 second
			maxBackoff: 5 * time.Minute, // Max 5 minutes
			active:     false,
		}
	}
}

// RemovePeer removes a peer from reconnection tracking
func (rm *ReconnectManager) RemovePeer(addr string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.reconnects, addr)
}

// TriggerReconnect marks a peer for immediate reconnection
func (rm *ReconnectManager) TriggerReconnect(addr string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if state, exists := rm.reconnects[addr]; exists {
		state.active = true
		state.lastAttempt = time.Time{} // Reset to trigger immediate attempt
	}
}

// reconnectionLoop periodically checks for peers that need reconnection
func (rm *ReconnectManager) reconnectionLoop() {
	defer rm.wg.Done()
	
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.checkReconnections()
		}
	}
}

// checkReconnections examines all peers and attempts reconnection if needed
func (rm *ReconnectManager) checkReconnections() {
	rm.mu.Lock()
	peerAddrs := rm.config.Peers
	reconnectStates := make(map[string]*reconnectState)
	for addr, state := range rm.reconnects {
		reconnectStates[addr] = state
	}
	rm.mu.Unlock()
	
	// Check if any configured peers are disconnected
	for _, addr := range peerAddrs {
		if rm.shouldReconnect(addr) {
			go rm.attemptReconnection(addr)
		}
	}
}

// shouldReconnect determines if we should attempt to reconnect to a peer
func (rm *ReconnectManager) shouldReconnect(addr string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Check if peer is currently connected
	if rm.isConnected(addr) {
		return false
	}
	
	state, exists := rm.reconnects[addr]
	if !exists {
		// New peer, add it and attempt connection
		rm.mu.RUnlock()
		rm.AddPeer(addr)
		rm.mu.RLock()
		return true
	}
	
	// Don't reconnect if already attempting
	if state.active {
		return false
	}
	
	// Check if enough time has passed since last attempt
	now := time.Now()
	if now.Sub(state.lastAttempt) < state.backoff {
		return false
	}
	
	return true
}

// isConnected checks if we're currently connected to a peer address
func (rm *ReconnectManager) isConnected(addr string) bool {
	// Check if any of our current peers match this address
	peerList := rm.heartbeat.GetPeerList()
	for _, peerAddr := range peerList {
		if peerAddr == addr {
			return true
		}
	}
	return false
}

// attemptReconnection tries to reconnect to a specific peer
func (rm *ReconnectManager) attemptReconnection(addr string) {
	rm.mu.Lock()
	state, exists := rm.reconnects[addr]
	if !exists {
		rm.mu.Unlock()
		return
	}
	
	// Mark as active to prevent concurrent attempts
	state.active = true
	state.lastAttempt = time.Now()
	state.attempts++
	rm.mu.Unlock()
	
	// Attempt connection
	remoteID, err := rm.peer.Connect(addr)
	
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if err != nil {
		// Connection failed, increase backoff
		state.backoff = time.Duration(float64(state.backoff) * 1.5)
		if state.backoff > state.maxBackoff {
			state.backoff = state.maxBackoff
		}
		state.active = false
		
		// Log the failure (if logger is available)
		if rm.logger != nil {
			// In a real implementation, we'd call the logger here
		}
	} else {
		// Connection successful, reset backoff and mark as inactive
		state.backoff = 1 * time.Second
		state.attempts = 0
		state.active = false
		
		// Add to heartbeat monitoring
		rm.heartbeat.AddPeer(remoteID, addr, nil)
		
		// Log the success (if logger is available)
		if rm.logger != nil {
			// In a real implementation, we'd call the logger here
		}
	}
}

// GetReconnectionStats returns statistics about reconnection attempts
func (rm *ReconnectManager) GetReconnectionStats() map[string]ReconnectStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	stats := make(map[string]ReconnectStats)
	for addr, state := range rm.reconnects {
		stats[addr] = ReconnectStats{
			Address:     addr,
			Attempts:    state.attempts,
			LastAttempt: state.lastAttempt,
			NextAttempt: state.lastAttempt.Add(state.backoff),
			Backoff:     state.backoff,
			Active:      state.active,
		}
	}
	
	return stats
}

// ReconnectStats provides information about reconnection attempts for a peer
type ReconnectStats struct {
	Address     string
	Attempts    int
	LastAttempt time.Time
	NextAttempt time.Time
	Backoff     time.Duration
	Active      bool
}

// CalculateBackoff calculates exponential backoff with jitter
func CalculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	if attempt <= 0 {
		return baseDelay
	}
	
	// Exponential backoff: base * 2^attempt
	backoff := float64(baseDelay) * math.Pow(2, float64(attempt))
	
	// Add jitter (±25%)
	jitterFactor := (float64(time.Now().UnixNano()%100) - 50) / 200 // ±25%
	jitter := backoff * jitterFactor
	backoff += jitter
	
	// Cap at maximum
	if time.Duration(backoff) > maxDelay {
		backoff = float64(maxDelay)
	}
	
	return time.Duration(backoff)
}