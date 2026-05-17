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
