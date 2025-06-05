# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running
- `go build ./cmd/p2p` - Build the main P2P chat application
- `go run ./cmd/p2p -addr localhost:8080 -peer localhost:8081` - Run a peer instance
- `go run ./cmd/p2p -addr localhost:8081 -peer localhost:8080` - Run second peer to connect

### Testing
- `go test ./...` - Run all tests across packages
- `go test ./pkg/peer -v` - Run specific package tests with verbose output
- `go test -run TestName` - Run specific test by name

### Code Quality
- `go vet ./...` - Static analysis for potential issues
- `go fmt ./...` - Format all Go code

## Architecture Overview

This is a P2P text chat application built in Go using TCP sockets and JSON message serialization. The architecture follows a decentralized gossip protocol design.

### Core Components

**Message Flow**: Messages originate from user input → get broadcast to all connected peers → peers forward to their connections (with deduplication) → eventually reach all nodes in the network.

**Peer Management**: Each peer maintains a unique hex-encoded ID and a thread-safe map of active TCP connections. Peers can both accept incoming connections and initiate outgoing connections.

**Deduplication**: Uses an LRU-based deduplication system (`pkg/dedup`) to prevent message loops. Messages are identified by `senderID/sequenceNo` pairs.

**Handshake Protocol**: When peers connect, they exchange their IDs over the TCP connection using a simple text protocol (ID + newline).

### Package Structure

- `cmd/p2p/main.go` - CLI entry point handling flags, user input, and message display
- `pkg/peer/peer.go` - Core peer logic including connection management, broadcasting, and message routing
- `pkg/message/message.go` - Message structure and JSON serialization/deserialization
- `pkg/dedup/dedup.go` - LRU-based message deduplication to prevent infinite loops

### Key Patterns

**Connection Handling**: Each peer connection runs in its own goroutine (`readLoop`) that continuously reads messages and forwards them to other peers.

**Error Recovery**: Failed writes to peers automatically trigger connection cleanup and removal from the active peer list.

**Thread Safety**: All peer state modifications use mutex locks to ensure safe concurrent access.

### Testing Approach

Tests use `net.Pipe()` for in-memory connections to simulate peer interactions without actual network I/O. Integration tests spin up multiple peer instances to verify end-to-end message flow.