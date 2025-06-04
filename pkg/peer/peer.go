package peer

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"sync"
)

// Peer represents a node in the chat network.
type Peer struct {
	ID   string
	Addr string

	mu    sync.Mutex
	conns map[string]net.Conn
}

// New creates a new peer listening on the given address.
func New(addr string) *Peer {
	return &Peer{
		ID:    randomID(),
		Addr:  addr,
		conns: make(map[string]net.Conn),
	}
}

func randomID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// AddConn adds a connection to another peer.
func (p *Peer) AddConn(id string, conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conns[id] = conn
}

// RemoveConn removes a connection.
func (p *Peer) RemoveConn(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if c, ok := p.conns[id]; ok {
		c.Close()
		delete(p.conns, id)
	}
}

// Connections returns the number of active connections.
func (p *Peer) Connections() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.conns)
}
