# my-crafting

A collection of low-level implementations built from scratch to deepen understanding of how things work under the hood.

Each topic is organized by directory and may include multiple language implementations.

> Japanese version: [docs/README.ja.md](docs/README.ja.md)

## Crafts

| Topic | Languages | Status |
|-------|-----------|--------|
| [HTTP Server](./crafts/http-server) | Go | ✅ |
| [typescript-mini-zod](./crafts/typescript-mini-zod) | TypeScript | ✅ |
| [react-mini-zustand](./crafts/react-mini-zustand) | TypeScript | ✅ |
| [react-mini-jotai](./crafts/react-mini-jotai) | TypeScript | ✅ |
| [react-mini-xstate](./crafts/react-mini-xstate) | TypeScript | ✅ |
| [react-mini-valtio](./crafts/react-mini-valtio) | TypeScript | ✅ |
| [Redis Clone](./crafts/redis-clone) | Go | Planned |
| [Kafka Clone](./crafts/kafka-clone) | Go | Planned |
| [Reverse Proxy / Load Balancer](./crafts/reverse-proxy) | Go | Planned |
| [Rate Limiter](./crafts/rate-limiter) | Go | Planned |
| [CRON Scheduler](./crafts/cron-scheduler) | Go | Planned |
| [WebSocket Broker](./crafts/websocket-broker) | Go | Planned |
| [Distributed Cache](./crafts/distributed-cache) | Go | Planned |
| [Search Engine](./crafts/search-engine) | Go | Planned |
| [URL Shortener](./crafts/url-shortener) | Go | Planned |

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
│   ├── redis-clone/
│   ├── kafka-clone/
│   ├── reverse-proxy/
│   ├── rate-limiter/
│   ├── cron-scheduler/
│   ├── websocket-broker/
│   ├── distributed-cache/
│   ├── search-engine/
│   └── url-shortener/
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
