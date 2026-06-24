# Step 4：Pub/Sub ルーティング（前提知識）

---

## このステップで何が変わるか

1対1のエコーから、複数クライアントが「トピック」を通じてメッセージを交換できるブローカーに拡張します。複数の goroutine から共有状態（トピックマップ）を安全に操作するための設計が必要になります。

---

## ブローカーの構造

```
                    ┌─────────────────────────────┐
Client A ──────────→│                             │
SUBSCRIBE:sports    │    Hub                      │──────→ Client B
                    │                             │──────→ Client D
Client C ──────────→│  topics:                    │
PUBLISH:sports:gol! │  "sports" → [B, D]          │
                    │  "news"   → [A, C]          │
                    └─────────────────────────────┘
```

---

## メッセージプロトコルの設計

WebSocket には「チャンネル」の概念がないため、テキストフレームのペイロードでアプリケーションレベルのプロトコルを定義します。

| メッセージ形式 | 意味 |
|---|---|
| `SUBSCRIBE:news` | `news` トピックを購読する |
| `PUBLISH:news:hello` | `news` トピックに `hello` を発行する |

受信時にプレフィックスで分岐:
```
payload を ":" で最大 3 分割
[0] == "SUBSCRIBE" → hub.Subscribe(client, topic)
[0] == "PUBLISH"   → hub.Publish(topic, message)
```

---

## Hub の設計

```go
type Hub struct {
    mu     sync.RWMutex
    topics map[string][]*Client
}
```

| メソッド | ロック | 処理 |
|---|---|---|
| `Subscribe(client, topic)` | `Lock()` | `topics[topic]` に client を追加 |
| `Publish(topic, message)` | `RLock()` | `topics[topic]` の全 client に送信 |
| `Unsubscribe(client)` | `Lock()` | 全トピックから client を削除 |

---

## goroutine per connection と共有状態

各接続は goroutine で独立して動いていますが、Hub は全接続で共有します。

```
goroutine A (Client A) ─┐
goroutine B (Client B) ─┼─→ Hub.topics（共有）← sync.RWMutex で保護
goroutine C (Client C) ─┘
```

- 複数 goroutine が同時に読む（Publish）: `RLock()` で並行 OK
- 書き込み（Subscribe/Unsubscribe）: `Lock()` で排他

---

## Client 構造体と送信チャネル

各接続を表す構造体として `Client` を用意します。

```go
type Client struct {
    conn net.Conn
    send chan []byte
}
```

`send` チャネルで goroutine 間の送信を分離します:
- Publish 側: `client.send <- message`（ノンブロッキング select で溢れたら捨てる）
- 送信 goroutine: `for msg := range client.send` → `WriteFrame(conn, msg)`

この設計により、Publish 側が送信の完了を待たずに次の処理に進めます。

---

## 📌 まとめ: Step 4 のフロー

1. グローバルな `Hub` を作る
2. 接続ごとに `Client` を作り goroutine を起動する
3. 送信用 goroutine を起動する（`send` チャネルを監視して `WriteFrame`）
4. 受信ループで SUBSCRIBE/PUBLISH をコマンド解析して Hub を操作する
5. 接続が切れたら `Unsubscribe` → `close(client.send)` → `conn.Close()`
