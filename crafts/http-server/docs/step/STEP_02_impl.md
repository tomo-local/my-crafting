# Step 2 実装ガイド：HTTP 最小構成でのレスポンス

## ゴール

`curl -v http://localhost:8080` で `HTTP/1.1 200 OK` と `Hello, World!` が返ること。ブラウザでアクセスしても表示されること。

```
# ターミナル A: サーバー起動
go run main.go

# ターミナル B: curl で確認
curl -v http://localhost:8080

# 期待出力（抜粋）
< HTTP/1.1 200 OK
< Content-Length: 13
<
Hello, World!
```

---

## 変更するファイル

```
go/
└── main.go    ← handleConn を修正するだけ
```

---

## 1. `main()`

Step 1 から変更なし。

---

## 2. `handleConn(conn net.Conn)`

**内部でやること（順番どおり）:**

1. `defer conn.Close()`
2. バッファ `buf := make([]byte, 4096)` を用意する
3. `conn.Read(buf)` でリクエストを読み捨てる（エラーのみチェック）
4. レスポンス文字列を組み立てる

   ```go
   body := "Hello, World!"
   response := "HTTP/1.1 200 OK\r\n" +
       "Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
       "\r\n" +
       body
   ```

5. `conn.Write([]byte(response))` でレスポンスを返す

> **なぜ Read が必要か**
> ブラウザはレスポンスを受け取る前にリクエスト全体を送信します。サーバー側が Read を呼ばないと TCP の送信バッファが詰まってデッドロックする場合があります。Step 2 では内容を見なくていいので、1 回 Read して捨てるだけで十分です。

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: curl で確認

```bash
curl -v http://localhost:8080
```

`HTTP/1.1 200 OK` と `Hello, World!` が表示されれば成功。

### ステップ 3: ブラウザで確認

`http://localhost:8080` を開いて `Hello, World!` が表示されること。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| curl がタイムアウトする | Read を呼んでいないため TCP バッファが詰まっている | `conn.Read(buf)` を必ず呼ぶ |
| ブラウザが何も表示しない | `\r\n` でなく `\n` だけになっている | 行末を `\r\n` に統一する |
| `Content-Length` がボディと合わない | 文字列のバイト数と文字数を混同している | `len(body)` はバイト数（ASCII は一致するが日本語は不一致） |
| curl の 2 回目が失敗する | Accept ループが正しく回っていない | `handleConn` が `return` でなく正常終了しているか確認する |
