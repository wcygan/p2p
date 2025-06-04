package peer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"

	"example.com/p2p/pkg/dedup"
	"example.com/p2p/pkg/message"
	"sync"
)

// Peer represents a node in the chat network.
type Peer struct {
	ID   string
	Addr string

	mu    sync.Mutex
	conns map[string]net.Conn
	seen  *dedup.Deduper
	// Messages delivers incoming messages from other peers.
	Messages chan *message.Message
}

// New creates a new peer listening on the given address.
func New(addr string) *Peer {
	return &Peer{
		ID:       randomID(),
		Addr:     addr,
		conns:    make(map[string]net.Conn),
		seen:     dedup.New(100),
		Messages: make(chan *message.Message, 16),
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

// Broadcast sends the given message to all connected peers.
func (p *Peer) Broadcast(msg *message.Message) error {
	data, err := msg.Marshal()
	if err != nil {
		return err
	}

	// mark our own message as seen to prevent rebroadcast loops
	p.seen.Seen(fmt.Sprintf("%s/%d", msg.SenderID, msg.SequenceNo))

	p.mu.Lock()
	defer p.mu.Unlock()
	for id, c := range p.conns {
		if _, err := c.Write(data); err != nil {
			return fmt.Errorf("write to %s: %w", id, err)
		}
	}
	return nil
}

// Seen reports whether the message has been encountered before and records it
// if it has not.
func (p *Peer) Seen(msg *message.Message) bool {
	return p.seen.Seen(fmt.Sprintf("%s/%d", msg.SenderID, msg.SequenceNo))
}

// HandleConn registers the connection and starts processing incoming messages.
func (p *Peer) HandleConn(id string, conn net.Conn) {
	p.AddConn(id, conn)
	go p.readLoop(id, conn)
}

func (p *Peer) readLoop(id string, conn net.Conn) {
	defer p.RemoveConn(id)
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		msg, err := message.Unmarshal(buf[:n])
		if err != nil {
			continue
		}
		if p.Seen(msg) {
			continue
		}
		select {
		case p.Messages <- msg:
		default:
		}
		_ = p.Broadcast(msg)
	}
}
