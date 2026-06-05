# Step 6 実装ガイド（Ruby）：Keep-Alive（持続的接続）

## ゴール

1 つの TCP 接続で複数のリクエストを処理できること。

```bash
# 同じ接続で 3 リクエスト送る
curl -v --http1.1 \
  http://localhost:8080/ \
  http://localhost:8080/about \
  http://localhost:8080/
# → 3 つとも 200 OK が返り、curl の出力に
#   "* Re-using existing connection" が表示されること
```

---

## 現状の整理

Step 5 で `serve_conn` に per-request ループを実装済み。ループのインフラは揃っている。

```ruby
# serve_conn（現状）
loop do
  request = HttpServer::Request.parse(socket)
  keep_alive = request.wants_keep_alive?
  @handler.call(socket, request)
  break unless keep_alive   # ← keep_alive が true なら次のリクエストへ
end
```

**ただし `main.rb` が `Connection: close` を固定返却している**ため：

1. `wants_keep_alive?` はリクエストを見て `true` を返す（HTTP/1.1 デフォルト）
2. レスポンスの `Connection: close` を受け取ったクライアントが接続を閉じる
3. `serve_conn` がループして次を読もうとすると `ConnectionClosed` が発生
4. 実質 1 接続 1 リクエストで終わっている

---

## 変更するファイル

```
ruby/
└── main.rb    ← Connection ヘッダーを keep_alive に応じて切り替える
```

`server.rb`・`request.rb` は変更不要。

---

## 修正方針

`main.rb` のレスポンスヘッダーを、リクエストの keep_alive 意思に合わせて切り替える。

```ruby
# Before（固定）
"Connection: close"

# After（リクエストに応じて切り替え）
connection_header = req.wants_keep_alive? ? "keep-alive" : "close"
"Connection: #{connection_header}"
```

---

## Go との比較

Go では Step 6 で `handleConn` にリクエストループを追加した。  
Ruby は Step 5 の `serve_conn` 抽出時にループを先に実装したため、Step 6 の変更は `Connection` ヘッダーの切り替えのみ。

| | Go Step 6 | Ruby Step 6 |
|---|---|---|
| リクエストループ | Step 6 で追加 | Step 5 で実装済み |
| 変更点 | `handleConn` をループ構造に書き換え | `main.rb` の Connection ヘッダーを切り替えるだけ |
| タイムアウト | `conn.SetReadDeadline` | `IO.select` （発展課題）|

---

## タイムアウト（発展課題）

Keep-Alive 接続はクライアントが切るまでスレッドが生き続ける。無操作のまま放置されるとスレッドがリークする。

`IO.select` を使うと「一定時間データが来なければ切断」を実装できる。

```ruby
# serve_conn のループ内で socket.gets の前に挟む
ready = IO.select([socket], nil, nil, 30)
break if ready.nil?   # 30 秒待ってもデータが来なかった
```

Step 6 の本題ではないが、実用サーバーでは必須の考慮事項。

---

## 実装の確認手順

### ステップ 1: 起動確認

```bash
mise exec -- ruby main.rb
```

### ステップ 2: 単一リクエスト（既存動作の確認）

```bash
curl http://localhost:8080/
# → 200 OK "Hello, World!"
```

### ステップ 3: Keep-Alive で複数リクエスト

```bash
curl -v --http1.1 \
  http://localhost:8080/ \
  http://localhost:8080/about \
  http://localhost:8080/
```

`curl -v` の出力で `* Re-using existing connection` が表示されれば成功。

### ステップ 4: Connection: close で明示的に切断

```bash
curl -v -H "Connection: close" http://localhost:8080/
```

サーバーログに `Closed connection from` が出てスレッドが終了すること。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `Re-using existing connection` が出ない | レスポンスヘッダーがまだ `Connection: close` になっている | `wants_keep_alive?` の分岐を確認 |
| 2 つ目のリクエストが返らない | `serve_conn` のループが `break` している | `keep_alive` の値をログで確認 |
| スレッドが終了しない | タイムアウトがない | 発展課題として `IO.select` を追加する |
