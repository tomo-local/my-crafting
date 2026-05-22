# Step 1 実装ガイド：最小Pub/Sub

## ゴール

`nc` でSUB/PUBコマンドを打つと、同トピックのサブスクライバーにMSGが届くこと。

```bash
# ターミナル A: ブローカー起動
go run main.go -port 4222

# ターミナル B: サブスクライバー
nc localhost 4222
SUB news
# → +OK

# ターミナル C: パブリッシャー
nc localhost 4222
PUB news Hello World
# → +OK

# ターミナル B の表示
MSG news Hello World
```

---

## 変更するファイル

```
go/
├── main.go
└── broker/
    └── broker.go    ← 新規作成
```

---

## 1. `broker/broker.go`

### `Subscriber` 構造体

```go
type Subscriber struct {
    conn  net.Conn
    ch    chan string
    topic string
}
```

### `Broker` 構造体

```go
type Broker struct {
    mu          sync.RWMutex
    subscribers map[string][]*Subscriber
}

func New() *Broker {
    return &Broker{subscribers: make(map[string][]*Subscriber)}
}
```

### `(b *Broker) Subscribe(topic string, sub *Subscriber)`

1. `b.mu.Lock()` / `defer b.mu.Unlock()`
2. `b.subscribers[topic] = append(b.subscribers[topic], sub)`

### `(b *Broker) Publish(topic, message string)`

1. `b.mu.RLock()` / `defer b.mu.RUnlock()`
2. `subs := b.subscribers[topic]` でサブスクライバーリストを取得する
3. `for _, sub := range subs { sub.ch <- message }` でファンアウトする

> Step 4 でこの `sub.ch <- message` を non-blocking に変更する。今は blocking でよい。

---

## 2. `main.go`

### `main()`

1. `-port` フラグを定義して `flag.Parse()`
2. `broker.New()` でブローカーを初期化する
3. `net.Listen("tcp", ":"+port)` でリスナーを作る
4. `defer listener.Close()`
5. 起動ログを出力する
6. 無限ループで `Accept` → `go handleClient(conn, b)` する

### `handleClient(conn net.Conn, b *broker.Broker)`

**内部でやること（順番どおり）:**

1. `defer conn.Close()`
2. `sub := &broker.Subscriber{...}` を作る（`ch: make(chan string, 64)`）
3. `go writeLoop(sub)` を起動する
4. `readLoop(sub, b)` を呼ぶ
5. `readLoop` が終わったら（切断）`close(sub.ch)` する

### `writeLoop(sub *broker.Subscriber)`

1. `for msg := range sub.ch { ... }` でチャネルを読む
2. `fmt.Fprintf(sub.conn, "MSG %s %s\r\n", sub.topic, msg)` で書き出す

> `sub.topic` は SUB コマンドで設定される。Step 3 で複数トピック対応に拡張する。

### `readLoop(sub *broker.Subscriber, b *broker.Broker)`

1. `scanner := bufio.NewScanner(sub.conn)` でスキャナーを作る
2. `for scanner.Scan() { ... }` で1行ずつ読む
3. `line := scanner.Text()` を取得する
4. `fields := strings.Fields(line)` で分割する
5. `fields[0]` で `switch` する:

   ```go
   case "SUB":
       topic := fields[1]
       sub.topic = topic
       b.Subscribe(topic, sub)
       fmt.Fprintf(sub.conn, "+OK\r\n")

   case "PUB":
       topic := fields[1]
       message := strings.Join(fields[2:], " ")
       b.Publish(topic, message)
       fmt.Fprintf(sub.conn, "+OK\r\n")

   default:
       fmt.Fprintf(sub.conn, "-ERR unknown command\r\n")
   ```

6. `scanner.Err()` でエラーを確認してログに出す（Scan が false を返したとき）

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: SUB なしで PUB しても crash しないこと

```bash
nc localhost 4222
PUB news Hello
# → +OK （サブスクライバーがいなくても正常終了）
```

### ステップ 3: SUB → PUB の基本フロー

別ターミナルで SUB してから PUB して MSG が届くことを確認。

### ステップ 4: 切断してもブローカーがクラッシュしないこと

```bash
# サブスクライバーの nc を Ctrl+C で切断
# → ブローカーのログにエラーが出るが起動し続けること
# → その後 PUB しても panic しないこと
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `nc` を切断するとブローカーがパニックする | `close(sub.ch)` 後に `sub.ch <- msg` を呼んでいる | `readLoop` 終了後に `close`。Publish 側の排他制御を確認 |
| MSG が届かない（+OK は返る） | `writeLoop` が goroutine になっていない | `go writeLoop(sub)` になっているか確認 |
| MSG のトピック名が空になる | `sub.topic` を設定する前に `writeLoop` が走っている | `writeLoop` は `fmt.Fprintf` 時点で `sub.topic` を参照するので、SUB コマンド処理で `sub.topic = topic` を忘れずに |
| 2回 SUB すると MSG が2回届く | リストに重複追加されている | Step 3 で対処。今は仕様として許容してよい |
