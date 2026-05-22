# Step 4 実装ガイド：ヘルスチェック

## ゴール

落ちたアップストリームへのリクエストが自動で迂回されること。

```bash
# 2台起動
go run upstream/main.go -port 9001 -id upstream-1 &
go run upstream/main.go -port 9002 -id upstream-2 &

go run main.go -upstreams localhost:9001,localhost:9002 -port 8080

# 正常時: 2台に振り分けられる
curl http://localhost:8080/  # upstream-1 or upstream-2

# upstream-1 を停止
kill <upstream-1-pid>

# ヘルスチェック間隔（デフォルト10秒）待ってから確認
sleep 12
curl http://localhost:8080/  # → upstream-2 のみに届くこと（upstream-1 には届かない）

# 全台停止
kill <upstream-2-pid>
curl http://localhost:8080/
# → HTTP/1.1 503 Service Unavailable
```

---

## 変更するファイル

```
go/
├── main.go
├── upstream/
│   └── main.go        ← /health エンドポイントを追加
└── balancer/
    ├── roundrobin.go  ← Upstream 構造体追加・alive チェック追加
    └── health.go      ← 新規作成
```

---

## 1. `upstream/main.go` への `/health` 追加

`/health` エンドポイントを追加して 200 を返すだけ。

```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
})
```

---

## 2. `balancer/roundrobin.go` の修正

> **Step 3 との差分**
> `[]string` から `[]*Upstream` に変更し、`alive` チェックを追加する。

**`Upstream` 構造体を追加:**

```go
type Upstream struct {
    Addr  string
    alive bool
    mu    sync.RWMutex
}

func (u *Upstream) SetAlive(alive bool) {
    u.mu.Lock()
    defer u.mu.Unlock()
    u.alive = alive
}

func (u *Upstream) IsAlive() bool {
    u.mu.RLock()
    defer u.mu.RUnlock()
    return u.alive
}
```

**`RoundRobin.Next()` の修正:**

1. 最大 `len(upstreams)` 回ループする
2. カウンターをインクリメントしてインデックスを計算する
3. 対象の `IsAlive()` が `true` なら `Addr` を返す
4. 全てのアップストリームが dead なら `""` を返す

```go
func (r *RoundRobin) Next() string {
    for i := 0; i < len(r.upstreams); i++ {
        n := atomic.AddUint64(&r.counter, 1)
        u := r.upstreams[n%uint64(len(r.upstreams))]
        if u.IsAlive() {
            return u.Addr
        }
    }
    return ""
}
```

**`RoundRobin.Upstreams()` を追加:** ヘルスチェックgoroutineが全台をイテレートできるよう `[]*Upstream` を返すメソッドを追加する。

---

## 3. `balancer/health.go`

**`StartHealthCheck(upstreams []*Upstream, interval time.Duration)` 関数:**

1. goroutine を起動する
2. goroutine 内で `time.NewTicker(interval)` を作る
3. `for range ticker.C { ... }` でティックのたびに全台チェックする
4. チェック処理 `checkUpstream(u *Upstream)`:
   - `http.Get("http://" + u.Addr + "/health")` をタイムアウト付きで送る（`http.Client{Timeout: 3*time.Second}`）
   - 200 系なら `u.SetAlive(true)`
   - エラーや 5xx なら `u.SetAlive(false)`
   - ステータス変化があった場合はログに出力する

---

## 4. `main.go` の修正

> **Step 3 との差分**
> `-health-interval` フラグを追加し、起動時に `StartHealthCheck` を呼ぶ。

1. `-health-interval` フラグを追加する（`time.Duration` 型、デフォルト `10s`）
2. `balancer.New(upstreams)` 呼び出し後に `balancer.StartHealthCheck(b.Upstreams(), *healthInterval)` を呼ぶ
3. `handleConn` 内で `b.Next()` が `""` を返したとき 503 レスポンスを返す

```go
upstream := b.Next()
if upstream == "" {
    fmt.Fprintf(client, "HTTP/1.1 503 Service Unavailable\r\nContent-Length: 19\r\n\r\nService Unavailable")
    return
}
```

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: ヘルスチェックのログ確認

起動後、ログに `health check: upstream-1 alive` のような出力が出ること。

### ステップ 3: フェイルオーバーの確認

```bash
# アップストリームを停止
kill <upstream-1-pid>

# ヘルスチェック間隔後にログを確認
# → health check: upstream-1 dead  のようなログが出ること

# アクセスして upstream-2 のみに届くことを確認
for i in $(seq 1 5); do curl -s http://localhost:8080/; done
# → upstream-2 のレスポンスのみ
```

### ステップ 4: 503 の確認

```bash
# 全台停止
curl -v http://localhost:8080/
# → HTTP/1.1 503 Service Unavailable
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| 落ちたサーバーにもリクエストが届く | ヘルスチェックgoroutineが起動していない | `StartHealthCheck` の呼び出し漏れを確認 |
| ヘルスチェックがすぐにタイムアウトする | アップストリームに `/health` エンドポイントがない | upstream に `/health` を追加する |
| 全台 dead 判定になってしまう | ヘルスチェックのタイムアウトが短すぎる | `http.Client{Timeout}` の値を増やす |
| データ競合が発生する | `alive` フィールドを直接読み書きしている | `SetAlive` / `IsAlive` を必ず通す |
| サーバー復活後も dead のまま | ヘルスチェックが alive に戻す処理がない | チェック成功時に `SetAlive(true)` を呼んでいるか確認 |
