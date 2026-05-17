# Step 4：リクエストボディの解析（前提知識）

Step 3 まではリクエストラインだけを見ていました。Step 4 では POST リクエストで送られる**ボディを正しく読み取る**方法を学びます。

---

## 1. ヘッダーとボディの境界

HTTP リクエスト全体の構造を改めて確認します。

```
POST /echo HTTP/1.1\r\n        ← リクエストライン
Host: localhost:8080\r\n       ← ヘッダー
Content-Type: text/plain\r\n   ← ヘッダー
Content-Length: 5\r\n          ← ヘッダー
\r\n                           ← 空行（ヘッダー終端）
hello                          ← ボディ（5 バイト）
```

空行（`\r\n\r\n`）がヘッダーとボディの境界です。境界を正確に検出しなければ、ボディの先頭位置がわかりません。

---

## 2. `Content-Length` で読むバイト数を決める

TCP はストリームなので「どこがボディの終わりか」自体には印がありません。HTTP が `Content-Length` ヘッダーで**ボディのバイト数**を明示することで、読む量を決めます。

```
Content-Length: 5
```

この場合、空行のあとの 5 バイトだけ読めばボディ全体です。**それ以上読もうとするとブロックします**（次のリクエストを待ち続ける）。

---

## 3. `bufio.Reader` を使う理由

1 回の `conn.Read` ではヘッダーとボディが一緒に届くか、分割して届くかが不定です。`bufio.Reader` は内部バッファを持ち、複数回の `Read` を透過的にまとめて「1 行ずつ」取り出す操作を可能にします。

```go
reader := bufio.NewReader(conn)

// 1 行ずつヘッダーを読む
for {
    line, _ := reader.ReadString('\n')
    if line == "\r\n" { break } // 空行 = ヘッダー終端
    // Content-Length を探してパース
}

// ボディを contentLength バイトだけ読む
body := make([]byte, contentLength)
io.ReadFull(reader, body)
```

`io.ReadFull` は指定バイト数が揃うまで繰り返し読んでくれるため、分割配送に対しても正確に動きます。

---

## 4. Step 4 でやること

`/echo` エンドポイントを実装し、受け取ったボディをそのままレスポンスボディとして返します（エコーサーバー）。

| パス | メソッド | 動作 |
|---|---|---|
| `/echo` | POST | ボディをそのまま返す |
| それ以外 | 任意 | Step 3 と同じルーティング |

---

## 📌 まとめ：Step 4 のフロー

1. `bufio.NewReader(conn)` を作る
2. `\r\n` 区切りでヘッダーを 1 行ずつ読む
3. `Content-Length` の値を整数としてパースして保持する
4. 空行が来たらヘッダー終端と判断する
5. `io.ReadFull(reader, body)` でボディを読む
6. パスとメソッドでルーティングしてレスポンスを返す
