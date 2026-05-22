# Step 3 実装ガイド：Subscribe/Unsubscribeのライフサイクル

## ゴール

`UNSUB` および切断後、そのサブスクライバーへの配信が止まりブローカーがクラッシュしないこと。

```bash
# ターミナル B: SUB news してから UNSUB news
nc localhost 4222
SUB news
# → +OK
UNSUB news
# → +OK

# ターミナル C: PUB news Hello
nc localhost 4222
PUB news Hello
# → B には届かないこと

# ターミナル B を Ctrl+C で切断後、PUB しても panic しないこと
```

---

## 変更するファイル

```
go/
├── main.go        ← handleClient の切断処理を修正
└── broker/
    └── broker.go  ← Unsubscribe 追加、Subscriber 構造体の拡張
```

---

## 1. `broker/broker.go` の修正

### `Subscriber` 構造体の拡張

> **Step 2 との差分**
> 1トピック固定から、複数トピックを管理する構造に変える。

```go
type Subscription struct {
    ch    chan string
    topic string
}

type Subscriber struct {
    conn          net.Conn
    mu            sync.Mutex
    subscriptions map[string]*Subscription  // topic → Subscription
}

func NewSubscriber(conn net.Conn) *Subscriber {
    return &Subscriber{
        conn:          conn,
        subscriptions: make(map[string]*Subscription),
    }
}
```

### `(b *Broker) Subscribe(topic string, sub *Subscriber) chan string`

1. `b.mu.Lock()` / `defer b.mu.Unlock()`
2. `sub.mu.Lock()` で重複チェック: 既に購読済みなら既存の `ch` を返す
3. `sub.mu.Unlock()`
4. 新しい `Subscription{ch: make(chan string, 64), topic: topic}` を作る
5. `sub.subscriptions[topic] = subscription`
6. `b.subscribers[topic] = append(b.subscribers[topic], sub)` でブローカーリストに追加する
7. `subscription.ch` を返す

### `(b *Broker) Unsubscribe(topic string, sub *Subscriber)`

1. `b.mu.Lock()` / `defer b.mu.Unlock()`
2. `b.subscribers[topic]` から `sub` を除去する
3. `sub.mu.Lock()` で `sub.subscriptions[topic]` の `ch` を取得する
4. `delete(sub.subscriptions, topic)` する
5. `sub.mu.Unlock()`
6. `close(ch)` する（`writeLoop` を終了させる）

### `(b *Broker) UnsubscribeAll(sub *Subscriber)`

全トピックに対して `Unsubscribe` を呼ぶ。切断時に使う。

```go
func (b *Broker) UnsubscribeAll(sub *Subscriber) {
    sub.mu.Lock()
    topics := make([]string, 0, len(sub.subscriptions))
    for topic := range sub.subscriptions {
        topics = append(topics, topic)
    }
    sub.mu.Unlock()

    for _, topic := range topics {
        b.Unsubscribe(topic, sub)
    }
}
```

### `(b *Broker) Publish(topic, message string)` の修正

> **Step 2 との差分**
> チャネルへの書き込み先が `sub.ch` から `sub.subscriptions[topic].ch` に変わる。

```go
func (b *Broker) Publish(topic, message string) {
    b.mu.RLock()
    subs := make([]*Subscriber, len(b.subscribers[topic]))
    copy(subs, b.subscribers[topic])
    b.mu.RUnlock()

    for _, sub := range subs {
        sub.mu.Lock()
        subscription, ok := sub.subscriptions[topic]
        sub.mu.Unlock()
        if ok {
            subscription.ch <- message
        }
    }
}
```

---

## 2. `main.go` の修正

### `handleClient` の修正

> **Step 2 との差分**
> `readLoop` 終了後に `UnsubscribeAll` を呼ぶ。

```go
func handleClient(conn net.Conn, b *broker.Broker) {
    defer conn.Close()
    sub := broker.NewSubscriber(conn)
    readLoop(sub, b)
    b.UnsubscribeAll(sub)  // ← 切断時に全購読を解除
}
```

### `readLoop` の修正

`SUB` 処理でチャネルを受け取り、`go writeLoop` を起動する。

```go
case "SUB":
    topic := fields[1]
    ch := b.Subscribe(topic, sub)
    go writeLoop(topic, ch, sub.conn)
    fmt.Fprintf(sub.conn, "+OK\r\n")

case "UNSUB":
    topic := fields[1]
    b.Unsubscribe(topic, sub)
    fmt.Fprintf(sub.conn, "+OK\r\n")
```

### `writeLoop` のシグネチャ変更

```go
func writeLoop(topic string, ch chan string, conn net.Conn) {
    for msg := range ch {
        fmt.Fprintf(conn, "MSG %s %s\r\n", topic, msg)
    }
}
```

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: UNSUB 後に届かないことを確認

```bash
SUB news → UNSUB news → PUB news Hello
# B には届かないこと
```

### ステップ 3: goroutine リークがないことを確認

```bash
go run main.go -port 4222
```

複数クライアントを接続・切断した後、ブローカー側のgoroutine数が増え続けないことを確認する。（`runtime.NumGoroutine()` を定期ログに追加して確認）

### ステップ 4: -race でデータ競合がないこと

```bash
go run -race main.go -port 4222
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `panic: send on closed channel` | `close(ch)` 後に `Publish` がそのチャネルに書き込んでいる | `Unsubscribe` でリストから削除してから `close` する順序を守る |
| UNSUB 後も MSG が届く | リストから削除できていない（インデックスの比較ではなくポインタ比較が必要） | `s != target`（ポインタ比較）でフィルタリングする |
| goroutine が終了しない | `close(ch)` が呼ばれていない | `UnsubscribeAll` 内で全チャネルが `close` されているか確認 |
| 切断後にブローカーがフリーズ | `Publish` の `sub.ch <- message` がブロックしている（切断後に誰も読まないチャネルに書いている） | `close` 済みチャネルには書かない（`UnsubscribeAll` と `Publish` の排他制御を確認） |
