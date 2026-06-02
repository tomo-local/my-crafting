# Step 3 実装ガイド：HTTP サーバー + CRUD エンドポイント

## ゴール

```bash
go run ./go/main.go &

curl -s -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","age":30}'
# → {"status":"ok","version":1}

curl -s http://localhost:8080/v1/collections/users/documents/alice
# → {"name":"Alice","age":30}

curl -s -X DELETE http://localhost:8080/v1/collections/users/documents/alice
# → {"status":"ok","version":2}

curl -s http://localhost:8080/v1/collections/users/documents/alice
# → 404
```

---

## 変更するファイル

```
go/
├── main.go
└── internal/
    └── server/
        └── server.go
```

---

## `server.go` の実装手順

### 1. Server 構造体

```go
type Server struct {
    addr    string
    store   *store.Store
    watcher *watcher.Watcher
}

func New(addr string, s *store.Store, w *watcher.Watcher) *Server {
    return &Server{addr: addr, store: s, watcher: w}
}
```

### 2. ListenAndServe

内部でやること:
1. `mux := http.NewServeMux()` を作る
2. 4つのルートを登録する
3. `http.ListenAndServe(s.addr, mux)` を呼ぶ

```go
mux.HandleFunc("PUT /v1/collections/{col}/documents/{doc}", s.handlePut)
mux.HandleFunc("GET /v1/collections/{col}/documents/{doc}", s.handleGet)
mux.HandleFunc("DELETE /v1/collections/{col}/documents/{doc}", s.handleDelete)
mux.HandleFunc("GET /v1/collections/{col}/documents", s.handleList)
```

### 3. handlePut

内部でやること（順番どおり）:
1. `col`, `doc` を `r.PathValue` で取り出す
2. `path := col + "/" + doc`
3. `json.NewDecoder(r.Body).Decode(&fields)` でボディを読む（失敗 → 400）
4. `_, exists := s.store.Get(path)` で存在チェック
5. `v := s.store.Put(path, fields)`
6. `s.watcher.Publish(path, watcher.Event{...})` — `changeType` は exists ? Modified : Added
7. `json.NewEncoder(w).Encode(map[string]any{"status":"ok","version":v})`

### 4. handleGet

内部でやること:
1. `path` を組み立てる
2. `doc, ok := s.store.Get(path)` — `ok` が false なら 404
3. `json.NewEncoder(w).Encode(doc.Fields)`

### 5. handleDelete

内部でやること:
1. `path` を組み立てる
2. `v, ok := s.store.Delete(path)` — `ok` が false なら 404
3. `s.watcher.Publish(path, watcher.Event{ChangeType: watcher.Removed, ...})`
4. `json.NewEncoder(w).Encode(map[string]any{"status":"ok","version":v})`

### 6. handleList

内部でやること:
1. `col := r.PathValue("col")`
2. `docs := s.store.List(col)`
3. `json.NewEncoder(w).Encode(docs)`

---

## `main.go` の実装手順

```go
db := store.New()
hub := watcher.NewWatcher()
srv := server.New(":8080", db, hub)
slog.Info("starting firestore", "addr", ":8080")
if err := srv.ListenAndServe(); err != nil {
    slog.Error("server error", "err", err)
    os.Exit(1)
}
```

---

## 実装の確認手順

```bash
go build ./go/...

go run ./go/main.go &
curl -s -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -H "Content-Type: application/json" -d '{"name":"Alice"}'
# → {"status":"ok","version":1}
kill %1
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `r.PathValue` が空文字 | Go 1.22 未満 | `go.mod` の go バージョンを確認する |
| PUT が 400 になる | `Content-Type: application/json` を付け忘れ | curl に `-H "Content-Type: application/json"` を追加 |
| List が空を返す | `List` の prefix チェックで末尾 `/` を付け忘れ | `strings.HasPrefix(path, col+"/")` |
