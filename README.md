# my-crafting

A collection of low-level implementations built from scratch to deepen understanding of how things work under the hood.

Each topic is organized by directory and may include multiple language implementations.

> Japanese version: [docs/README.ja.md](docs/README.ja.md)

## Topics

| Topic | Languages | Status |
|-------|-----------|--------|
| [HTTP Server](./crafts/http-server) | Go | ✅ |
| [typescript-mini-zod](./crafts/typescript-mini-zod) | TypeScript | Done |
| [react-mini-zustand](./crafts/react-mini-zustand) | TypeScript | Done |
| [react-mini-jotai](./crafts/react-mini-jotai) | TypeScript | Done |
| [react-mini-xstate](./crafts/react-mini-xstate) | TypeScript | Done |
| [react-mini-valtio](./crafts/react-mini-valtio) | TypeScript | Done |
| [Redis Clone](./crafts/redis-clone) | Go | Planned |
| [Kafka Clone](./crafts/kafka-clone) | Go | Planned |
| [Reverse Proxy / Load Balancer](./crafts/reverse-proxy) | Go | Planned |
| [Rate Limiter](./crafts/rate-limiter) | Go | Planned |
| [CRON Scheduler](./crafts/cron-scheduler) | Go | Planned |
| [WebSocket Broker](./crafts/websocket-broker) | Go | Planned |
| [Distributed Cache](./crafts/distributed-cache) | Go | Planned |
| [Search Engine](./crafts/search-engine) | Go | Planned |
| [URL Shortener](./crafts/url-shortener) | Go | Planned |

## Structure

```
my-crafting/
├── crafts/
│   ├── http-server/
│   ├── redis-clone/
│   ├── kafka-clone/
│   ├── reverse-proxy/
│   ├── rate-limiter/
│   ├── cron-scheduler/
│   ├── websocket-broker/
│   ├── distributed-cache/
│   ├── search-engine/
│   └── url-shortener/
└── docs/
    └── README.ja.md
```

## Philosophy

- Build it to understand it
- Re-implement in different languages to see how the same concept translates
- Prefer simplicity over completeness
