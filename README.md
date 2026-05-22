# my-crafting

A collection of low-level implementations built from scratch to deepen understanding of how things work under the hood.

Each topic is organized by directory and may include multiple language implementations.

> Japanese version: [docs/README.ja.md](docs/README.ja.md)

## Crafts

| Topic | Languages | Status |
|-------|-----------|--------|
| [typescript-mini-zod](./crafts/typescript-mini-zod) | TypeScript | ✅ |
| [react-mini-zustand](./crafts/react-mini-zustand) | TypeScript | ✅ |
| [react-mini-jotai](./crafts/react-mini-jotai) | TypeScript | ✅ |
| [react-mini-xstate](./crafts/react-mini-xstate) | TypeScript | ✅ |
| [react-mini-valtio](./crafts/react-mini-valtio) | TypeScript | ✅ |
| [HTTP Server](./crafts/http-server) | Go | ✅ |
| [Reverse Proxy / Load Balancer](./crafts/reverse-proxy) | Go | Planned |
| [Rate Limiter](./crafts/rate-limiter) | Go | Planned |
| [URL Shortener](./crafts/url-shortener) | Go | Planned |
| [Pub/Sub Broker](./crafts/pub-sub) | Go | Planned |
| [WebSocket Broker](./crafts/websocket-broker) | Go | Planned |
| [Redis Clone](./crafts/redis-clone) | Go | Planned |
| [CRON Scheduler](./crafts/cron-scheduler) | Go | Planned |
| [Distributed Cache](./crafts/distributed-cache) | Go | Planned |
| [Search Engine](./crafts/search-engine) | Go | Planned |
| [Kafka Clone](./crafts/kafka-clone) | Go | Planned |

## Languages

| Language | Repo | Status |
|----------|------|--------|
| [Haskell](./languages/haskell) | [study-haskell](https://github.com/tomo-local/study-haskell) | 🚧 |
| [C](./languages/c) | [study-C](https://github.com/tomo-local/study-C) | 🚧 |
| [Ruby](./languages/ruby) | — | 🚧 |

## Libraries

| Library / Runtime | Repo | Status |
|-------------------|------|--------|
| [htmx](./libraries/htmx) | [study-htmx](https://github.com/tomo-local/study-htmx) | 🚧 |
| [Deno](./libraries/deno) | [study-deno](https://github.com/tomo-local/study-deno) | 🚧 |
| [gRPC](./libraries/grpc) | [study-gRPC](https://github.com/tomo-local/study-gRPC) | 🚧 |
| [XState](./libraries/xstate) | [study-xstate](https://github.com/tomo-local/study-xstate) | 🚧 |
| [Plasmo](./libraries/plasmo) | [study-plasmo](https://github.com/tomo-local/study-plasmo) | 🚧 |
| [WXT](./libraries/wxt) | [study-wxt](https://github.com/tomo-local/study-wxt) | 🚧 |
| [Jotai](./libraries/jotai) | [study-jotai](https://github.com/tomo-local/study-jotai) | 🚧 |

## Structure

```
my-crafting/
├── crafts/
│   ├── http-server/
│   ├── reverse-proxy/
│   ├── rate-limiter/
│   ├── url-shortener/
│   ├── pub-sub/
│   ├── websocket-broker/
│   ├── redis-clone/
│   ├── cron-scheduler/
│   ├── distributed-cache/
│   ├── search-engine/
│   └── kafka-clone/
├── languages/
│   ├── haskell/
│   ├── c/
│   └── ruby/
├── libraries/
│   ├── htmx/
│   ├── deno/
│   ├── grpc/
│   ├── xstate/
│   ├── plasmo/
│   ├── wxt/
│   └── jotai/
└── docs/
    └── README.ja.md
```

## Philosophy

- Build it to understand it
- Re-implement in different languages to see how the same concept translates
- Prefer simplicity over completeness
