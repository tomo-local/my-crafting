# Step 6 実装ガイド：Replayバッファ

## ゴール

新規サブスクライブ時に過去N件のメッセージが届いてから通常配信に移ること。

```bash
# ブローカー起動（バッファサイズ 10）
go run main.go -port 4222 -buffer 10

# 先にメッセージを3件パブリッシュ（サブスクライバーなし）
nc localhost 4222
PUB news Message-1
PUB news Message-2
PUB news Message-3

# 後からサブスクライブ（replay=3）
nc localhost 4222
SUB news 3
# → 即座に以下が届くこと:
# MSG news Message-1
# MSG news Message-2
# MSG news Message-3
# その後、新しいメッセージが PUB されれば通常通り届くこと
```

---

## 変更するファイル

```
go/
├── main.go
└── broker/
    ├── broker.go    ← TopicData 構造体に変更、Subscribe/Publish の修正
    └── ringbuf.go   ← 新規作成
```

---

## 1. `broker/ringbuf.go`

**`RingBuffer` 構造体:**

```go
type RingBuffer struct {
    buf  []string
    head int
    size int
    cap  int
}

func NewRingBuffer(cap int) *RingBuffer {
    return &RingBuffer{buf: make([]string, cap), cap: cap}
}
```

**`(r *RingBuffer) Push(msg string)`:**

1. `r.buf[(r.head+r.size)%r.cap] = msg` で書き込む
2. `r.size < r.cap` なら `r.size++`
3. `r.size == r.cap` なら `r.head = (r.head + 1) % r.cap`（古いものを上書き）

**`(r *RingBuffer) Snapshot(n int) []string`:**

1. `n` が `r.size` を超えるなら `r.size` に切り下げる
2. 結果スライスを `make([]string, n)` で作る
3. `start := (r.head + r.size - n + r.cap) % r.cap` で開始インデックスを計算する
4. `for i := 0; i < n; i++` で `result[i] = r.buf[(start+i)%r.cap]` を埋める
5. `result` を返す

---

## 2. `broker/broker.go` の修正

### データ構造の変更

> **Step 5 との差分**
> `map[string][]*Subscriber` から `map[string]*TopicData` に変える。

```go
type TopicData struct {
    subscribers []*Subscriber
    buffer      *RingBuffer
}

type Broker struct {
    mu     sync.RWMutex
    topics map[string]*TopicData
    bufCap int
}

func New(bufCap int) *Broker {
    return &Broker{
        topics: make(map[string]*TopicData),
        bufCap: bufCap,
    }
}
```

### `getOrCreateTopic(topic string) *TopicData`（プライベートヘルパー）

`b.topics[topic]` が nil なら新しい `TopicData` を作って返す。**ロック保持中に呼ぶこと。**

### `(b *Broker) Publish(topic, message string)` の修正

1. `b.mu.Lock()` でロックする（バッファ書き込みのためフルロック）
2. `td := b.getOrCreateTopic(topic)`
3. `td.buffer.Push(message)`
4. サブスクライバーのスライスをコピーする
5. `b.mu.Unlock()`
6. コピーしたスライスにファンアウトする（Step 4 のノンブロッキング送信）

### `(b *Broker) Subscribe(topic string, sub *Subscriber, replayN int) chan string` の修正

> **Step 5 との差分**
> `replayN` 引数を追加。Replay スナップショットをロック内で取得してからロック解除。

```go
func (b *Broker) Subscribe(topic string, sub *Subscriber, replayN int) chan string {
    ch := make(chan string, 64)

    b.mu.Lock()
    td := b.getOrCreateTopic(topic)

    // スナップショットを取る（ロック中）
    var snapshot []string
    if replayN > 0 {
        snapshot = td.buffer.Snapshot(replayN)
    }

    // サブスクライバーをリストに追加（以降の Publish はここに届く）
    sub.mu.Lock()
    sub.subscriptions[topic] = &Subscription{ch: ch, topic: topic}
    sub.mu.Unlock()
    td.subscribers = append(td.subscribers, sub)

    b.mu.Unlock()

    // Replay（ロック外で送信）
    for _, msg := range snapshot {
        ch <- msg
    }

    return ch
}
```

---

## 3. `main.go` の修正

**`-buffer` フラグを追加:**

```go
bufSize := flag.Int("buffer", 100, "replay buffer size per topic")
```

`broker.New(*bufSize)` に渡す。

**`readLoop` の SUB パース修正:**

```go
case "SUB":
    topic := fields[1]
    replayN := 0
    if len(fields) > 2 {
        replayN, _ = strconv.Atoi(fields[2])
    }
    ch := b.Subscribe(topic, sub, replayN)
    go writeLoop(topic, ch, sub.conn)
    fmt.Fprintf(sub.conn, "+OK\r\n")
```

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: `RingBuffer` のユニットテスト

```go
// broker/ringbuf_test.go
func TestRingBuffer(t *testing.T) {
    r := NewRingBuffer(3)
    r.Push("a"); r.Push("b"); r.Push("c"); r.Push("d")  // "a" が上書きされる

    snap := r.Snapshot(3)
    // → ["b", "c", "d"]
    if snap[0] != "b" || snap[1] != "c" || snap[2] != "d" {
        t.Fatalf("unexpected: %v", snap)
    }
}
```

```bash
go test ./broker/...
```

### ステップ 3: Replay なしの通常動作が壊れていないこと

```bash
SUB news  # replayN=0
PUB news Hello
# → MSG news Hello が届くこと
```

### ステップ 4: Replay の動作確認

上記ゴールのコマンドを実行する。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| Replay とリアルタイムが逆順になる | Replay をロック外でスナップショット取得している | スナップショットは必ずロック中に取り、サブスクライバー追加もロック中に行う |
| Replay メッセージが重複する | Replay 中にリストに追加されて Publish も届いている + Replay も届いている | 上記の順序（ロック中にスナップショット + リスト追加）を守れば重複しない |
| `Snapshot` が空を返す | Push 前に Snapshot を呼んでいる | Publish でバッファに Push してからサブスクライバーにファンアウトする順序を確認 |
| リングバッファのインデックスがパニック | `(start+i)%r.cap` の計算で `r.cap` が 0 になっている | `New(0)` で作らないようにフラグのデフォルト値と最小値チェックを追加する |
