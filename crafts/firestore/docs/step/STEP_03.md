# Step 3：HTTP サーバー + CRUD エンドポイント（前提知識）

## このステップで何が変わるか

Store と Watcher を HTTP でつなぐ。`PUT` したら Store に書き込まれ、かつ Watcher 経由で購読者にイベントが流れる状態にする。

---

## ルーティング設計

```
PUT    /v1/collections/{col}/documents/{doc}  → Put
GET    /v1/collections/{col}/documents/{doc}  → Get
DELETE /v1/collections/{col}/documents/{doc}  → Delete
GET    /v1/collections/{col}/documents        → List
```

Go 1.22 以降、`net/http` の `ServeMux` はメソッドとパスパラメータをサポートする:

```go
mux.HandleFunc("PUT /v1/collections/{col}/documents/{doc}", h.handlePut)
```

`r.PathValue("col")` でパラメータを取り出せる。

---

## PUT ハンドラの責務

```
1. リクエストボディを JSON デコード → fields map[string]interface{}
2. path = col + "/" + doc を組み立てる
3. store.Put(path, fields) → version を得る
4. watcher.Publish(path, Event{ChangeType: Added or Modified, ...})
5. JSON レスポンス {"status":"ok","version":N} を返す
```

**Added か Modified かの判断**: `Put` を呼ぶ前に `store.Get(path)` で存在チェックする。

---

## エラーレスポンスの設計

| ケース | HTTP status |
|---|---|
| JSON デコード失敗 | 400 Bad Request |
| ドキュメントが存在しない（GET/DELETE） | 404 Not Found |
| その他の内部エラー | 500 Internal Server Error |

---

## 📌 まとめ: Step 3 のフロー

1. `Server` 構造体（`store *store.Store`, `watcher *watcher.Watcher`）を定義する
2. `New(addr string, s *store.Store, w *watcher.Watcher) *Server` を実装する
3. `ListenAndServe()` で `ServeMux` を作りハンドラを登録して `http.ListenAndServe` を呼ぶ
4. `handlePut` — 存在チェック → Put → Publish → JSON レスポンス
5. `handleGet` — Get → JSON レスポンス or 404
6. `handleDelete` — Delete → Publish(Removed) → JSON レスポンス or 404
7. `handleList` — List → JSON レスポンス
8. `main.go` で Store + Watcher + Server を組み立てて起動する
