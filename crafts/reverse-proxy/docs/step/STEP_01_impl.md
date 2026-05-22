# Step 1 実装ガイド：最小リクエスト転送

## ゴール

プロキシ経由でアップストリームにリクエストが届き、レスポンスがクライアントに返ること。

```bash
# ターミナル A: アップストリーム起動
go run upstream/main.go -port 9001 -id upstream-1

# ターミナル B: プロキシ起動
go run main.go -upstream localhost:9001 -port 8080

# ターミナル C: 動作確認
curl -s http://localhost:8080/
# → upstream-1 からのレスポンスが返ること

curl -s http://localhost:8080/any/path
# → 同様にアップストリームのレスポンスが返ること
```

---

## 変更するファイル

```
go/
├── main.go              ← プロキシ本体
└── upstream/
    └── main.go          ← テスト用アップストリームサーバー（新規作成）
```

---

## 1. `upstream/main.go`

動作確認用のシンプルなHTTPサーバー。

**内部でやること（順番どおり）:**

1. `-port` と `-id` フラグを定義して `flag.Parse()` する
2. `http.HandleFunc("/", ...)` でどのパスも受け付けるハンドラを登録する
3. ハンドラの中でレスポンスボディに `id` と受け取ったパスを含めて返す
4. `http.ListenAndServe(":port", nil)` で起動する

```go
fmt.Fprintf(w, "Hello from %s, path: %s\n", *id, r.URL.Path)
```

---

## 2. `main.go`

**フラグの定義:**

1. `-upstream` フラグ（デフォルト: `localhost:9001`）
2. `-port` フラグ（デフォルト: `8080`）
3. `flag.Parse()` を呼ぶ

---

## 3. `main()` の実装

**内部でやること（順番どおり）:**

1. `net.Listen("tcp", ":"+port)` でリスナーを作る
2. エラーなら `log.Fatal(err)`
3. `defer listener.Close()`
4. 起動ログを出力する（例: `Proxy listening on :8080, upstream: localhost:9001`）
5. 無限ループ `for { ... }` を開始する
6. `listener.Accept()` で接続を受け取る
7. エラーなら `log.Println(err)` して `continue`（Fatal にしない — 1接続のエラーでサーバーを落とさない）
8. `go handleConn(conn, upstream)` で goroutine に委譲してすぐ次の Accept へ

---

## 4. `handleConn(client net.Conn, upstream string)` の実装

**内部でやること（順番どおり）:**

1. `defer client.Close()`
2. `net.Dial("tcp", upstream)` でアップストリームに接続する
3. エラーなら `log.Println(err)` して `return`
4. `defer upstreamConn.Close()`
5. `done := make(chan struct{}, 2)` でgoroutine完了通知用チャネルを作る
6. goroutine 1 を起動: `io.Copy(upstreamConn, client)` → 完了したら `done <- struct{}{}` と `upstreamConn.CloseWrite()`

   > `CloseWrite()` は「送信方向だけ閉じる」操作です。アップストリームに EOF を伝えつつ、アップストリームからのレスポンスはまだ受け取れます。

7. goroutine 2 を起動: `io.Copy(client, upstreamConn)` → 完了したら `done <- struct{}{}`
8. `<-done` と `<-done` で両方の完了を待つ

```go
go func() {
    io.Copy(upstreamConn, client)
    upstreamConn.(*net.TCPConn).CloseWrite()
    done <- struct{}{}
}()
go func() {
    io.Copy(client, upstreamConn)
    done <- struct{}{}
}()
<-done
<-done
```

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
cd go
go build ./...
```

### ステップ 2: アップストリーム単体の動作確認

```bash
go run upstream/main.go -port 9001 -id upstream-1
curl http://localhost:9001/hello
# → Hello from upstream-1, path: /hello
```

### ステップ 3: プロキシ経由の動作確認

```bash
go run main.go -upstream localhost:9001 -port 8080
curl http://localhost:8080/hello
# → Hello from upstream-1, path: /hello  （プロキシ経由で同じレスポンス）
```

### ステップ 4: POSTボディの転送確認

```bash
curl -X POST http://localhost:8080/ -d "body-data"
# → エラーなく転送されること（アップストリームがボディを無視していても接続が切れないこと）
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `curl` がハングして返ってこない | `io.Copy` の片方しか goroutine にしていない | 両方を goroutine で並列に実行する |
| 1回目は成功するが2回目がタイムアウトする | Accept ループの外に処理が残っている | `handleConn` が返るまでループが進んでいる — `go handleConn(...)` になっているか確認 |
| アップストリームに接続できない | ポート番号の不一致 or アップストリームが起動していない | `nc -zv localhost 9001` で疎通確認 |
| レスポンスの途中で接続が切れる | `CloseWrite` を呼ばずに `Close` している | 送信完了後は `CloseWrite`、最終的な切断は `defer Close` に任せる |
| goroutine が終了しない | 片方の `io.Copy` が EOF を受け取らない | 送信側の接続を `CloseWrite` してEOFを伝える |
