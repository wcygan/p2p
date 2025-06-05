package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"example.com/p2p/pkg/config"
	"example.com/p2p/pkg/logger"
	"example.com/p2p/pkg/message"
	"example.com/p2p/pkg/peer"
)

const Version = "1.0.0"

type addrList []string

func (a *addrList) String() string     { return strings.Join(*a, ",") }
func (a *addrList) Set(s string) error { *a = append(*a, s); return nil }

// App represents the main P2P chat application
type App struct {
	config    *config.Config
	logger    *logger.Logger
	peer      *peer.Peer
	heartbeat *peer.HeartbeatManager
	listener  net.Listener
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewApp creates a new P2P chat application
func NewApp(cfg *config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())
	
	app := &App{
		config: cfg,
		logger: logger.New(cfg),
		ctx:    ctx,
		cancel: cancel,
	}
	
	return app
}

// Start initializes and starts the P2P chat application
func (app *App) Start() error {
	app.logger.LogServerStarted("initializing", app.config.ListenAddr)
	
	// Create peer
	app.peer = peer.NewWithConfig(app.config.ListenAddr, app.config)
	peerLogger := app.logger.WithPeer(app.peer.ID)
	
	// Create heartbeat manager
	app.heartbeat = peer.NewHeartbeatManager(app.config, app.peer.ID, app.onPeerDead)
	
	// Start listening
	ln, err := net.Listen("tcp", app.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", app.config.ListenAddr, err)
	}
	app.listener = ln
	
	actualAddr := ln.Addr().String()
	peerLogger.LogServerStarted(app.peer.ID, actualAddr)
	fmt.Printf("🚀 P2P Chat v%s\n", Version)
	fmt.Printf("📡 Peer ID: %s\n", app.peer.ID)
	fmt.Printf("🔗 Listening on: %s\n", actualAddr)
	fmt.Printf("💬 Type messages to chat, Ctrl+C to exit\n\n")
	
	// Start server goroutine
	app.wg.Add(1)
	go app.runServer()
	
	// Connect to initial peers
	app.connectToInitialPeers()
	
	// Start heartbeat monitoring
	app.heartbeat.Start()
	
	// Start message processing
	app.wg.Add(1)
	go app.processMessages()
	
	// Start user input processing
	app.wg.Add(1)
	go app.processUserInput()
	
	return nil
}

// Stop gracefully shuts down the application
func (app *App) Stop() {
	app.logger.Info("Shutting down P2P chat application")
	
	// Cancel context to signal all goroutines to stop
	app.cancel()
	
	// Stop heartbeat monitoring
	if app.heartbeat != nil {
		app.heartbeat.Stop()
	}
	
	// Close listener
	if app.listener != nil {
		app.listener.Close()
	}
	
	// Wait for all goroutines to finish
	app.wg.Wait()
	
	app.logger.LogServerStopped(app.peer.ID)
	fmt.Println("\n👋 P2P Chat stopped")
}

// Wait blocks until the application is stopped
func (app *App) Wait() {
	app.wg.Wait()
}

// runServer handles incoming connections
func (app *App) runServer() {
	defer app.wg.Done()
	
	for {
		select {
		case <-app.ctx.Done():
			return
		default:
			// Set accept timeout to check for context cancellation
			if tcpListener, ok := app.listener.(*net.TCPListener); ok {
				tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
			}
			
			conn, err := app.listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Check context and try again
				}
				if !strings.Contains(err.Error(), "use of closed network connection") {
					app.logger.LogConnectionError("accept", err)
				}
				return
			}
			
			// Handle connection in background
			go app.handleIncomingConnection(conn)
		}
	}
}

// handleIncomingConnection processes a new incoming connection
func (app *App) handleIncomingConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	app.logger.Debug("Incoming connection", "remote_addr", remoteAddr)
	
	// Perform handshake
	remoteID, err := peer.Handshake(conn, app.peer.ID)
	if err != nil {
		app.logger.LogConnectionError(remoteAddr, fmt.Errorf("handshake failed: %w", err))
		conn.Close()
		return
	}
	
	// Check connection limits
	if app.peer.Connections() >= app.config.MaxConnections {
		app.logger.Warn("Connection limit reached, rejecting peer",
			"peer_id", remoteID,
			"limit", app.config.MaxConnections)
		conn.Close()
		return
	}
	
	// Register peer
	app.peer.HandleConn(remoteID, conn)
	app.heartbeat.AddPeer(remoteID, remoteAddr, conn)
	app.logger.LogPeerConnected(remoteID, remoteAddr)
}

