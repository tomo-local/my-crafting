# Step 3：ラウンドロビン負荷分散（前提知識）

Step 2 までは転送先が1台固定でした。Step 3 では複数のアップストリームを管理し、リクエストを順番に振り分けるラウンドロビンを実装します。

---

## 1. ラウンドロビンのアルゴリズム

リクエストが届くたびにグローバルカウンターをインクリメントし、アップストリーム台数で割った余りをインデックスとして使います。

```
アップストリームリスト: [A, B, C]  (長さ 3)

1回目: counter=0 → 0 % 3 = 0 → A
2回目: counter=1 → 1 % 3 = 1 → B
3回目: counter=2 → 2 % 3 = 2 → C
4回目: counter=3 → 3 % 3 = 0 → A  （ループ）
```

---

## 2. goroutine-safe なカウンター

複数のリクエストが同時に来ると、複数の goroutine が同時にカウンターを読み書きします。通常の `int` 変数を複数goroutineから操作するとデータ競合が起きます。

### `sync/atomic` を使う理由

`atomic.AddUint64` は CPU の不可分命令（アトミック命令）を使うため、Mutex なしでgoroutine-safe にインクリメントできます。

```go
var counter uint64

// goroutine-safe なインクリメントと読み取り
next := atomic.AddUint64(&counter, 1)
index := next % uint64(len(upstreams))
```

`atomic.AddUint64` は「加算した後の値」を返すため、別途 `Load` する必要はありません。

---

## 3. アップストリームリストの設計

複数のアップストリームをどこで管理するかは設計上重要です。Step 3 では以下のシンプルな構造を使います。

```go
type RoundRobinBalancer struct {
    upstreams []string
    counter   uint64
}

func (b *RoundRobinBalancer) Next() string {
    n := atomic.AddUint64(&b.counter, 1)
    return b.upstreams[n % uint64(len(b.upstreams))]
}
```

---

## 4. コマンドラインでの複数アップストリーム指定

`-upstream` フラグをカンマ区切りで受け取り、`strings.Split` で分割します。

```bash
go run main.go -upstreams localhost:9001,localhost:9002,localhost:9003
```

---

## 📌 まとめ：Step 3 のフロー

1. 起動時に `-upstreams` フラグをパースして `[]string` に変換する
2. `RoundRobinBalancer` を初期化する
3. `Accept` でクライアントの接続を受け取る
4. `balancer.Next()` で次に使うアップストリームを選ぶ
5. Step 2 と同じ手順でヘッダーを書き換えて転送する
