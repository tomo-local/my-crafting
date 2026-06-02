# Step 4 実装ガイド：Listen ストリーム（SSE）

## ゴール

```bash
# ターミナル 1: 購読開始（接続が維持されたまま待機状態になる）
curl -N -X POST http://localhost:8080/v1/listen \
  -H "Content-Type: application/json" \
  -d '{"type":"AddTarget","path":"users/alice","targetId":1}'

# ターミナル 2: 書き込む
curl -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -H "Content-Type: application/json" -d '{"name":"Alice","age":30}'

# ターミナル 1 に以下が届くこと:
# data: {"changeType":"ADDED","path":"users/alice","version":1,...}
```

---

## 変更するファイル

```
go/
└── internal/
    └── server/
        └── server.go  （handleListen を追加）
```

---

## `handleListen` の実装手順

### 1. リクエスト解析

```go
var req struct {
    Type     string `json:"type"`
    Path     string `json:"path"`
    TargetID int    `json:"targetId"`
}
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

### 2. Flusher アサート

```go
flusher, ok := w.(http.Flusher)
if !ok {
    http.Error(w, "streaming not supported", http.StatusInternalServerError)
    return
}
```

### 3. SSE ヘッダー

内部でやること（順番どおり）:
1. `w.Header().Set("Content-Type", "text/event-stream")`
2. `w.Header().Set("Cache-Control", "no-cache")`
3. `w.Header().Set("Connection", "keep-alive")`
4. `w.WriteHeader(http.StatusOK)`
5. `flusher.Flush()`

### 4. 購読 + イベントループ

内部でやること（順番どおり）:
1. `ch := s.watcher.Subscribe(req.Path, req.TargetID)`
2. `defer s.watcher.Unsubscribe(req.Path, req.TargetID)`
3. `ctx := r.Context()`
4. ループ開始
5. `select { case ev := <-ch: ... case <-ctx.Done(): return }`
6. `ev` を JSON にエンコードして `fmt.Fprintf(w, "data: %s\n\n", jsonBytes)`
7. `flusher.Flush()`

### 5. ルート登録

`ListenAndServe` に以下を追加:

```go
mux.HandleFunc("POST /v1/listen", s.handleListen)
```

---

## 実装の確認手順

```bash
go build ./go/...

# ターミナル 1
go run ./go/main.go

# ターミナル 2
curl -N -X POST http://localhost:8080/v1/listen \
  -H "Content-Type: application/json" \
  -d '{"type":"AddTarget","path":"users/alice","targetId":1}'

# ターミナル 3
curl -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice"}'
# → ターミナル 2 に data: {...} が届くこと
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| curl が何も表示しない | `Flush()` を呼んでいない | `flusher.Flush()` をイベント書き込み後に呼ぶ |
| curl が即座に終了する | `WriteHeader` の前にヘッダー外で書き込みをした | ヘッダーをセットしてから `WriteHeader(200)` の順番を守る |
| goroutine が増え続ける | `ctx.Done()` を見ていない | `select` に `case <-ctx.Done(): return` を追加する |
| `Ctrl+C` で終了しても接続が残る | サーバーがシャットダウンしていない | 今は許容。Step 6 以降で `http.Server.Shutdown` を検討する |
