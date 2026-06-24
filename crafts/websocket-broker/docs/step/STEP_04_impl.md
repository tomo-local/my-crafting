# Step 4 実装ガイド：Pub/Sub ルーティング

---

## ゴール

```bash
# ターミナル A
wscat -c ws://localhost:8080
> SUBSCRIBE:news

# ターミナル B
wscat -c ws://localhost:8080
> PUBLISH:news:hello everyone

# ターミナル A に "hello everyone" が届くこと
```

---

## 変更するファイル

```
websocket-broker/go/
├── main.go
└── internal/
    └── ws/
        ├── handshake.go
        ├── frame.go
        └── hub.go      # 新規追加
```

---

## 実装手順

### 1. `internal/ws/hub.go`

**Client 構造体**:
```go
type Client struct {
    conn net.Conn
    send chan []byte
}

func NewClient(conn net.Conn) *Client {
    return &Client{conn: conn, send: make(chan []byte, 64)}
}
```

**Hub 構造体と初期化**:
```go
type Hub struct {
    mu     sync.RWMutex
    topics map[string][]*Client
}

func NewHub() *Hub {
    return &Hub{topics: make(map[string][]*Client)}
}
```

**Subscribe(client *Client, topic string)**:
1. `h.mu.Lock()` / `defer h.mu.Unlock()`
2. `h.topics[topic] = append(h.topics[topic], client)`

**Publish(topic string, message []byte)**:
1. `h.mu.RLock()` / `defer h.mu.RUnlock()`
2. `h.topics[topic]` のリストをループ
3. 各 client の send チャネルに送る（詰まっても捨てる）:
   ```go
   select {
   case c.send <- message:
   default:
   }
   ```

**Unsubscribe(client *Client)**:
1. `h.mu.Lock()` / `defer h.mu.Unlock()`
2. 全トピックをループして client を除くスライスを作り直す:
   ```go
   for topic, clients := range h.topics {
       filtered := clients[:0]
       for _, c := range clients {
           if c != client {
               filtered = append(filtered, c)
           }
       }
       h.topics[topic] = filtered
   }
   ```

---

### 2. `main.go` の handleConn を更新

`hub *Hub` を引数に追加し、`main()` で `ws.NewHub()` を作って渡します。

**handleConn(conn net.Conn, hub *ws.Hub) の内部でやること**:

1. `ws.Upgrade(conn)` でハンドシェイク
2. `client := ws.NewClient(conn)`
3. 送信用 goroutine を起動:
   ```go
   go func() {
       for msg := range client.Send {
           ws.WriteFrame(conn, ws.Frame{FIN: true, Opcode: 0x1, Payload: msg})
       }
   }()
   ```
4. 終了処理を defer で登録:
   ```go
   defer func() {
       hub.Unsubscribe(client)
       close(client.Send)
       conn.Close()
   }()
   ```
5. 受信ループ:
   - `ws.ReadFrame(conn)` でフレームを読む
   - Close なら return
   - `strings.SplitN(string(frame.Payload), ":", 3)` でコマンド解析
   - `parts[0] == "SUBSCRIBE"` → `hub.Subscribe(client, parts[1])`
   - `parts[0] == "PUBLISH"` → `hub.Publish(parts[1], []byte(parts[2]))`

---

## 実装の確認手順

```bash
go build ./...
go run main.go

# ターミナル A
wscat -c ws://localhost:8080
> SUBSCRIBE:news

# ターミナル B
wscat -c ws://localhost:8080
> PUBLISH:news:hello
# ターミナル A に "hello" が届く
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| Publish したのに届かない | Subscribe より先に Publish している | 先に SUBSCRIBE を送ってから PUBLISH する |
| データ競合でクラッシュ | topics の読み書きを Mutex で保護していない | Subscribe/Unsubscribe に `Lock()`、Publish に `RLock()` |
| goroutine リーク | `client.send` を `close()` していない | defer で確実に `close(client.send)` する |
| panic: send on closed channel | Unsubscribe 後に Publish が来ている | Publish 内の `select { case ...: default: }` で防げる |
