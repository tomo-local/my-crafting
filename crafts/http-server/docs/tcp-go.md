# Go の `net` パッケージと TCP ソケット

`study.md` / `STEP_01.md` で解説した OS レベルの概念が、Go の `net` パッケージのどの API に対応するかをまとめます。

---

## 対応表

| OS レベルの操作 | Go の API | 戻り値 |
|---|---|---|
| socket + bind + listen | `net.Listen("tcp", ":8080")` | `net.Listener, error` |
| accept（ブロッキング） | `listener.Accept()` | `net.Conn, error` |
| read | `conn.Read(buf)` | `n int, error` |
| write | `conn.Write(buf)` | `n int, error` |
| close | `conn.Close()` | `error` |

> `net.Listen` は socket / bind / listen の 3 ステップを 1 呼び出しで済ませます。

---

## net.Listener

```go
listener, err := net.Listen("tcp", ":8080")
```

- ポート 8080 を OS に「このプロセスへ回してくれ」と登録する
- `listener` は接続の受付窓口。それ自体ではデータの読み書きをしない
- プロセス終了時に忘れずに `defer listener.Close()`

---

## net.Conn — 接続 1 本を表す型

`listener.Accept()` が返す `net.Conn` は、クライアントとの 1 対 1 の専用ソケットです。

```go
conn, err := listener.Accept() // 誰かが来るまでここでブロック
```

`net.Conn` は `io.Reader` と `io.Writer` の両方を満たすため、標準ライブラリの読み書きユーティリティをそのまま使えます。

```go
// io.Reader として使う例
buf := make([]byte, 4096)
n, err := conn.Read(buf) // OS の受信バッファからコピー
data := buf[:n]

// io.Writer として使う例
conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
```

---

## ストリームと bufio.Reader

TCP はデータの「区切り」を持たないストリームです（詳細は `study.md` §3）。
`conn.Read` は **1 回で全データが取れるとは限りません**。

そのため HTTP ヘッダーの読み取りには `bufio.Reader` を使い、`\r\n\r\n` を区切りとして読み進めるのが定石です。

```go
reader := bufio.NewReader(conn)

// \n まで読む（ストリームから 1 行ずつ取り出す）
line, err := reader.ReadString('\n')
```

`bufio.Reader` は内部バッファを持ち、複数回の `conn.Read` を透過的にまとめてくれます。

---

## Accept ループの基本骨格

```go
listener, _ := net.Listen("tcp", ":8080")
defer listener.Close()

for {
    conn, err := listener.Accept()
    if err != nil {
        // listener が閉じられたらループを抜ける
        break
    }
    handleConn(conn) // ← ここを go handleConn(conn) にすると並行処理（Step 5）
}
```

`handleConn` の中で必ず `defer conn.Close()` を呼ぶことで、処理が終わったソケットが OS に返却されます。

---

## エラーハンドリングの要点

| 状況 | `conn.Read` の戻り値 |
|---|---|
| 正常にデータを受信 | `n > 0, err == nil` |
| クライアントが接続を切った | `n == 0, err == io.EOF` |
| ネットワークエラー | `n == 0, err != nil`（`io.EOF` 以外） |

`err == io.EOF` はエラーではなく「相手が Close した」という正常な終了信号です。`io.EOF` を受け取ったらサーバー側も `conn.Close()` してループを終了します。

---

## 次のステップ

- **Step 1** (`STEP_01.md`): この骨格を動かして `nc` でキャッチボールする
- **Step 2**: `conn.Write` で HTTP レスポンスを返す
- **Step 5**: `handleConn` を goroutine に渡して並行処理を導入する

---

## HTTPサーバーの処理を層に分けて考える

`handleConn` の中には複数の「仕事」が混在している。それぞれが **何を受け取って何を返すか** に注目すると、自然に層として分離できる。

### データの変換チェーン

```
[]byte（生バイト）
  └─[パース]──→ Request{ Method, Path, Version }
                  └─[ルーティング]──→ (status, body string)
                                        └─[レスポンス組み立て]──→ []byte（送信データ）
```

