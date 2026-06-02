# Step 2 実装ガイド：Watcher（pub/sub）

## ゴール

```bash
go test ./go/internal/watcher/...
# ok  github.com/tomo-local/firestore/internal/watcher
```

---

## 変更するファイル

```
go/
└── internal/
    └── watcher/
        ├── watcher.go
        └── watcher_test.go
```

---

## `watcher.go` の実装手順

### 1. 型定義

```go
type ChangeType string

const (
    Added    ChangeType = "ADDED"
    Modified ChangeType = "MODIFIED"
    Removed  ChangeType = "REMOVED"
)

type Event struct {
    ChangeType ChangeType
    Path       string
    Document   *store.Document // REMOVED のときは nil
    Version    store.SnapshotVersion
}
```

### 2. Subscription と Watcher

```go
type Subscription struct {
    ch       chan Event
    targetID int
}

type Watcher struct {
    mu   sync.Mutex
    subs map[string][]*Subscription
}

func NewWatcher() *Watcher {
    return &Watcher{subs: make(map[string][]*Subscription)}
}
```

### 3. Subscribe

内部でやること（順番どおり）:
1. `w.mu.Lock()` / `defer w.mu.Unlock()`
2. `ch := make(chan Event, 64)` を作る
3. `sub := &Subscription{ch: ch, targetID: targetID}` を作る
4. `w.subs[path] = append(w.subs[path], sub)` に追加
5. `return ch` を返す（`<-chan Event` にキャスト）

### 4. Unsubscribe

内部でやること（順番どおり）:
1. `w.mu.Lock()` / `defer w.mu.Unlock()`
2. `w.subs[path]` をループして `targetID` が一致するものを除いたスライスで置き換える
3. スライスが空になったら `delete(w.subs, path)`

### 5. Publish

内部でやること（順番どおり）:
1. `w.mu.Lock()` / `defer w.mu.Unlock()`
2. `w.subs[path]` をループする
3. 各 `sub.ch` に `select { case sub.ch <- event: default: }` で送る（ブロック回避）

---

## `watcher_test.go` の実装手順

以下の3ケースをテストする:

1. `Subscribe` → `Publish` でチャネルにイベントが届くこと
2. `Unsubscribe` 後に `Publish` してもチャネルに届かないこと
3. 2つの subscriber がいるとき、両方に届くこと

テストの書き方ヒント:

```go
ch := w.Subscribe("users/alice", 1)
w.Publish("users/alice", Event{...})
select {
case ev := <-ch:
    // assert ev
case <-time.After(100 * time.Millisecond):
    t.Fatal("no event received")
}
```

---

## 実装の確認手順

```bash
go build ./go/...
go test ./go/internal/watcher/...
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `Publish` が block する | バッファなしチャネル + `<-` で直接送っている | `select { case ch <- e: default: }` に変える |
| `Unsubscribe` 後もイベントが届く | スライスの更新がコピーで元を変えていない | `w.subs[path] = newSlice` で置き換える |
| `-race` でデータ競合 | `Publish` と `Subscribe` が同じ `subs` に無保護でアクセス | `Publish` 内でも `mu.Lock()` する |
