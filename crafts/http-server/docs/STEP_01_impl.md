# Step 1 実装ガイド：TCPソケットで文字を受け取る

## ゴール

`nc localhost 8080` で接続し、入力した文字がサーバー側のコンソールに表示されること。
この Step では **`Write`（レスポンス送信）は行わない**。

```
# ターミナル A: サーバー起動
go run main.go

# ターミナル B: nc で接続
nc localhost 8080
hello    ← 入力
world    ← 入力
^C       ← Ctrl+C で切断

# ターミナル A に表示されること
[127.0.0.1:xxxxx] hello
[127.0.0.1:xxxxx] world
[127.0.0.1:xxxxx] connection closed
```

---

## 変更するファイル

```
go/
└── main.go    ← すべてここに書く
```

---

## 1. `main()`

リスナーを起動し、Accept ループを回すエントリポイント。

**内部でやること（順番どおり）:**

1. `net.Listen("tcp", ":8080")` でリスナーを作る
2. エラーがあれば `log.Fatal(err)` で終了する
3. `defer listener.Close()`
4. 起動ログを出力する（例: `Listening on :8080`）
5. 無限ループ `for { ... }` を開始する
6. `listener.Accept()` を呼ぶ（誰かが来るまでここでブロック）
7. Accept でエラーが返ったら `log.Fatal(err)` でプロセスを終了する
8. `handleConn(conn)` を呼ぶ

> **Step 5 との差分**
> Step 5 では `8` を `go handleConn(conn)` に変える。
> Step 1 では goroutine を使わず、1 接続ずつ直列に処理する。

---

## 2. `handleConn(conn net.Conn)`

1 本の接続に対して読み取り → 出力 → 切断を担う関数。`main` と同じファイルに書く。

**内部でやること（順番どおり）:**

1. `defer conn.Close()`
2. 接続元アドレスをログ出力（`conn.RemoteAddr()`）
3. バッファ `buf := make([]byte, 4096)` を用意する
4. 読み取りループ `for { ... }` を開始する
5. `n, err := conn.Read(buf)` を呼ぶ
6. `n > 0` であれば `buf[:n]` を文字列に変換して標準出力に表示する
7. `err == io.EOF` なら「接続が閉じられた」として break する
8. `err != nil`（EOF 以外）なら エラーログを出力して break する

**ループの流れ図:**

```
conn.Read(buf)
    │
    ├─ n > 0 かつ err == nil  → 表示してループ継続
    ├─ n > 0 かつ err == io.EOF → 表示してから break
    ├─ err == io.EOF          → break（接続終了）
    └─ err != nil             → エラーログ → break
```

> **注意: `err` と `n` は同時に評価する**
> TCP はストリームなので、最後のデータと `io.EOF` が同時に返ることがあります。
> `err` だけ見て先に break すると、そのデータが捨てられます。
> 必ず「`n > 0` なら先に表示、その後 `err` を確認する」順番で書いてください。

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

エラーなしでバイナリが生成されれば OK。

### ステップ 2: nc で接続して表示されること

```bash
# ターミナル A
go run main.go
# → Listening on :8080

# ターミナル B
nc localhost 8080
hello
world
# Ctrl+C で切断

# ターミナル A の期待出力
Listening on :8080
[127.0.0.1:xxxxx] connected
hello
world
[127.0.0.1:xxxxx] connection closed
```

### ステップ 3: 2 回目の接続も受け付けること

nc を切断した後、再度 `nc localhost 8080` して接続できれば Accept ループが正しく動いている。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `bind: address already in use` | 前のプロセスがまだポートを使っている | `lsof -i :8080` でプロセスを確認して終了する |
| nc を切断してもサーバーが止まる | ループ内でエラーを `return` してしまっている | `handleConn` を `return` ではなく `break` で抜ける |
| 最後の1行が表示されない | `io.EOF` で先に break して `buf[:n]` を表示し忘れている | EOF チェックの前に `n > 0` の表示処理を書く |
| 2回目の接続が受け付けられない | `Accept` ループの外で `conn.Close()` を呼んでいる | `defer conn.Close()` は `handleConn` の先頭に置く |
