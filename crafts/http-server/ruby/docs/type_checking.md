# Ruby 型チェック（RBS + Steep）

## 概要

このプロジェクトでは **RBS** で型定義を書き、**Steep** で静的型チェックを行う。

```bash
bundle exec steep check
```

---

## 仕組み

### RBS（Ruby Signature）

Ruby 3.0 から組み込まれた型定義の言語。`.rbs` ファイルに型情報を書く。
**実装コード（`.rb`）は一切変更しない**のが特徴。

```
ruby/
├── lib/
│   ├── request.rb       ← 実装（変更なし）
│   ├── response.rb
│   └── server.rb
├── sig/
│   ├── http_server.rbs  ← 型定義
│   └── main.rbs
└── Steepfile            ← Steep の設定
```

### Steep

RBS ファイルを読んで `.rb` のコードを静的解析するツール（gem）。
コンパイル時チェックではなく、実行前に型の整合性を検証する。

---

## ファイルの役割

### Steepfile

チェック対象のファイルと RBS の場所を指定する設定ファイル。

```ruby
target :lib do
  signature "sig"      # sig/ ディレクトリの RBS を使う

  check "lib"          # lib/*.rb を型チェック
  check "main.rb"

  library "socket"     # 標準ライブラリの型定義を読み込む
  library "logger"
end
```

`library` で標準ライブラリの型定義を有効にする。ここに書かないと `TCPSocket` や `Logger` が未知の型として扱われる。

### sig/http_server.rbs

クラスとメソッドの型シグネチャを定義する。

```rbs
module HttpServer
  class Request
    attr_reader method: String          # String 型の読み取り専用属性
    attr_reader content_length: Integer

    def self.parse: (TCPSocket socket) -> Request   # クラスメソッド
    def wants_keep_alive?: () -> bool               # 引数なし、bool を返す
  end

  class Server
    # ハンドラは (Request, Response) を受け取り void を返す Proc
    def initialize: (Integer port, ^(Request, Response) -> void handler) -> void
  end
end
```

`^(X, Y) -> Z` は Proc の型。`(X, Y) -> Z` はメソッドの型。

### sig/main.rbs

トップレベルの定数を宣言する。
トップレベルの `def` は RBS では書けないため、`Object` クラスに定義する。

```rbs
LOG: Logger   # 定数の型

class Object
  private
  def handle_conn: (HttpServer::Request req, HttpServer::Response res) -> void
end
```

---

## 今回検出されたバグ

Steep を導入した際に `request.rb` で実際のバグが2件検出された。

### ① `name.downcase` で nil の可能性

```ruby
# Before
name, value = header.split(":", 2)
next if value.nil?
case name.downcase   # ← Steep: name が nil の可能性
```

`a, b = array` の多重代入では、配列の要素数が足りない場合に変数が `nil` になる。
`split` の結果から `name` が実質 `nil` になることはないが、型システム上は `String | nil` と推論される。

```ruby
# After
next if name.nil? || value.nil?
```

### ② `socket.read` が nil を返す可能性

```ruby
# Before
body = content_length&.positive? ? socket.read(content_length) : ""
# ← Steep: socket.read は String | nil
```

`IO#read` は EOF 時に `nil` を返す仕様のため、`String | nil` が正しい型。

```ruby
# After
body = content_length.positive? ? (socket.read(content_length) || "") : ""
```

---

## RBS の主な構文

| 構文 | 意味 | 例 |
|---|---|---|
| `String` | String 型 | `attr_reader name: String` |
| `Integer \| nil` | Union 型（nil 許容） | `attr_reader count: Integer \| nil` |
| `bool` | true または false | `def ok?: () -> bool` |
| `void` | 戻り値なし | `def close: () -> void` |
| `^(X) -> Y` | Proc 型 | `^(Request, Response) -> void` |
| `self.method:` | クラスメソッド | `def self.parse: (Socket) -> Request` |
| `attr_reader name: T` | 読み取り専用属性 | `attr_reader body: String` |

---

## Method vs Proc の違い

`method(:handle_conn)` は `Method` オブジェクトを返すが、RBS の `^(...)` は `Proc` 型。
Steep はこれを型エラーとして検出する。

```ruby
# NG: Method 型は ^(...) に渡せない
HttpServer::Server.new(8080, method(:handle_conn))

# OK: lambda は Proc 型
HttpServer::Server.new(8080, ->(req, res) { handle_conn(req, res) })
```

---

## 実行方法

```bash
# 型チェック
bundle exec steep check

# 特定ファイルのみ
bundle exec steep check lib/request.rb
```
