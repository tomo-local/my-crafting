# my-crafting

仕組みを深く理解するために、低レイヤーの実装をゼロから自作するリポジトリです。

トピックごとにディレクトリで管理し、同じ内容を複数言語で実装することもあります。

## トピック一覧

| トピック | 言語 | 状態 |
|---------|------|------|
| [HTTP サーバー](../crafts/http-server) | Go | 進行中 |
| [typescript-mini-zod](../crafts/typescript-mini-zod) | TypeScript | 完了 |
| [react-mini-zustand](../crafts/react-mini-zustand) | TypeScript | 完了 |
| [react-mini-jotai](../crafts/react-mini-jotai) | TypeScript | 完了 |
| [react-mini-xstate](../crafts/react-mini-xstate) | TypeScript | 完了 |
| [react-mini-valtio](../crafts/react-mini-valtio) | TypeScript | 完了 |
| [Redis クローン](../crafts/redis-clone) | Go | 予定 |
| [Kafka クローン（メッセージブローカー）](../crafts/kafka-clone) | Go | 予定 |
| [リバースプロキシ / ロードバランサー](../crafts/reverse-proxy) | Go | 予定 |
| [レートリミッター](../crafts/rate-limiter) | Go | 予定 |
| [CRON タスクスケジューラー](../crafts/cron-scheduler) | Go | 予定 |
| [WebSocket ブローカー](../crafts/websocket-broker) | Go | 予定 |
| [分散キャッシュ（一貫性ハッシュ）](../crafts/distributed-cache) | Go | 予定 |
| [ミニ検索エンジン](../crafts/search-engine) | Go | 予定 |
| [URL 短縮サービス](../crafts/url-shortener) | Go | 予定 |

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
└── docs/
    └── README.ja.md
```

## 方針

- 作ることで理解する
- 同じ概念を別言語で再実装して、思考の差異を体験する
- 完全性よりシンプルさを優先する
