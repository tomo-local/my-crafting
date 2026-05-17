# Step 3 実装ガイド：リクエストラインのパースとルーティング

## ゴール

- `GET /` → 200 トップページ
- `GET /about` → 200 About ページ
- それ以外 → 404 Not Found

```bash
curl http://localhost:8080/
# → 200 トップページ

curl http://localhost:8080/about
# → 200 About ページ

curl -v http://localhost:8080/missing
# → HTTP/1.1 404 Not Found
```

---

## 変更するファイル

```
go/
└── main.go    ← handleConn を修正
```

---

## 1. `main()`

Step 2 から変更なし。

---

## 2. `handleConn(conn net.Conn)`

**内部でやること（順番どおり）:**

1. `defer conn.Close()`
2. バッファ `buf := make([]byte, 4096)` を用意する
3. `n, err := conn.Read(buf)` でリクエストを読む
4. エラーなら return
5. リクエスト文字列を取り出す

   ```go
   reqText := string(buf[:n])
   ```

6. 1 行目（リクエストライン）を取り出す

   ```go
   lines := strings.Split(reqText, "\r\n")
   fields := strings.Fields(lines[0]) // ["GET", "/about", "HTTP/1.1"]
   ```

7. `len(fields) < 2` のとき（不正なリクエスト）は 400 を返して return
8. パス `fields[1]` で分岐してステータスとボディを決める

   ```go
   var status, body string
   switch fields[1] {
   case "/":
       status = "200 OK"
       body = "Welcome!"
   case "/about":
       status = "200 OK"
       body = "About page"
   default:
       status = "404 Not Found"
       body = "Not Found"
   }
   ```

9. レスポンスを組み立てて `conn.Write` する

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: 各パスの確認

```bash
curl -s http://localhost:8080/
curl -s http://localhost:8080/about
curl -v http://localhost:8080/no-such-path
```

3 番目で `HTTP/1.1 404 Not Found` が返ること。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| すべてのパスが 200 になる | `fields[1]` でなく `fields[0]`（メソッド）を比較している | インデックスを確認する |
| ブラウザで `/` にアクセスすると `/favicon.ico` も来る | ブラウザが自動でファビコンをリクエストする | `/favicon.ico` を default の 404 に任せるか明示的に処理する |
| `index out of range` パニック | フィールドが 3 つ未満のリクエストが来た | `len(fields) < 2` チェックを先に入れる |
