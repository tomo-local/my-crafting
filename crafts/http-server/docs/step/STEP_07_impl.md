# Step 7 実装ガイド：Connection ヘッダーとグレースフルな接続終了

## ゴール

クライアントが `Connection: close` を送ったとき、レスポンス後に接続が切れること。

```bash
# Connection: close を明示 → レスポンス後に接続が切れる
curl -v -H "Connection: close" http://localhost:8080/
# < Connection: close  ← レスポンスヘッダーに含まれる
# * Closing connection  ← curl が接続終了を検出

# Connection ヘッダーなし（デフォルト）→ 接続が維持される
curl -v --http1.1 http://localhost:8080/ http://localhost:8080/about
# * Re-using existing connection  ← 同じ接続を再利用
```

---

## 変更するファイル

```
go/
├── internal/request/request.go    ← Connection フィールドを追加してパース
├── internal/response/response.go  ← Connection ヘッダーをレスポンスに追加
└── internal/server/server.go      ← close のときループを抜ける
```

---

## 1. `internal/request/request.go`

### `Request` 構造体に `Connection` フィールドを追加

```go
type Request struct {
    Method        string
    Path          string
    Version       string
    ContentLength int
    Connection    string  // "keep-alive" | "close" | ""
    Body          io.Reader
}
```

### `Parse` 関数でヘッダーをパース

ヘッダーループに `Connection:` のチェックを追加します。

```go
if strings.HasPrefix(strings.ToLower(line), "connection:") {
    parts := strings.SplitN(line, ":", 2)
    connection = strings.TrimSpace(strings.ToLower(parts[1]))
}
```

`strings.ToLower` で正規化してから比較します（HTTP ヘッダー名と値は大文字小文字を区別しません）。

---

## 2. `internal/response/response.go`

### `Write` の引数に `connection` を追加

```go
type Write = func(status StatusCode, body string, connection string) error

func (r *Response) Write(status StatusCode, body string, connection string) error {
    if connection == "" {
        connection = "keep-alive"
    }
    response := "HTTP/1.1 " + string(status) + "\r\n" +
        "Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
        "Connection: " + connection + "\r\n" +
        "\r\n" +
        body
    _, err := r.conn.Write([]byte(response))
    return err
}
```

---

## 3. `internal/server/server.go`

### `ServeConn` のループに終了判定を追加

レスポンス書き込み後に `req.Connection` を確認します。

```go
s.handler(req, res.Write)

if req.Connection == "close" {
    break
}
```

---

## 4. `main.go`

`writeResponse` の呼び出し箇所すべてに `req.Connection` を渡します。

```go
writeResponse(response.StatusOK, "Welcome!", req.Connection)
writeResponse(response.StatusNotFound, "Not Found", req.Connection)
// ... 他も同様
```

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: `Connection: close` で切断されること

```bash
curl -v -H "Connection: close" http://localhost:8080/
```

レスポンスヘッダーに `Connection: close` が含まれ、curl が `* Closing connection` を表示すること。

### ステップ 3: デフォルトで Keep-Alive が維持されること

```bash
curl -v --http1.1 \
  http://localhost:8080/ \
  http://localhost:8080/about \
  http://localhost:8080/
```

2 つ目以降で `* Re-using existing connection` が表示されること。

### ステップ 4: HTTP/1.0 クライアントとの互換確認（任意）

```bash
curl -v --http1.0 http://localhost:8080/
```

HTTP/1.0 クライアントは `Connection` ヘッダーを送らないため、現実装ではそのまま Keep-Alive になります（HTTP/1.0 の厳密な対応は発展課題）。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `Connection: close` を送っても接続が切れない | `ServeConn` の break 条件が入っていない | レスポンス後に `req.Connection == "close"` を確認する |
| すべてのリクエストで接続が切れる | `connection` のデフォルト値が `"close"` になっている | 空文字なら `"keep-alive"` にフォールバックする |
| ヘッダーが大文字で来ると `Connection: Keep-Alive` を検出できない | 大文字小文字を区別して比較している | `strings.ToLower` で正規化してから比較する |
| `Write` の呼び出し箇所でコンパイルエラー | 引数の数が変わったのに `main.go` を更新していない | `main.go` の `writeResponse` 呼び出し箇所を全部更新する |
