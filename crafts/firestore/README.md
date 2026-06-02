# Firestore

A simplified implementation of Firestore's real-time database internals.

The goal is to understand **how Firestore's real-time sync actually works** — which is fundamentally different from a plain WebSocket pub/sub.

## Status

🚧 In progress

## What makes Firestore real-time different from WebSocket

| | WebSocket | Firestore (gRPC Listen stream) |
|---|---|---|
| Protocol | WebSocket (HTTP upgrade) | gRPC bidirectional streaming (HTTP/2) |
| Client subscription | arbitrary events | typed `AddTarget` / `RemoveTarget` requests |
| Resume on reconnect | manual, app-level | built-in **resume tokens** |
| Consistency | none | **snapshot markers** guarantee ordering |
| Delivery unit | raw message | `DocumentChange` + `TargetChange` events |

## Core concepts

### Listen RPC
The client opens a **long-lived gRPC stream** and sends `ListenRequest` messages:
- `AddTarget` — start watching a document or query
- `RemoveTarget` — stop watching

The server pushes `ListenResponse` messages:
- `DocumentChange` — a document was added/modified/deleted
- `TargetChange` — watch state changed (ADDED, CURRENT, RESET, REMOVED)
- `filter` — used to verify local cache consistency

### Resume token
Each `TargetChange(CURRENT)` carries a **resume token** (opaque bytes, internally a snapshot version).  
On reconnect, the client sends the token so the server replays only missed changes — no full re-fetch.

### Snapshot version
Every write is stamped with a **monotonically increasing version** (Timestamp in real Firestore).  
The server uses this to determine which changes a reconnecting client missed.

## Architecture (simplified)

```
Client ──AddTarget(path)──▶ ListenStream ──subscribe──▶ Watcher
                                  ▲                         │
Client ──PUT /doc──▶ Store ───change event─────────────────▶│
                                  │                         │
                         DocumentChange push ◀──────────────┘
                                  │
                               Client
```

1. **Store** — in-memory document store with snapshot versioning
2. **Watcher** — maps watch targets to subscriber streams; delivers ordered `DocumentChange` events
3. **ListenStream** — long-lived connection handling `AddTarget` / `RemoveTarget` and emitting responses
4. **Server** — HTTP for CRUD writes, custom stream handler for Listen

## Usage

```bash
go run ./go/main.go
```

```bash
# Write a document (triggers DocumentChange to all watchers)
curl -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","age":30}'

# Open a Listen stream and add a watch target
# (gRPC-like, simplified over HTTP/2 or SSE in this implementation)
```
