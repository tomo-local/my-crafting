# my-crafting

仕組みを深く理解するために、低レイヤーの実装をゼロから自作するリポジトリです。

トピックごとにディレクトリで管理し、同じ内容を複数言語で実装することもあります。

## クラフト一覧

| トピック | 言語 | 状態 |
|---------|------|------|
| [HTTP サーバー](../crafts/http-server) | Go | ✅ |
| [typescript-mini-zod](../crafts/typescript-mini-zod) | TypeScript | ✅ |
| [react-mini-zustand](../crafts/react-mini-zustand) | TypeScript | ✅ |
| [react-mini-jotai](../crafts/react-mini-jotai) | TypeScript | ✅ |
| [react-mini-xstate](../crafts/react-mini-xstate) | TypeScript | ✅ |
| [react-mini-valtio](../crafts/react-mini-valtio) | TypeScript | ✅ |
| [Redis クローン](../crafts/redis-clone) | Go | 予定 |
| [Kafka クローン（メッセージブローカー）](../crafts/kafka-clone) | Go | 予定 |
| [リバースプロキシ / ロードバランサー](../crafts/reverse-proxy) | Go | 予定 |
| [レートリミッター](../crafts/rate-limiter) | Go | 予定 |
| [CRON タスクスケジューラー](../crafts/cron-scheduler) | Go | 予定 |
| [WebSocket ブローカー](../crafts/websocket-broker) | Go | 予定 |
| [分散キャッシュ（一貫性ハッシュ）](../crafts/distributed-cache) | Go | 予定 |
| [ミニ検索エンジン](../crafts/search-engine) | Go | 予定 |
| [URL 短縮サービス](../crafts/url-shortener) | Go | 予定 |

## 言語学習

| 言語 | リポジトリ | 状態 |
|------|-----------|------|
| [Haskell](../languages/haskell) | [study-haskell](https://github.com/tomo-local/study-haskell) | 🚧 |
| [C](../languages/c) | [study-C](https://github.com/tomo-local/study-C) | 🚧 |
| [Ruby](../languages/ruby) | — | 🚧 |

## ライブラリ・ランタイム

| ライブラリ / ランタイム | リポジトリ | 状態 |
|----------------------|-----------|------|
| [htmx](../libraries/htmx) | [study-htmx](https://github.com/tomo-local/study-htmx) | 🚧 |
| [Deno](../libraries/deno) | [study-deno](https://github.com/tomo-local/study-deno) | 🚧 |
| [gRPC](../libraries/grpc) | [study-gRPC](https://github.com/tomo-local/study-gRPC) | 🚧 |
| [XState](../libraries/xstate) | [study-xstate](https://github.com/tomo-local/study-xstate) | 🚧 |
| [Plasmo](../libraries/plasmo) | [study-plasmo](https://github.com/tomo-local/study-plasmo) | 🚧 |
| [WXT](../libraries/wxt) | [study-wxt](https://github.com/tomo-local/study-wxt) | 🚧 |
| [Jotai](../libraries/jotai) | [study-jotai](https://github.com/tomo-local/study-jotai) | 🚧 |

## ディレクトリ構成

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

## 方針

- 作ることで理解する
- 同じ概念を別言語で再実装して、思考の差異を体験する
- 完全性よりシンプルさを優先する
