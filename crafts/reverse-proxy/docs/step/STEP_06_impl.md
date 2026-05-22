# Step 6 実装ガイド：コネクションプーリング

## ゴール

100リクエストの合計時間がプーリングなしより短くなること。

```bash
# プーリングなし（Step 5 まで）
time for i in $(seq 1 100); do curl -s http://localhost:8080/ > /dev/null; done

# プーリングあり（Step 6）
go run main.go -upstreams localhost:9001 -port 8080 -pool-size 10
time for i in $(seq 1 100); do curl -s http://localhost:8080/ > /dev/null; done
# → プーリングありの方が速いこと
```

---

## 変更するファイル

```
go/
├── main.go
└── pool/
    └── connpool.go     ← 新規作成
```

---

## 1. `pool/connpool.go`

**`ConnPool` 構造体:**

```go
type ConnPool struct {
    addr    string
    idle    chan net.Conn
    maxIdle int
}
```

**`New(addr string, maxIdle int) *ConnPool`:**

1. `idle: make(chan net.Conn, maxIdle)` でチャネルを初期化する
2. 構造体を返す

**`(p *ConnPool) Get() (net.Conn, error)`:**

1. `select` で `idle` チャネルから取り出しを試みる
2. 取り出せたら `conn` を返す（ノンブロッキングで試みる、`default` に `net.Dial` を書く）

   ```go
   select {
   case conn := <-p.idle:
       return conn, nil
   default:
       return net.Dial("tcp", p.addr)
   }
   ```

**`(p *ConnPool) Put(conn net.Conn)`:**

1. `select` で `idle` チャネルへ返却を試みる
2. 満杯なら（`default` に入ったら）`conn.Close()` する

   ```go
   select {
   case p.idle <- conn:
   default:
       conn.Close()
   }
   ```

---

## 2. アップストリームごとのプール管理

各アップストリームに対して1つの `ConnPool` を持ちます。`Upstream` 構造体にフィールドを追加するか、`map[string]*pool.ConnPool` で管理します。

`Upstream` 構造体に追加する場合:

```go
type Upstream struct {
    Addr        string
    alive       bool
    mu          sync.RWMutex
    connections int64
    Pool        *pool.ConnPool  // ← 追加
}
```

`balancer.New` 内で `pool.New(addr, maxIdle)` を呼んで初期化する。

---

## 3. `handleConn` の修正

> **Step 5 との差分**
> `net.Dial` の代わりに `upstream.Pool.Get()` を使う。完了後は `Pool.Put()` で返却する。

**内部でやること（順番どおり）:**

1. `upstream := selectUpstream(b)` で転送先を選ぶ
2. `upstreamConn, err := upstream.Pool.Get()` で接続を取得する
3. `req.Write(upstreamConn)` で送信する。エラーが返った場合:
   - `upstreamConn.Close()` する（使えない接続は捨てる）
   - `net.Dial` で新規接続を確立してリトライする（1回だけ）
4. `io.Copy(client, upstreamConn)` でレスポンスを転送する
5. レスポンスの `Connection` ヘッダーを確認する

   > レスポンスヘッダーのパースは `http.ReadResponse(bufio.NewReader(upstreamConn), req)` を使う。ただし、レスポンスボディを読み切ってからでないとコネクションを再利用できない点に注意。

6. `Connection: close` なら `upstreamConn.Close()`、そうでなければ `upstream.Pool.Put(upstreamConn)` する

---

## 4. `main.go` への `-pool-size` フラグ追加

```go
poolSize := flag.Int("pool-size", 10, "max idle connections per upstream")
```

`balancer.New` 呼び出し時に `poolSize` を渡す。

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: 接続が再利用されていることを確認

`tcpdump` でポート9001への接続確立（SYN）の回数を数える。プーリングなしなら100リクエストで100回の SYN が見えるが、プーリングありなら数回のみのはず。

```bash
sudo tcpdump -i lo 'port 9001 and tcp[tcpflags] & tcp-syn != 0' -c 20
# 別ターミナルで
for i in $(seq 1 100); do curl -s http://localhost:8080/ > /dev/null; done
```

### ステップ 3: 速度比較

```bash
time for i in $(seq 1 100); do curl -s http://localhost:8080/ > /dev/null; done
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| 2回目以降のリクエストでエラーになる | レスポンスボディを読み切らずにコネクションを返却している | `io.ReadAll(resp.Body)` で全て読んでから `Pool.Put` する |
| `connection reset by peer` が頻発する | アップストリームが接続をclose済みの接続に書き込んでいる | `req.Write` エラー時に新規接続へのリトライを実装する |
| プールから取り出した接続がすぐ切れる | `idle` 接続のTTLがない、アップストリームのkeepaliveタイムアウトを超えた | `ConnPool` に `maxIdleTime` を設け、期限切れ接続は `Get` 時に捨てて新規確立する |
| メモリが増え続ける | エラー時に `upstreamConn.Close()` を呼ばずに `Put` している | エラー後の接続は必ず `Close` する |
