# Step 2 実装ガイド：プロキシヘッダーの付加

## ゴール

アップストリームに届くリクエストに `X-Forwarded-For`、`Via`、正しい `Host` が含まれること。

```bash
# アップストリームをヘッダーエコーモードで起動（受け取ったヘッダーをレスポンスに返す）
go run upstream/main.go -port 9001 -echo-headers

curl -s http://localhost:8080/
# レスポンスに以下が含まれること:
# X-Forwarded-For: 127.0.0.1
# Via: 1.1 reverse-proxy
# Host: localhost:9001   （プロキシのアドレスではなくアップストリームのアドレス）
```

---

## 変更するファイル

```
go/
├── main.go              ← handleConn を大幅修正
└── upstream/
    └── main.go          ← -echo-headers フラグを追加
```

---

## 1. `upstream/main.go` への `-echo-headers` 追加

`-echo-headers` フラグが true のとき、受け取ったリクエストヘッダーをそのままレスポンスボディに書き出すように修正する。

```go
if *echoHeaders {
    for key, values := range r.Header {
        fmt.Fprintf(w, "%s: %s\n", key, strings.Join(values, ", "))
    }
}
```

---

## 2. hop-by-hop ヘッダーのリスト定義

`main.go` のパッケージスコープに定数として定義する。

```go
var hopByHopHeaders = []string{
    "Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
    "TE", "Trailers", "Transfer-Encoding", "Upgrade",
}
```

---

## 3. `removeHopByHopHeaders(header http.Header)` 関数

**内部でやること（順番どおり）:**

1. `header.Get("Connection")` で `Connection` ヘッダーの値を取得する
2. カンマ区切りで分割し、各エントリを `header.Del()` で削除する（`Connection: close, X-Custom` のようなケースへの対応）
3. `hopByHopHeaders` のリストを `range` で回して `header.Del()` する

---

## 4. `handleConn(client net.Conn, upstream string)` の修正

> **Step 1 との差分**
> `io.Copy` による生ストリームコピーを廃止し、`http.ReadRequest` でパースしてから転送する方式に変える。

**内部でやること（順番どおり）:**

1. `defer client.Close()`
2. `reader := bufio.NewReader(client)` でバッファリングリーダーを作る
3. `req, err := http.ReadRequest(reader)` でリクエストを解析する
4. エラーなら `log.Println(err)` して `return`
5. `removeHopByHopHeaders(req.Header)` を呼ぶ
6. `req.Host` をアップストリームのアドレスに設定する（`req.Host = upstream`）
7. クライアントIPを `X-Forwarded-For` に追加する

   ```go
   clientIP, _, _ := net.SplitHostPort(client.RemoteAddr().String())
   if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
       clientIP = prior + ", " + clientIP
   }
   req.Header.Set("X-Forwarded-For", clientIP)
   ```

8. `Via` ヘッダーを追加する

   ```go
   req.Header.Set("Via", "1.1 reverse-proxy")
   ```

9. `net.Dial("tcp", upstream)` でアップストリームに接続する
10. `defer upstreamConn.Close()`
11. `req.Write(upstreamConn)` で書き換えたリクエストを送信する
12. `io.Copy(client, upstreamConn)` でレスポンスをクライアントに返す

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: ヘッダーの確認

```bash
go run upstream/main.go -port 9001 -echo-headers &
go run main.go -upstream localhost:9001 -port 8080 &

curl -s http://localhost:8080/
```

期待する出力（ヘッダーの一部）:
```
X-Forwarded-For: 127.0.0.1  # curl 127.0.0.1:8080 の場合
X-Forwarded-For: ::1         # curl localhost:8080 の場合（macOS は IPv6 で解決されるため）
Via: 1.1 reverse-proxy
Host: localhost:9001
```

> **注意**: macOS では `localhost` が IPv6 (`::1`) で解決されるため、`curl localhost:8080` を使うと `X-Forwarded-For: ::1` になります。`127.0.0.1` を確認したい場合は `curl 127.0.0.1:8080` で明示的に IPv4 を指定してください。

### ステップ 3: `Host` がプロキシのアドレスでないことを確認

`Host: localhost:8080` がレスポンスに**含まれていない**こと。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `http.ReadRequest` がエラーを返す | クライアントが HTTP/1.1 以外を送っている（CONNECT メソッドなど） | Step 2 では HTTP/1.1 の GET/POST のみを想定してよい |
| レスポンスが返ってこない（ハング） | `req.Write` 後の `io.Copy` を goroutine にしていない、かつリクエストにボディがある | Step 2 ではリクエストボディなし（GET）で確認する |
| `Host` が書き換わっていない | `req.Header.Set("Host", ...)` では変わらない | `req.Host = upstream` を使う（`Header` の `Host` と `req.Host` は別物） |
| `X-Forwarded-For` に `[::1]:port` が入る | `RemoteAddr()` がIPv6ループバックを返す | `net.SplitHostPort` でIPだけ抜き出して使う |
