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
| [HTTP Server](./crafts/http-server/go) | Go | ✅ |
| [HTTP Server](./crafts/http-server/ruby) | Ruby | ✅ |
| [Reverse Proxy / Load Balancer](./crafts/reverse-proxy) | Go | 🚧 |
| [Pub/Sub Broker](./crafts/pub-sub) | Go | 🚧 |
| [Rate Limiter](./crafts/rate-limiter) | Go | Planned |
| [URL Shortener](./crafts/url-shortener) | Go | Planned |
| [WebSocket Broker](./crafts/websocket-broker) | Go | Planned |
| [Redis Clone](./crafts/redis-clone) | Go | Planned |
| [CRON Scheduler](./crafts/cron-scheduler) | Go | Planned |
| [Distributed Cache](./crafts/distributed-cache) | Go | Planned |
| [Search Engine](./crafts/search-engine) | Go | Planned |
| [Kafka Clone](./crafts/kafka-clone) | Go | Planned |
| [Game Boy Emulator](./crafts/gameboy-emulator) | Go | Planned |

## Lab

### Languages

| Language | Repo | Status |
|----------|------|--------|
| [Haskell](./lab/lang/haskell) | [study-haskell](https://github.com/tomo-local/study-haskell) | 🚧 |
| [C](./lab/lang/c) | [study-C](https://github.com/tomo-local/study-C) | 🚧 |
| [Ruby](./lab/lang/ruby) | — | 🚧 |

### Libraries

| Library / Runtime | Repo | Status |
|-------------------|------|--------|
| [htmx](./lab/lib/htmx) | [study-htmx](https://github.com/tomo-local/study-htmx) | 🚧 |
| [Deno](./lab/lib/deno) | [study-deno](https://github.com/tomo-local/study-deno) | 🚧 |
| [gRPC](./lab/lib/grpc) | [study-gRPC](https://github.com/tomo-local/study-gRPC) | 🚧 |
| [XState](./lab/lib/xstate) | [study-xstate](https://github.com/tomo-local/study-xstate) | 🚧 |
| [Plasmo](./lab/lib/plasmo) | [study-plasmo](https://github.com/tomo-local/study-plasmo) | 🚧 |
| [WXT](./lab/lib/wxt) | [study-wxt](https://github.com/tomo-local/study-wxt) | 🚧 |
| [Jotai](./lab/lib/jotai) | [study-jotai](https://github.com/tomo-local/study-jotai) | 🚧 |

### Infrastructure

| Topic | Status |
|-------|--------|
| [Kubernetes Build-Break-Fix](./lab/infra/kubernetes/build-break-fix) | 🚧 |

## Structure

```
my-crafting/
├── crafts/
│   ├── http-server/
│   ├── reverse-proxy/
│   ├── pub-sub/
│   ├── rate-limiter/
│   ├── url-shortener/
│   ├── websocket-broker/
│   ├── redis-clone/
│   ├── cron-scheduler/
│   ├── distributed-cache/
│   ├── search-engine/
│   └── kafka-clone/
├── lab/
│   ├── lang/
│   │   ├── haskell/
│   │   ├── c/
│   │   └── ruby/
│   ├── lib/
│   │   ├── htmx/
│   │   ├── deno/
│   │   ├── grpc/
│   │   ├── xstate/
│   │   ├── plasmo/
│   │   ├── wxt/
│   │   └── jotai/
│   └── infra/
│       └── kubernetes/
└── docs/
    └── README.ja.md
```

## Philosophy

- Build it to understand it
- Re-implement in different languages to see how the same concept translates
- Prefer simplicity over completeness
