# P2P Text Chat

What you’ll learn:

- Fundamentals of socket programming (TCP / UDP)
- Peer discovery (simple bootstrap or multicast)
- Message routing and event loops

## Overview

Implement a console-based chat where any node can join without a central server. Each peer maintains connections to a small number of known peers and gossips new peers’ addresses.

## Implementation steps:

Peer identity: assign each instance a unique ID (UUID).

Bootstrap node (optional): hardcode one IP:port as “introducer.” On startup, a new peer connects to the introducer to fetch a small list of live peers.

Connection management: maintain a list of active peer endpoints (IP:port). On incoming connection, add to your peer list; on outgoing, add reciprocally.

Message format: define a simple struct like {senderID, sequenceNo, payload}; serialize via JSON or Protobuf.

Broadcasting: when someone types a line, wrap it in your message format and forward to all connected peers. Each recipient forwards to their peers unless they’ve seen that message sequenceNo before (use a small LRU cache).

Error handling: detect dead peers by failed writes or periodic heartbeats. Remove them from your list.

## Suggested stack:

Language: Go (using net + channels).

Serialization: JSON for simplicity; switch to Protobuf once core works.