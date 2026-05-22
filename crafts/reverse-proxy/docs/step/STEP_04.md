# Step 4：ヘルスチェック（前提知識）

Step 3 ではアップストリームが落ちても気づかずリクエストを送り続けます。Step 4 ではバックグラウンドで定期的にアップストリームの死活を確認し、落ちたサーバーを自動的に迂回します。

---

## 1. アクティブヘルスチェックの仕組み

プロキシ起動時にバックグラウンドgoroutineを起動し、一定間隔（例: 10秒ごと）で各アップストリームに HTTP リクエストを送ります。

```
goroutine: 10秒ごとに全アップストリームへ GET /health
    → 200 OK かつタイムアウト内: alive = true
    → 接続エラー or タイムアウト or 5xx: alive = false
```

ヘルスチェックに使うエンドポイントは `/health` が慣習ですが、疎通確認だけなら `HEAD /` でも構いません。

---

## 2. サーバー状態の管理

各アップストリームの状態（alive/dead）を管理する構造体が必要です。

```go
type Upstream struct {
    addr  string
    alive bool
}
```

`balancer.Next()` でアップストリームを選ぶとき、`alive == false` のサーバーをスキップするよう変更します。

---

## 3. `sync.RWMutex` による排他制御

ヘルスチェックgoroutineが `alive` を**書き換え**、リクエスト転送goroutineが `alive` を**読む**という競合が発生します。

`sync.RWMutex` は「読み取りは複数goroutineが同時にOK、書き込みは排他」という特性を持ちます。

```
読み取り（振り分け選択）: RLock() / RUnlock()
書き込み（ヘルスチェック更新）: Lock() / Unlock()
```

通常の `sync.Mutex` でも正しく動きますが、読み取りが多く書き込みが少い場合は `RWMutex` の方が並列性が高くなります。

---

## 4. 全台 dead のフォールバック

全アップストリームが dead の場合、リクエストに対して 503 Service Unavailable を返します。

```
HTTP/1.1 503 Service Unavailable\r\n
Content-Length: 19\r\n
\r\n
Service Unavailable
```

この場合はアップストリームへの接続を試みず、すぐにクライアントに返します。

---

## 5. ヘルスチェックの間隔設計

| 間隔 | メリット | デメリット |
|---|---|---|
| 短い（1〜2秒） | 障害の検知が速い | アップストリームへの余分なトラフィックが増える |
| 長い（30秒〜） | トラフィックが少ない | 障害検知まで時間がかかる |

実装では `-health-interval` フラグで設定可能にするのがベストです。デフォルト 10 秒が一般的です。

---

## 📌 まとめ：Step 4 のフロー

**起動時:**
1. 各アップストリームを `alive = true` で初期化する
2. `startHealthCheck(upstreams, interval)` でバックグラウンドgoroutineを起動する

**ヘルスチェックgoroutine（定期実行）:**
1. `time.Ticker` で指定間隔ごとにトリガーする
2. 全アップストリームに対して `http.Get("http://addr/health")` を送る
3. 結果に応じて `alive` を Lock 付きで更新する

**リクエスト転送時:**
1. `RLock` で `alive == true` のアップストリームだけを候補に挙げる
2. 候補がなければ 503 を返す
3. 候補からラウンドロビンで1台選んで転送する
