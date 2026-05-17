# Step 6：Keep-Alive（前提知識）

Step 5 までは 1 接続 = 1 リクエストで `Close` していました。Step 6 では **1 つの TCP 接続で複数のリクエストを処理する** Keep-Alive を実装します。

---

## 1. なぜ Keep-Alive が必要か

HTTP/1.0 では 1 リクエストごとに TCP 接続を開いて閉じていました。しかし Web ページには HTML・CSS・JS・画像など多数のリソースがあり、毎回 TCP ハンドシェイクをすると遅延が積み重なります。

HTTP/1.1 では `Connection: close` を明示しない限り**接続を維持するのがデフォルト**（Keep-Alive）です。

---

## 2. 変更点：接続を閉じないループ

現在の `handleConn` は「1 回読んで → 1 回書いて → 閉じる」です。Keep-Alive では「読んで → 書く」を**ループで繰り返し**、クライアントが接続を切るまで続けます。

```
// Before（1 リクエストで終了）
Read → Write → Close

// After（Keep-Alive）
for {
    Read → if EOF then break
    Write
}
Close
```

---

## 3. 接続終了の検出

クライアントが接続を切ると `conn.Read` が `io.EOF` を返します。これがループの終了条件です。

また、レスポンスに `Connection: close` ヘッダーを付けてクライアント側から切断を促すことも可能です。Step 6 では `Connection: keep-alive` を返し、接続を維持します。

---

## 4. タイムアウトの設定（実用上の考慮）

Keep-Alive 接続はクライアントが切るまで goroutine が生き続けます。無操作のまま放置されると goroutine がリークします。実用的には `conn.SetReadDeadline` でタイムアウトを設けて、一定時間リクエストが来なければ切断します。

```go
conn.SetReadDeadline(time.Now().Add(30 * time.Second))
```

Step 6 ではタイムアウト実装は任意です。

---

## 📌 まとめ：Step 6 のフロー

1. **`Accept`** で接続を受け入れる
2. **`bufio.NewReader`** を作る（接続ごとに 1 つ、ループの外で作る）
3. リクエスト読み取りループを開始する
4. リクエストラインを読む → **`io.EOF` なら break**
5. ヘッダー・ボディを読む
6. ルーティングしてレスポンスを **`Write`**
7. ループ先頭に戻って次のリクエストを待つ
8. ループを抜けたら **`Close`**