各変換が独立しているため、「どこまでを誰が知るべきか」で切り出す場所が決まる。

---

### 各層の責務と「知っていること」

| 層 | 入力 | 出力 | 知っていること |
|---|---|---|---|
| パース | `[]byte` | `Request` | HTTPプロトコルの書式（`METHOD PATH VERSION`） |
| ルーティング | `Request` | `(status, body)` | アプリのURL設計（`/` → 200, `/about` → 200 など） |
| レスポンス組み立て | `(status, body)` | `[]byte` | HTTPレスポンスの書式（`HTTP/1.1 ...`） |
| 接続制御 | `net.Conn` | — | TCP接続のライフサイクル（読む・書く・閉じる） |

**ポイント：** 「ルーティング」だけがアプリ固有の知識を持つ。他の3層はHTTPプロトコルやTCPの知識であり、アプリが変わっても変化しない。

---

### どこを切り出すかの判断基準

> **「この処理は、別のHTTPアプリを作ったときも再利用できるか？」**

- パース → YES：どんなHTTPサーバーでも同じ書式を読む
- ルーティング → NO：アプリごとにURLの設計が違う
- レスポンス組み立て → YES：HTTP/1.1の書式は共通
- 接続制御（`for` ループ） → YES：read → handle → write のループはどのHTTPサーバーも同じ

再利用できる処理 = `request` パッケージや `server` パッケージに切り出す候補。
再利用できない処理（ルーティング）= `main.go` 側に残す、または外から注入する。

---

### インターフェースを使う理由

ルーティングは「再利用できないが、毎回必ず必要」な処理。
この矛盾を解消するのがインターフェースの役割。

```go
// request パッケージ側：「ルーティングしてくれる何か」が必要だと宣言するだけ
type Router interface {
    Route(req Request) (status, body string)
}
```

```go
// main.go 側：アプリ固有のルーティングを実装して渡す
type myRouter struct{}
func (r *myRouter) Route(req request.Request) (string, string) { ... }
```

`request` パッケージはルーティングの**詳細を知らずに呼び出せる**。
これにより「再利用できる処理」と「アプリ固有の処理」が混ざらなくなる。

---

---

## 実装の補助線：response / request / server の分割

### パッケージの依存関係

```
main   → server, request, response
server → request, response
request → (stdlib のみ)
response → net (stdlib)
```

### 各パッケージが公開するもの

**`request` パッケージ**

```go
type Request struct { Method, Path, Version string }

// server から呼ばれるので公開が必要
func Parse(buf []byte) (Request, error)
```

**`response` パッケージ**

```go
// net.Conn をラップして HTTP レスポンスの書き込みを担う
type ResponseWriter struct { /* net.Conn を持つ */ }

func NewResponseWriter(conn net.Conn) *ResponseWriter
func (w *ResponseWriter) Write(status, body string) error
```

**`server` パッケージ**

```go
// Handler の型を変更：生の conn から パース済みの型 へ
type Handler func(w *response.ResponseWriter, r *request.Request)

// for ループが server 内部に移動する
func (s *Server) serveConn(conn net.Conn) {
    // 1. conn.Read でバイト列を読む
    // 2. request.Parse を呼ぶ
    // 3. response.NewResponseWriter を作る
    // 4. s.handler(w, &req) を呼ぶ
}
```

**`main.go`**

```go
// ルーティングロジックだけが残る
srv := server.NewServer(":8080", func(w *response.ResponseWriter, r *request.Request) {
    switch r.Path {
    case "/":
        w.Write("200 OK", "Welcome!")
    // ...
    }
})
```

---

### net/http との対応

Go 標準の `net/http` も同じ考え方で設計されている。

```
net.Conn（生バイト）
  └─[net/http 内部でパース]──→ *http.Request
                                  └─[Handler.ServeHTTP を呼ぶ]──→ ResponseWriter に書く
```

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

`ServeHTTP` がこのプロジェクトの `Route` に相当する。
標準ライブラリもパース・接続制御を内部に隠し、アプリ側には `Handler` インターフェースだけを要求している。