// connectToInitialPeers connects to peers specified in configuration
func (app *App) connectToInitialPeers() {
	for _, peerAddr := range app.config.Peers {
		go app.connectToPeer(peerAddr)
	}
}

// connectToPeer establishes connection to a specific peer
func (app *App) connectToPeer(addr string) {
	app.logger.Debug("Attempting to connect to peer", "address", addr)
	
	remoteID, err := app.peer.Connect(addr)
	if err != nil {
		app.logger.LogConnectionError(addr, err)
		return
	}
	
	app.heartbeat.AddPeer(remoteID, addr, nil) // conn is managed by peer
	app.logger.LogPeerConnected(remoteID, addr)
}

// processMessages handles incoming messages from other peers
func (app *App) processMessages() {
	defer app.wg.Done()
	
	for {
		select {
		case <-app.ctx.Done():
			return
		case msg := <-app.peer.Messages:
			app.handleMessage(msg)
		}
	}
}

// handleMessage processes different types of incoming messages
func (app *App) handleMessage(msg *message.Message) {
	app.heartbeat.UpdateLastSeen(msg.SenderID)
	
	switch {
	case msg.IsHeartbeat():
		app.heartbeat.ProcessHeartbeat(msg)
		app.logger.LogHeartbeatReceived(msg.SenderID, msg.SequenceNo)
	case msg.IsChatMessage():
		app.logger.LogMessageReceived(string(msg.Type), msg.SenderID, msg.SequenceNo)
		fmt.Printf("💬 [%s]: %s\n", msg.SenderID[:8], msg.Payload)
	default:
		app.logger.Debug("Received unknown message type",
			"type", msg.Type,
			"sender_id", msg.SenderID)
	}
}

// processUserInput handles user keyboard input
func (app *App) processUserInput() {
	defer app.wg.Done()
	
	scanner := bufio.NewScanner(os.Stdin)
	sequenceNo := 1
	
	for {
		select {
		case <-app.ctx.Done():
			return
		default:
			// Set a timeout for Scan to check context periodically
			done := make(chan bool, 1)
			var text string
			
			go func() {
				if scanner.Scan() {
					text = scanner.Text()
					done <- true
				} else {
					done <- false
				}
			}()
			
			select {
			case <-app.ctx.Done():
				return
			case success := <-done:
				if !success {
					if err := scanner.Err(); err != nil {
						app.logger.Error("Error reading input", "error", err)
					}
					return
				}
				
				if text == "" {
					continue
				}
				
				// Create and broadcast chat message
				msg := message.NewChatMessage(app.peer.ID, sequenceNo, text)
				sequenceNo++
				
				if err := app.peer.Broadcast(msg); err != nil {
					app.logger.Error("Failed to broadcast message", "error", err)
					fmt.Printf("❌ Failed to send message: %v\n", err)
				} else {
					app.logger.LogMessageBroadcast(string(msg.Type), msg.SequenceNo, app.peer.Connections())
				}
			case <-time.After(100 * time.Millisecond):
				// Periodic check to allow context cancellation
				continue
			}
		}
	}
}

// onPeerDead is called when a peer is detected as dead
func (app *App) onPeerDead(peerID string) {
	app.peer.RemoveConn(peerID)
	app.logger.LogPeerTimedOut(peerID, "heartbeat timeout")
	fmt.Printf("⚠️  Peer %s disconnected (timeout)\n", peerID[:8])
}

func main() {
	// Command line flags
	var peers addrList
	configFile := flag.String("config", "", "path to configuration file")
	addr := flag.String("addr", "", "listen address (overrides config)")
	logLevel := flag.String("log-level", "", "log level (debug, info, warn, error)")
	version := flag.Bool("version", false, "show version and exit")
	flag.Var(&peers, "peer", "peer address to connect to (may be repeated)")
	flag.Parse()
	
	// Show version and exit
	if *version {
		fmt.Printf("P2P Chat v%s\n", Version)
		os.Exit(0)
	}
	
	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Configuration error: %v\n", err)
		os.Exit(1)
	}
	
	// Override config with command line flags
	if *addr != "" {
		cfg.ListenAddr = *addr
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}
	if len(peers) > 0 {
		cfg.Peers = append(cfg.Peers, []string(peers)...)
	}
	
	// Validate final configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid configuration: %v\n", err)
		os.Exit(1)
	}
	
	// Create and start application
	app := NewApp(cfg)
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Start application
	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to start: %v\n", err)
		os.Exit(1)
	}
	
	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\n🛑 Received shutdown signal...")
	
	// Graceful shutdown
	app.Stop()
}