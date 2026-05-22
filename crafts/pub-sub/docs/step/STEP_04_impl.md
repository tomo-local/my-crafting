# Step 4 実装ガイド：バックプレッシャーと遅いサブスクライバー対策

## ゴール

遅いサブスクライバーがいても速いサブスクライバーへの配信が止まらないこと。遅いサブスクライバーはドロップログが出ること。

```bash
# 遅いサブスクライバーを起動（500ms delay）
go run tools/subscriber/main.go -port 4222 -topic news -delay 500ms

# 速いサブスクライバーを起動
go run tools/subscriber/main.go -port 4222 -topic news -delay 0ms

# 大量メッセージをパブリッシュ
go run tools/publisher/main.go -port 4222 -topic news -count 200 -interval 1ms

# → 速いサブスクライバーはほぼ全件受信
# → ブローカーのログにドロップ件数が出ること
# → ブローカーがハングしないこと
```

---

## 変更するファイル

```
go/
├── broker/
│   └── broker.go        ← Publish をノンブロッキングに変更、dropped フィールド追加
└── tools/
    ├── subscriber/
    │   └── main.go      ← 新規作成（-delay フラグ付きサブスクライバー）
    └── publisher/
        └── main.go      ← 新規作成（-count, -interval フラグ付きパブリッシャー）
```

---

## 1. `broker/broker.go` の修正

### `Subscriber` に `dropped` フィールドを追加

```go
type Subscriber struct {
    conn          net.Conn
    mu            sync.Mutex
    subscriptions map[string]*Subscription
    dropped       int64  // ← 追加
}
```

### `Publish` をノンブロッキングに変更

> **Step 3 との差分**
> `subscription.ch <- message` を `select` に変える。

```go
for _, sub := range subs {
    sub.mu.Lock()
    subscription, ok := sub.subscriptions[topic]
    sub.mu.Unlock()
    if !ok {
        continue
    }

    select {
    case subscription.ch <- message:
    default:
        dropped := atomic.AddInt64(&sub.dropped, 1)
        if dropped%100 == 0 {  // 100件ごとにログ
            log.Printf("[WARN] %s topic=%s dropped=%d",
                sub.conn.RemoteAddr(), topic, dropped)
        }
    }
}
```

---

## 2. `tools/subscriber/main.go`

動作確認用のサブスクライバークライアント。

**フラグ:**
- `-port`: ブローカーのポート（デフォルト `4222`）
- `-topic`: 購読するトピック
- `-delay`: 各メッセージの処理時間（`time.Duration`、デフォルト `0`）

**実装:**
1. `net.Dial("tcp", ...)` でブローカーに接続する
2. `fmt.Fprintf(conn, "SUB %s\r\n", topic)` で購読する
3. `bufio.Scanner` で1行ずつ読む
4. `MSG` で始まる行なら `time.Sleep(*delay)` を呼んでから受信ログを出す

---

## 3. `tools/publisher/main.go`

動作確認用のパブリッシャークライアント。

**フラグ:**
- `-port`: ブローカーのポート
- `-topic`: 発行するトピック
- `-count`: 発行件数
- `-interval`: 発行間隔（`time.Duration`）

**実装:**
1. `net.Dial` で接続する
2. `for i := 0; i < count; i++` でループする
3. `fmt.Fprintf(conn, "PUB %s message-%d\r\n", topic, i)` を送る
4. `time.Sleep(*interval)` を挟む

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: ドロップが発生することを確認

```bash
go run tools/subscriber/main.go -topic news -delay 500ms &
go run tools/publisher/main.go -topic news -count 200 -interval 1ms
```

ブローカーのログに `[WARN] ... dropped=100` のような出力が出ること。

### ステップ 3: 速いサブスクライバーへの影響がないことを確認

```bash
go run tools/subscriber/main.go -topic news -delay 500ms &
go run tools/subscriber/main.go -topic news -delay 0ms &
go run tools/publisher/main.go -topic news -count 100 -interval 1ms
```

速いサブスクライバーは100件全て受信し、遅いサブスクライバーは数件しか受信できないこと。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| ドロップが発生しない | チャネルのバッファが大きすぎる | `make(chan string, 64)` を小さくして（例: 8）試す |
| 速いサブスクライバーへの配信も止まる | `select` の `default` を追加し忘れた | `Publish` で `subscription.ch <- message` が直書きになっていないか確認 |
| `dropped` のカウントが不正確 | `atomic.AddInt64` でなく `sub.dropped++` を使っている | 必ず `atomic.AddInt64(&sub.dropped, 1)` を使う |
