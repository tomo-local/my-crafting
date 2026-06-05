# Step 5 実装ガイド（Ruby）：並行処理の導入（Thread）

## ゴール

複数の curl を同時に投げても詰まらずにレスポンスが返ること。

```bash
# 3 つ同時にリクエストを投げる
curl http://localhost:8080/ & curl http://localhost:8080/ & curl http://localhost:8080/
# → 3 つとも 200 OK が返ること（直列なら 2 つ目・3 つ目が待たされる）
```

---

## 変更するファイル

```
ruby/
└── lib/server.rb    ← listen_and_serve と serve_conn に分割
```

---

## 現状と問題点

変更前の `listen_and_serve` は直列処理かつ accept と接続処理が混在：

```
accept → parse → handler.call → accept → parse → handler.call → ...
```

`handler.call` が終わるまで次の `accept` に進めないため、同時接続があると後続が待たされる。

---

## 修正方針：`serve_conn` の抽出 + `Thread.new` による委譲

Go の `ServeConn` と同じ分離を行う。

- `listen_and_serve` → accept してスレッドに渡すだけ
- `serve_conn` → 1接続のライフサイクル全体（parse → handler → keep_alive ループ）

```ruby
# Before（直列・混在）
def listen_and_serve
  loop do
    socket = server.accept
    request = HttpServer::Request.parse(socket)
    keep_alive = request.wants_keep_alive?
    @handler.call(socket, request)
    socket.close unless keep_alive
  end
end

# After（並行・分離）
def listen_and_serve
  loop do
    socket = server.accept
    Thread.new(socket) { |sock| serve_conn(sock) }
  end
end

def serve_conn(socket)
  loop do
    request = HttpServer::Request.parse(socket)
    keep_alive = request.wants_keep_alive?
    @handler.call(socket, request)
    break unless keep_alive
  end
rescue => e
  LOG.error("Connection error: #{e.message}")
ensure
  socket.close
end
```

`Thread.new(socket)` とブロック引数 `|sock|` でソケットを渡す点がポイント。
ループ変数 `socket` をブロック内で直接参照すると、次の反復で上書きされた値を掴む可能性があるため、引数として渡して束縛する。

---

## Go との対応関係

| Go | Ruby |
|---|---|
| `ListenAndServe` | `listen_and_serve` |
| `go s.ServeConn(conn)` | `Thread.new(socket) { \|sock\| serve_conn(sock) }` |
| `ServeConn` | `serve_conn` |
| `defer conn.Close()` | `ensure socket.close` |

---

## Go との比較

| | Go | Ruby |
|---|---|---|
| 並行単位 | goroutine（軽量・M:N スレッド） | Thread（OS スレッド、1:1） |
| 起動構文 | `go s.ServeConn(conn)` | `Thread.new(socket) { \|sock\| serve_conn(sock) }` |
| 変数の受け渡し | 引数で渡す | `Thread.new(socket)` + ブロック引数で渡す |
| 例外処理 | recover() | rescue で明示的に捕捉する必要あり |

Ruby の Thread は OS スレッドに 1:1 対応するため、goroutine より生成コストが高い。
大量接続には Thread Pool（`thread_pool` gem など）が必要になるが、学習目的では `Thread.new` で十分。

---

## 例外処理

スレッド内で rescue しないと、例外はデフォルトで**サイレントに無視**される（Ruby 3.x でも同様）。
`serve_conn` 内に rescue を置くことで、接続エラーやパースエラーをログに残せる。
`ensure` でのクローズは例外発生時も確実にソケットを閉じるために必要。

---

## 実装の確認手順

### ステップ 1: 起動確認

```bash
mise exec -- ruby main.rb
```

### ステップ 2: 単純なリクエスト（既存動作の確認）

```bash
curl http://localhost:8080/
# → 200 OK "Hello, World!"
```

### ステップ 3: 並行処理の確認

スロー応答を一時的に `main.rb` のハンドラに追加して動作差を体感する：

```ruby
# main.rb の handle_conn に一時追加
when req.method == "GET" && req.path == "/slow"
  sleep(3)
  ["200 OK", "slow response"]
```

```bash
# /slow を叩きながら / にもアクセス
curl http://localhost:8080/slow &
curl http://localhost:8080/
# Thread 版 → / はすぐ返る
# 直列版   → /slow が終わるまで / が待たされる
```

確認後は `/slow` の分岐を削除する。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| スレッド内のエラーが出力されない | rescue がない | スレッド内に `rescue => e` を追加 |
| ソケットが閉じられないまま残る | 例外発生時に `close` が呼ばれない | `ensure` で `sock.close` |
| ループ変数を直接参照している | 次イテレーションで `socket` が上書きされる | `Thread.new(socket) \|sock\|` で引数渡し |
