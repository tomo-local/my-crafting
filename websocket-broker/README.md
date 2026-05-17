# WebSocket Broker

Building a WebSocket broker from scratch to understand the WebSocket upgrade handshake and real-time messaging.

## Implementations

- [Go](./go)

## Goals

- HTTP → WebSocket upgrade handshake (RFC 6455)
- Frame parsing and construction
- Pub/Sub style message routing between clients
- Heartbeat / ping-pong mechanism
