# Ruby HTTP サーバー 設計方針

## Rack スタイルのハンドラ

Ruby 実装では、標準の Web サーバーインターフェース **Rack** と同じ考え方を採用する。

### 基本方針

ハンドラは「入力(req) → 出力([status, body])」の純粋な変換として定義する。I/O はサーバーが一手に担う。

```
ハンドラ = 入力(req) → 出力([status, body])   # I/Oなし
サーバー  = パース → ハンドラ呼び出し → 書き込み  # I/Oを管理
```

### ファイル構成と責務

| ファイル | 責務 |
|---|---|
| `main.rb` | ハンドラを定義してサーバーに渡す |
| `lib/server.rb` | accept → parse → handler.call → write |
| `lib/request.rb` | バイト列 → `Request` オブジェクトに変換 |

### ハンドラの形（main.rb）

```ruby
handler = ->(req) {
  case req.path
  when "/"      then ["200 OK", "Hello!"]
  when "/about" then ["200 OK", "About page"]
  else               ["404 Not Found", "Not Found"]
  end
}
```

### サーバー側の処理イメージ（server.rb）

```ruby
req = Request.parse(socket)
status, body = @handler.call(req)
socket.write("HTTP/1.1 #{status}\r\n...")
```

## Go との比較

| | Go | Ruby |
|---|---|---|
| ハンドラ引数 | `(req, writeResponse)` | `(req)` のみ |
| レスポンスの渡し方 | `writeResponse` proc を呼ぶ | 戻り値 `[status, body]` で返す |
| I/O の所在 | ハンドラが `writeResponse` 経由で書く | サーバーが一手に書く |

Ruby は配列での多値返却が自然なため、ハンドラを I/O から完全に切り離せる。  
結果としてハンドラが純粋な関数になり、テストも容易になる。
