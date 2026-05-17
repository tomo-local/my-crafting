# Distributed Cache with Consistent Hashing

Building a distributed cache using consistent hashing to understand how data is partitioned across nodes.

## Implementations

- [Go](./go)

## Goals

- Consistent hashing ring with virtual nodes
- Cache node addition / removal with minimal key remapping
- Replication factor for fault tolerance
- LRU eviction policy per node
