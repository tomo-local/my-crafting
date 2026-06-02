# Step 4：Listen ストリーム（前提知識）

## このステップで何が変わるか

CRUD エンドポイントに加えて「長命接続」を追加する。クライアントが `POST /v1/listen` で接続し、`AddTarget` を送ると、以降の変更がプッシュされ続ける。

---

## Server-Sent Events（SSE）の仕組み

```
Client                          Server
  │── POST /v1/listen ──────────▶│
  │   body: {"type":"AddTarget", │
  │     "path":"users/alice",    │
  │     "targetId":1}            │
  │                              │
  │◀── HTTP 200 ─────────────────│
  │    Content-Type:             │
  │      text/event-stream       │
  │                              │
  │◀── data: {...}\n\n ──────────│  (別クライアントが PUT したとき)
  │◀── data: {...}\n\n ──────────│
  │         (接続維持)
```

**通常の HTTP と何が違うか**: サーバーがレスポンスを「書き終えない」。コネクションを保持し続けて、イベントがあるたびにチャンクを書き込む。

---

## `http.Flusher` が必要な理由

Go の `http.ResponseWriter` はデフォルトでレスポンスをバッファリングする。  
`Flush()` を呼ばないと、書き込んだデータがクライアントに届かない。

```
w.Write([]byte("data: {...}\n\n"))
w.(http.Flusher).Flush()   ← これがないとバッファに溜まったまま
```

---

## AddTarget プロトコル

このクラフトでのリクエスト JSON:

```json
{"type": "AddTarget", "path": "users/alice", "targetId": 1}
```

サーバーの処理:
1. リクエストボディを読んで `AddTarget` を解析する
2. `watcher.Subscribe(path, targetID)` でチャネルを得る
3. SSE ヘッダーをセットして `Flush()`
4. goroutine でチャネルを `range` しながら SSE フォーマットで書き込む

---

## クライアント切断の検知

`r.Context().Done()` がクローズされたとき、クライアントが切断している。

```
select {
case event := <-ch:
    // SSE 書き込み
case <-r.Context().Done():
    watcher.Unsubscribe(path, targetID)
    return
}
```

これを怠ると goroutine が永遠にチャネルを待ち続けてリークする。

---

## 📌 まとめ: Step 4 のフロー

1. `server.go` に `handleListen` ハンドラを追加する
2. `POST /v1/listen` ルートを登録する
3. `handleListen` 内:
   a. リクエストボディから `AddTarget` を JSON デコードする
   b. `w` を `http.Flusher` にアサートする（失敗したら 500）
   c. SSE ヘッダー（`Content-Type: text/event-stream` など）をセットする
   d. `w.WriteHeader(200)` + `Flush()`
   e. `ch := s.watcher.Subscribe(path, targetID)` で購読
   f. `defer s.watcher.Unsubscribe(path, targetID)`
   g. `for` ループで `select { case ev := <-ch: ... case <-ctx.Done(): return }`
   h. イベントを `data: <json>\n\n` フォーマットで書き込み `Flush()`
