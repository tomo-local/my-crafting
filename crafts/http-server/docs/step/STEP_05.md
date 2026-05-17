# Step 5：並行処理の導入（前提知識）

Step 4 まではリクエストを 1 つずつ直列に処理していました。1 つのリクエストを処理している間は次の接続が待たされます。Step 5 では **goroutine** を使って複数の接続を同時に処理できるようにします。

---

## 1. 問題：直列処理のボトルネック

現在の Accept ループは次のような流れです。

```
Accept() → handleConn() → Accept() → handleConn() → ...
```

`handleConn` が終わるまで次の `Accept` に進めません。接続が 2 つ同時に来た場合、2 つ目は 1 つ目が終わるまでキューで待たされます。

---

## 2. goroutine とは

Go の軽量な並行実行単位です。OS スレッドとは異なり、数千〜数万起動してもメモリ・スケジューリングコストが低く抑えられます。

```go
go someFunction() // goroutine として非同期に実行
```

`go` キーワードを付けるだけで、その関数は呼び出し元と**並行して**実行されます。呼び出し元はすぐに次の行へ進みます。

---

## 3. 修正は 1 行だけ

Accept ループの中で `handleConn(conn)` を `go handleConn(conn)` に変えるだけです。

```go
// Before（直列）
for {
    conn, _ := listener.Accept()
    handleConn(conn)
}

// After（並行）
for {
    conn, _ := listener.Accept()
    go handleConn(conn) // すぐ次の Accept へ進む
}
```

これだけで、`handleConn` が実行中でも次の接続を `Accept` できるようになります。

---

## 4. goroutine を使うときの注意点

### `conn` をクロージャでキャプチャする場合

ループ変数を goroutine 内で使う場合、クロージャが最新の値を参照するため意図しない動きになることがあります。`handleConn(conn)` のように引数として渡す形では問題ありません。

### 共有リソースへのアクセス

複数の goroutine が同じ変数を読み書きする場合はデータ競合が起きます。Step 5 では各 goroutine が独立した `conn` を持つだけなので競合は発生しません。グローバルなカウンター等を追加する場合は `sync.Mutex` や `sync/atomic` が必要になります。

---

## 📌 まとめ：直列 vs 並行の比較

| | 直列（Step 1–4） | 並行（Step 5） |
|---|---|---|
| 同時接続 | 1 つずつ | 無制限 |
| コード変更量 | — | 1 行（`go` キーワード） |
| データ競合リスク | なし | 共有状態があれば発生しうる |
