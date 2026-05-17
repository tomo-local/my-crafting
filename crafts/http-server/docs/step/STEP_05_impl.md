# Step 5 実装ガイド：並行処理の導入（goroutine）

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
go/
└── main.go    ← main() の Accept ループを 1 行修正
```

---

## 1. `main()`

Accept ループの `handleConn(conn)` に `go` を付けるだけです。

```go
// Before
handleConn(conn)

// After
go handleConn(conn)
```

---

## 2. `handleConn(conn net.Conn)`

Step 4 から変更なし。

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: 直列との挙動の違いを体感する

遅延を入れたエンドポイントを一時的に追加して確認します。

```go
// handleConn 内にテスト用エンドポイントを追加
case path == "/slow":
    time.Sleep(3 * time.Second)
    status = "200 OK"
    body = "slow response"
```

```bash
# /slow を叩きながら / にもアクセスする
curl http://localhost:8080/slow &
curl http://localhost:8080/
# goroutine 版 → / はすぐ返る
# 直列版  → /slow が終わるまで / が待たされる
```

### ステップ 3: race detector で競合がないことを確認

```bash
go run -race main.go
```

警告が出なければ OK。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `go vet` や race detector が警告を出す | goroutine 内でループ変数を直接使っている | `conn` は引数で渡しているので通常は問題なし |
| サーバーが `Ctrl+C` で落ちない | goroutine がブロック中 | Step 5 の範囲では許容。graceful shutdown は発展課題 |
