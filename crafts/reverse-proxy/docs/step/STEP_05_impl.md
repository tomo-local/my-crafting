# Step 5 実装ガイド：Least Connections

## ゴール

レスポンスが遅いサーバーに振り分けが集中しないこと。

```bash
# サーバーA: 即レスポンス
go run upstream/main.go -port 9001 -id fast -delay 0ms &

# サーバーB: 500ms 遅延
go run upstream/main.go -port 9002 -id slow -delay 500ms &

go run main.go -upstreams localhost:9001,localhost:9002 -port 8080 -balancer lc

# 10並列でアクセス
for i in $(seq 1 10); do curl -s http://localhost:8080/ & done
wait
# → fast が大多数を処理し、slow は少数の処理にとどまること
```

---

## 変更するファイル

```
go/
├── main.go              ← -balancer フラグ追加、balancer 切り替え
├── upstream/
│   └── main.go          ← -delay フラグを追加
└── balancer/
    ├── roundrobin.go    ← Upstream に connections フィールドを追加
    └── leastconn.go     ← 新規作成
```

---

## 1. `upstream/main.go` への `-delay` 追加

`-delay` フラグ（`time.Duration` 型）を追加し、ハンドラ内で `time.Sleep(*delay)` を呼ぶ。

---

## 2. `balancer/roundrobin.go` の修正

> **Step 4 との差分**
> `Upstream` に `connections int64` フィールドを追加する。

```go
type Upstream struct {
    Addr        string
    alive       bool
    mu          sync.RWMutex
    connections int64  // ← 追加
}
```

`Connections()` アクセサを追加する。

```go
func (u *Upstream) Connections() int64 {
    return atomic.LoadInt64(&u.connections)
}

func (u *Upstream) IncrConnections() {
    atomic.AddInt64(&u.connections, 1)
}

func (u *Upstream) DecrConnections() {
    atomic.AddInt64(&u.connections, -1)
}
```

---

## 3. `balancer/leastconn.go`

**`LeastConn` 構造体と `Next()` メソッド:**

1. `LeastConn` 構造体に `upstreams []*Upstream` を持たせる
2. `New(upstreams []*Upstream) *LeastConn` コンストラクタを実装する
3. `(lc *LeastConn) Next() string` メソッドを実装する:

   ```go
   func (lc *LeastConn) Next() string {
       var candidate *Upstream
       minConn := int64(math.MaxInt64)

       for _, u := range lc.upstreams {
           if !u.IsAlive() {
               continue
           }
           c := u.Connections()
           if c < minConn {
               minConn = c
               candidate = u
           }
       }

       if candidate == nil {
           return ""
       }
       return candidate.Addr
   }
   ```

4. `(lc *LeastConn) Upstreams() []*Upstream` を返すメソッドを追加する（ヘルスチェックで使う）

---

## 4. `main.go` の修正

> **Step 4 との差分**
> `-balancer` フラグで `rr`（ラウンドロビン）と `lc`（Least Connections）を切り替えられるようにする。

**インターフェースを定義する:**

```go
type Balancer interface {
    Next() string
    Upstreams() []*balancer.Upstream
}
```

**`-balancer` フラグに応じて実装を切り替える:**

```go
var b Balancer
switch *balancerFlag {
case "lc":
    b = balancer.NewLeastConn(upstreams)
default: // "rr"
    b = balancer.NewRoundRobin(upstreams)
}
```

**`handleConn` でコネクション数を管理する:**

```go
func handleConn(client net.Conn, b Balancer) {
    defer client.Close()

    addr := b.Next()
    if addr == "" {
        // 503 を返す
        return
    }

    // Upstream を addr から引いてコネクション数を管理する
    upstream := findUpstream(b.Upstreams(), addr)
    upstream.IncrConnections()
    defer upstream.DecrConnections()

    // ... Step 2 と同じ転送処理 ...
}
```

> `findUpstream` は `addr` 文字列から `*Upstream` を線形検索で引く小さなヘルパー関数。

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: -race で競合がないこと

```bash
go run -race main.go -upstreams localhost:9001,localhost:9002 -balancer lc -port 8080
for i in $(seq 1 50); do curl -s http://localhost:8080/ & done
wait
# → DATA RACE の出力がないこと
```

### ステップ 3: 遅いサーバーへの集中がないことを確認

```bash
for i in $(seq 1 10); do curl -s http://localhost:8080/ & done; wait
# → fast が多く、slow が少ないこと
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| 全てのリクエストが同じサーバーに行く | `connections` のインクリメントが `defer` より先に来ていない | `IncrConnections()` を呼んだ直後に `defer DecrConnections()` を書く |
| コネクション数がマイナスになる | `DecrConnections()` が多重呼び出しされている | `defer` は1回だけ呼ばれることを確認 |
| `-race` でデータ競合が出る | `connections` を `atomic` 経由でなく直接読み書きしている | 必ず `atomic.LoadInt64` / `atomic.AddInt64` を使う |
| ラウンドロビンに比べてLCが遅い | 全アップストリームを毎回スキャンしている | 台数が少ない（10台以下）なら線形スキャンで十分。台数が多い場合はヒープ構造を検討 |
