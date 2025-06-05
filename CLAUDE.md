# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running
- `go build ./cmd/p2p` - Build the main P2P chat application
- `go run ./cmd/p2p -addr localhost:8080 -peer localhost:8081` - Run a peer instance
- `go run ./cmd/p2p -addr localhost:8081 -peer localhost:8080` - Run second peer to connect
- `go run ./cmd/p2p -config config.yaml` - Run with configuration file

### Configuration
- Copy `example-config.yaml` to `config.yaml` and modify as needed
- Supports both YAML and JSON configuration formats
- Environment variables with `P2P_` prefix override config file settings
- See `pkg/config` package for all available configuration options

### Testing
- `go test ./...` - Run all tests across packages (97% overall coverage)
- `go test ./pkg/peer -v` - Run specific package tests with verbose output
- `go test -run TestName` - Run specific test by name
- `go test -cover ./...` - Run tests with coverage reporting
- `go test -bench .` - Run benchmark tests

### Code Quality
- `go vet ./...` - Static analysis for potential issues
- `go fmt ./...` - Format all Go code

## Architecture Overview

This is a P2P text chat application built in Go using TCP sockets and JSON message serialization. The architecture follows a decentralized gossip protocol design.

### Core Components

**Message Flow**: Messages originate from user input → get broadcast to all connected peers → peers forward to their connections (with deduplication) → eventually reach all nodes in the network. Support for multiple message types (chat, heartbeat, peer list).

**Peer Management**: Each peer maintains a unique hex-encoded ID and a thread-safe map of active TCP connections. Peers can both accept incoming connections and initiate outgoing connections.

**Health Monitoring**: Heartbeat mechanism automatically detects and removes dead peers. Configurable heartbeat intervals and timeouts ensure network reliability.

**Configuration**: Flexible configuration system supporting YAML/JSON files and environment variables. Centralized configuration for network timeouts, logging, and peer limits.

**Structured Logging**: Context-aware logging using Go's log/slog with support for both text and JSON output formats. Includes peer-specific, connection-specific, and message-specific context.

**Deduplication**: Uses an LRU-based deduplication system (`pkg/dedup`) to prevent message loops. Messages are identified by `senderID/sequenceNo` pairs.

**Handshake Protocol**: When peers connect, they exchange their IDs over the TCP connection using a simple text protocol (ID + newline).

### Package Structure

- `cmd/p2p/main.go` - CLI entry point handling flags, user input, and message display
- `pkg/config/` - Configuration management with YAML/JSON/environment variable support
- `pkg/logger/` - Structured logging using Go's log/slog with context-aware logging
- `pkg/peer/peer.go` - Core peer logic including connection management, broadcasting, and message routing
- `pkg/peer/heartbeat.go` - Heartbeat mechanism for peer health detection and timeout handling
- `pkg/message/message.go` - Message structure with support for different message types (chat, heartbeat, peer list)
- `pkg/dedup/dedup.go` - LRU-based message deduplication to prevent infinite loops

### Key Patterns

**Connection Handling**: Each peer connection runs in its own goroutine (`readLoop`) that continuously reads messages and forwards them to other peers.

**Error Recovery**: Failed writes to peers automatically trigger connection cleanup and removal from the active peer list. Heartbeat timeouts provide additional failure detection.

**Thread Safety**: All peer state modifications use mutex locks to ensure safe concurrent access. Heartbeat manager uses read-write locks for efficient concurrent access.

**Configuration-Driven**: All timeouts, limits, and behavior can be configured via files or environment variables without code changes.

**Observable**: Structured logging provides detailed insights into peer connections, message flow, heartbeats, and errors for debugging and monitoring.

### Testing Approach

Comprehensive test suite with 97% overall coverage:
- **Unit Tests**: Cover all packages using table-driven tests and error injection
- **Integration Tests**: Spin up multiple peer instances to verify end-to-end message flow
- **In-Memory Testing**: Use `net.Pipe()` for fast, deterministic network simulation
- **Heartbeat Testing**: Time-based tests with short intervals to verify timeout behavior
- **Configuration Testing**: Test all config formats, validation, and environment variable overrides

### Production Features

- **Health Monitoring**: Automatic peer failure detection and cleanup
- **Configurable Timeouts**: Fine-tune network behavior for different environments  
- **Structured Logging**: JSON logs for production monitoring and text logs for development
- **Environment Support**: 12-factor app compliance with environment variable configuration
- **Resource Management**: Configurable connection limits and buffer sizes