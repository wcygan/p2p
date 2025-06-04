package peer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

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

// handshake exchanges peer IDs on the given connection. The caller's ID is
// sent first, then the remote ID is read back. It returns the remote ID or an
// error.
func handshake(conn net.Conn, id string) (string, error) {
	if _, err := fmt.Fprintf(conn, "%s\n", id); err != nil {
		return "", err
	}
	var buf []byte
	tmp := make([]byte, 1)
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			return "", err
		}
		if n == 0 {
			continue
		}
		if tmp[0] == '\n' {
			break
		}
		buf = append(buf, tmp[0])
		if len(buf) > 64 {
			return "", fmt.Errorf("id too long")
		}
	}
	return strings.TrimSpace(string(buf)), nil
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

// Connect dials the given address, performs a handshake, and registers the
// connection. It returns the remote peer ID.
func (p *Peer) Connect(addr string) (string, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}
	remoteID, err := handshake(conn, p.ID)
	if err != nil {
		conn.Close()
		return "", err
	}
	p.HandleConn(remoteID, conn)
	return remoteID, nil
}

// Serve accepts incoming connections from the listener and handles them. It
// runs until the listener returns a non-nil error.
func (p *Peer) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		remoteID, err := handshake(conn, p.ID)
		if err != nil {
			conn.Close()
			continue
		}
		p.HandleConn(remoteID, conn)
	}
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
	conns := make(map[string]net.Conn, len(p.conns))
	for id, c := range p.conns {
		conns[id] = c
	}
	p.mu.Unlock()

	var firstErr error
	for id, c := range conns {
		if _, err := c.Write(data); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("write to %s: %w", id, err)
			}
			p.RemoveConn(id)
			continue
		}
	}
	return firstErr
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
