# Step 4 実装ガイド：リクエストボディの解析

## ゴール

`/echo` エンドポイントに POST したボディがそのまま返ってくること。

```bash
curl -X POST http://localhost:8080/echo \
  -H "Content-Type: text/plain" \
  -d "hello"
# → hello

curl -X POST http://localhost:8080/echo \
  -d "foo bar baz"
# → foo bar baz
```

---

## 変更するファイル

```
go/
└── main.go    ← handleConn を全面的に書き直す
```

---

## 1. `main()`

Step 3 から変更なし。

---

## 2. `handleConn(conn net.Conn)`

Step 4 から `conn.Read` の代わりに `bufio.Reader` を使います。

**内部でやること（順番どおり）:**

1. `defer conn.Close()`
2. `reader := bufio.NewReader(conn)` を作る
3. リクエストラインを読む

   ```go
   requestLine, err := reader.ReadString('\n')
   if err != nil { return }
   fields := strings.Fields(strings.TrimRight(requestLine, "\r\n"))
   if len(fields) < 2 { return }
   method, path := fields[0], fields[1]
   ```

4. ヘッダーを 1 行ずつ読んで `Content-Length` だけ取り出す

   ```go
   contentLength := 0
   for {
       line, err := reader.ReadString('\n')
       if err != nil { return }
       if line == "\r\n" { break } // 空行でヘッダー終端
       if strings.HasPrefix(line, "Content-Length:") {
           parts := strings.SplitN(line, ":", 2)
           contentLength, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
       }
   }
   ```

5. ボディを読む（`contentLength > 0` の場合のみ）

   ```go
   var bodyBytes []byte
   if contentLength > 0 {
       bodyBytes = make([]byte, contentLength)
       if _, err := io.ReadFull(reader, bodyBytes); err != nil { return }
   }
   ```

6. `method` と `path` でルーティング

   ```go
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
   ```

7. レスポンスを組み立てて `conn.Write` する

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: POST /echo の確認

```bash
curl -X POST http://localhost:8080/echo -d "hello"
# → hello
```

### ステップ 3: 長いボディの確認

```bash
curl -X POST http://localhost:8080/echo -d "$(python3 -c 'print("a"*5000)')"
# → aaaa...（5000 文字）が返ること
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| curl がフリーズする | `io.ReadFull` が `contentLength` バイト待ち続けている | `Content-Length` のパースが正しいか確認する |
| ボディが空で返ってくる | 空行の検出が `"\n"` になっており `"\r\n"` を見逃している | `line == "\r\n"` を確認する |
| ボディの末尾に余分な文字が付く | `contentLength` が実際のボディより大きい | `curl -v` で実際の `Content-Length` ヘッダーを確認する |
| 大きなボディで切れる | `contentLength` 上限を設けていない | 本番では上限チェックが必要（Step 4 では省略可） |
