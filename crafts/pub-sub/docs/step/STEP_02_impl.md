# Step 2 実装ガイド：ファンアウト

## ゴール

同一トピックのサブスクライバー全員にメッセージが届くこと。

```bash
# ターミナル B: SUB news
# ターミナル C: SUB news

# ターミナル D: PUB news Hello
# → B と C 両方に MSG news Hello が届くこと

# B と C が届くタイミングはほぼ同時であること
```

---

## 変更するファイル

```
go/
└── broker/
    └── broker.go    ← Publish のループを複数サブスクライバー対応に確認
```

---

## 1. `broker.go` の確認と修正

Step 1 で `Publish` のループを正しく実装していれば、コード上の変更はほぼありません。以下を確認してください。

**確認ポイント:**

1. `b.subscribers[topic]` が `[]*Subscriber` のスライスであること
2. `for _, sub := range subs` でスライス全体をループしていること
3. ループの中で `sub.ch <- message` を呼んでいること

**推奨: リストのコピー方式に変更する**

ロック保持時間を短くするために、以下の形に修正します。

```go
func (b *Broker) Publish(topic, message string) {
    b.mu.RLock()
    subs := make([]*Subscriber, len(b.subscribers[topic]))
    copy(subs, b.subscribers[topic])
    b.mu.RUnlock()

    for _, sub := range subs {
        sub.ch <- message
    }
}
```

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: 2サブスクライバーへの配信確認

```bash
# ターミナル1: ブローカー起動
go run main.go -port 4222

# ターミナル2: サブスクライバーA
nc localhost 4222
SUB news

# ターミナル3: サブスクライバーB
nc localhost 4222
SUB news

# ターミナル4: パブリッシャー
nc localhost 4222
PUB news Broadcast!

# → ターミナル2と3に MSG news Broadcast! が届くこと
```

### ステップ 3: -race でデータ競合がないことを確認

```bash
go run -race main.go -port 4222
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| 1人にしか届かない | `Publish` でループせず `subscribers[topic][0]` だけに送っている | `range` でスライス全体をループする |
| 2人に届くが順番がずれる | 配信順序は保証されない（goroutineのスケジューリングによる） | これは正常動作。保証が必要ならシーケンス番号を付ける |
| `-race` でデータ競合 | コピー前に `RUnlock` を呼んでいない | `copy` してから `RUnlock` する順番を確認 |
