# Step 3 実装ガイド：ラウンドロビン負荷分散

## ゴール

3台のアップストリームにリクエストが均等に振り分けられること。

```bash
# 3台のアップストリームを起動
go run upstream/main.go -port 9001 -id upstream-1 &
go run upstream/main.go -port 9002 -id upstream-2 &
go run upstream/main.go -port 9003 -id upstream-3 &

# プロキシを起動
go run main.go -upstreams localhost:9001,localhost:9002,localhost:9003 -port 8080

# 3回アクセス
for i in 1 2 3; do curl -s http://localhost:8080/; done
# → upstream-1, upstream-2, upstream-3 の順で返ること
```

---

## 変更するファイル

```
go/
├── main.go              ← フラグ変更 + balancer 組み込み
└── balancer/
    └── roundrobin.go    ← 新規作成
```

---

## 1. `balancer/roundrobin.go`

**内部でやること（順番どおり）:**

1. `package balancer` で宣言する
2. `RoundRobin` 構造体を定義する

   ```go
   type RoundRobin struct {
       upstreams []string
       counter   uint64
   }
   ```

3. `New(upstreams []string) *RoundRobin` コンストラクタを実装する
4. `(r *RoundRobin) Next() string` メソッドを実装する

   ```go
   func (r *RoundRobin) Next() string {
       n := atomic.AddUint64(&r.counter, 1)
       return r.upstreams[n%uint64(len(r.upstreams))]
   }
   ```

---

## 2. `main.go` のフラグ変更

> **Step 2 との差分**
> `-upstream`（単数）を `-upstreams`（複数、カンマ区切り）に変える。

**内部でやること（順番どおり）:**

1. `-upstream` フラグを `-upstreams` に変更する
2. `flag.Parse()` 後に `strings.Split(*upstreamsFlag, ",")` でスライスに変換する
3. スライスの長さが 0 なら `log.Fatal("no upstreams specified")` で終了する
4. `balancer.New(upstreams)` で `RoundRobin` を初期化する

---

## 3. `handleConn` のシグネチャ変更

> **Step 2 との差分**
> `upstream string` の代わりに `balancer *balancer.RoundRobin` を受け取る。

```go
func handleConn(client net.Conn, b *balancer.RoundRobin) {
    upstream := b.Next()
    // ... 以降は Step 2 と同じ
}
```

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: 振り分けの確認

```bash
for i in $(seq 1 9); do curl -s http://localhost:8080/; done
```

期待する出力（順序は upstream-1, 2, 3 のローテーション）:
```
Hello from upstream-1
Hello from upstream-2
Hello from upstream-3
Hello from upstream-1
Hello from upstream-2
Hello from upstream-3
...
```

### ステップ 3: データ競合がないことの確認

```bash
go run -race main.go -upstreams localhost:9001,localhost:9002,localhost:9003

# 別ターミナルで並列アクセス
for i in $(seq 1 50); do curl -s http://localhost:8080/ & done
wait
# → `DATA RACE` の出力がないこと
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| 常に同じアップストリームに振り分けられる | `counter` をインクリメントし忘れている or `counter` が0始まりで `n%len` が常に同じ | `atomic.AddUint64` の戻り値（加算後の値）を使っているか確認 |
| `index out of range` パニック | `upstreams` が空スライスの状態で `Next()` を呼んでいる | 起動時にスライスの長さチェックを追加する |
| `-race` でデータ競合が検出される | `counter` を `atomic` ではなく通常の `++` でインクリメントしている | `atomic.AddUint64(&r.counter, 1)` を使う |
| 振り分けが均等にならない | `uint64` のオーバーフロー後の剰余計算が狂う | `uint64` の最大値 `1.8×10^19` を超えることは実際にはないので気にしなくてよい |
