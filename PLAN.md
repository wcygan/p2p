# Development Plan

This document outlines a strategy for implementing the P2P text chat described in `README.md`. The goal is to build a console-based chat application in Go where peers communicate directly without a central server.

## 1. Project Setup

1. **Initialize Go module** – run `go mod init` to manage dependencies.
2. **Directory layout** – separate packages for networking, message handling, peer management, and CLI.
3. **Version control** – keep incremental commits with tests.

## 2. Incremental Implementation

Break down features from `README.md` into small tasks that can be implemented and tested individually.

### 2.1 Peer Identity

- Use UUIDs for each peer instance.
- Store in a struct with network address.
- Write unit tests for UUID generation and parsing.

### 2.2 Bootstrap / Peer Discovery

- Start with a simple bootstrap list from config or environment variable.
- Create a function that retrieves peers from the bootstrap node.
- Mock network interactions in tests.

### 2.3 Connection Management

- Maintain active connections in a thread-safe map or struct.
- Implement add/remove operations with tests.
- Handle incoming and outgoing connections symmetrically.

### 2.4 Message Format and Routing

- Define `Message` struct `{senderID, sequenceNo, payload}` using JSON.
- Implement message serialization and deserialization.
- Add unit tests covering JSON marshaling/unmarshaling and deduplication logic.

### 2.5 Broadcasting Logic

- When a user types a message, broadcast to connected peers.
- Use a small LRU cache to ignore duplicates.
- Test broadcasting with in-memory connections (using `net.Pipe`).

### 2.6 Error Handling and Heartbeats

- Detect dead peers via failed writes or periodic pings.
- Provide a cleanup routine that removes stale peers.
- Add tests that simulate dead connections.

## 3. Testing Infrastructure

Follow Go best practices to ensure reliability:

1. **Unit tests** – for all packages using the `testing` package.
2. **Table-driven tests** – to cover edge cases systematically.
3. **Mocks / fakes** – use interfaces to allow mock network components.
4. **Integration tests** – spin up multiple peers to verify real message flow.
5. **Continuous testing** – run `go test ./...` frequently; integrate with CI if available.

Include helper scripts or Makefile targets for running tests and linting. Use `go vet` and `golangci-lint` for static analysis.

## 4. Future Enhancements

- Switch serialization to Protobuf once JSON version is stable.
- Add persistence for chat history.
- Explore NAT traversal or more advanced peer discovery techniques.

By progressing incrementally and validating each component with tests, we’ll build a maintainable P2P chat application with confidence.
