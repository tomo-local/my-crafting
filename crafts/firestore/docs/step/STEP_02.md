# Step 2：Watcher（前提知識）

## このステップで何が変わるか

Step 1 では Store に書き込んでも誰にも通知されなかった。このステップで **「書き込みを誰かが購読していれば通知する」** pub/sub 層を追加する。

---

## イベント型

Watcher が配信するイベントの構造:

```
Event
├── ChangeType  ("ADDED" / "MODIFIED" / "REMOVED")
├── Path        (例: "users/alice")
├── Document    (*store.Document)  ← REMOVED のときは nil
└── Version     (SnapshotVersion)
```

---

## pub/sub の構造

```
Watcher
└── subs: map[path] → []Subscription

Subscription
├── ch:       chan Event  (バッファ付き)
└── targetId: int
```

**Subscribe(path, targetId)** — `Subscription` を作って `subs[path]` に追加し、チャネルを返す  
**Unsubscribe(path, targetId)** — `subs[path]` から該当エントリを削除する  
**Publish(path, event)** — `subs[path]` の全チャネルに event を送る

---

## チャネルのブロック問題

`Publish` を同期的に送ると、受信が遅い subscriber が詰まって **他の subscriber への配信も止まる**。

対策: `select` + `default` でドロップするか、十分なバッファ（例: 64）を持たせる。  
このクラフトでは **バッファ付きチャネル** で十分。

```
ch := make(chan Event, 64)
```

---

## goroutine リークとの関係

Watcher 自体は goroutine を持たない。**チャネルを読む goroutine（= Listen ハンドラ）** が切断時に `Unsubscribe` を呼ぶ責任を持つ。これは Step 4 で実装する。

---

## 📌 まとめ: Step 2 のフロー

1. `ChangeType` 型と定数（`Added`, `Modified`, `Removed`）を定義する
2. `Event` 構造体を定義する
3. `Subscription` 構造体（`ch chan Event`, `targetID int`）を定義する
4. `Watcher` 構造体（`mu sync.Mutex`, `subs map[string][]*Subscription`）を定義する
5. `NewWatcher() *Watcher` — 初期化
6. `Subscribe(path string, targetID int) <-chan Event` — 購読登録し読み取り専用チャネルを返す
7. `Unsubscribe(path string, targetID int)` — 購読解除
8. `Publish(path string, event Event)` — 全購読者に送信
