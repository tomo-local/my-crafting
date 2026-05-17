# Step 6 実装ガイド：Keep-Alive（持続的接続）

## ゴール

1 つの TCP 接続で複数のリクエストを処理できること。

```bash
curl -v --http1.1 http://localhost:8080/
# → Connection: keep-alive でも正常にレスポンスが返ること

# 同じ接続で 3 リクエスト送る
curl -v --http1.1 \
  http://localhost:8080/ \
  http://localhost:8080/about \
  http://localhost:8080/
# → 3 つとも 200 OK が返ること
```

---

## 変更するファイル

```
go/
└── main.go    ← handleConn を修正（リクエストループを追加）
```

---

## 1. `main()`

Step 5 から変更なし（`go handleConn(conn)` のまま）。

---

## 2. `handleConn(conn net.Conn)`

**変更のポイント:**

- `bufio.Reader` を接続ごとに 1 つ作り、**ループの外に置く**
- 「読んで → 書く」をループで繰り返す
- `io.EOF` やタイムアウトエラーでループを抜けて `Close` する

```go
func handleConn(conn net.Conn) {
    defer conn.Close()
    reader := bufio.NewReader(conn) // 接続ごとに 1 つ

    for {
        // (1) リクエストラインを読む
        requestLine, err := reader.ReadString('\n')
        if err != nil {
            // io.EOF = クライアントが接続を切った（正常終了）
            return
        }
        fields := strings.Fields(strings.TrimRight(requestLine, "\r\n"))
        if len(fields) < 2 {
            return
        }
        method, path := fields[0], fields[1]

        // (2) ヘッダーを読む
        contentLength := 0
        for {
            line, err := reader.ReadString('\n')
            if err != nil { return }
            if line == "\r\n" { break }
            if strings.HasPrefix(line, "Content-Length:") {
                parts := strings.SplitN(line, ":", 2)
                contentLength, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
            }
        }

        // (3) ボディを読む
        var bodyBytes []byte
        if contentLength > 0 {
            bodyBytes = make([]byte, contentLength)
            if _, err := io.ReadFull(reader, bodyBytes); err != nil { return }
        }

        // (4) ルーティング
        var status, body string
        switch {
        case method == "POST" && path == "/echo":
            status = "200 OK"
            body = string(bodyBytes)
        case path == "/":
            status = "200 OK"
            body = "Welcome!"
        case path == "/about":
            status = "200 OK"
            body = "About page"
        default:
            status = "404 Not Found"
            body = "Not Found"
        }

        // (5) レスポンスを返す（Connection: keep-alive を付ける）
        response := "HTTP/1.1 " + status + "\r\n" +
            "Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
            "Connection: keep-alive\r\n" +
            "\r\n" +
            body
        if _, err := conn.Write([]byte(response)); err != nil {
            return
        }
        // ループ先頭に戻って次のリクエストを待つ
    }
}
```

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: Keep-Alive で複数リクエストを送る

```bash
curl -v --http1.1 \
  http://localhost:8080/ \
  http://localhost:8080/about \
  http://localhost:8080/
```

`curl -v` の出力で `* Re-using existing connection` が表示されれば、同じ TCP 接続を再利用できています。

### ステップ 3: Connection: close で切断

```bash
curl -v -H "Connection: close" http://localhost:8080/
```

クライアントが `Connection: close` を送った場合は、レスポンスを返したあとで接続を切ることが理想です（発展課題）。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| 2 つ目のリクエストがタイムアウトする | `bufio.Reader` をループ内で毎回作っている | `reader` はループの外、接続ごとに 1 つだけ作る |
| goroutine がリークする | 接続が放置されて `Read` でブロックし続ける | `conn.SetReadDeadline` でタイムアウトを設定する |
| curl が `* Connection #0 to host left intact` を表示しない | `Connection: keep-alive` ヘッダーを返していない | レスポンスヘッダーに追加する |
