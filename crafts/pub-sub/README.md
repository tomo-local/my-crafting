# Pub/Sub Broker

Building an in-memory pub/sub broker from scratch to understand topic-based message routing and fan-out delivery.

## Implementations

- [Go](./go)

## Goals

- Simple line-based TCP protocol (NATS-inspired: SUB / PUB / MSG)
- Topic-based message routing and fan-out to multiple subscribers
- Subscribe / Unsubscribe lifecycle management
- Slow subscriber handling (backpressure and drop strategies)
- Wildcard subscriptions (`news.*`, `events.>`)
- Last-N message buffer for new subscribers (replay on join)
