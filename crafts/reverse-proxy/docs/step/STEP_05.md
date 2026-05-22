# Step 5：Least Connections（前提知識）

ラウンドロビンはリクエストを均等に振り分けますが、処理時間が不均一な場合は問題が起きます。重いリクエストが特定のサーバーに固まってしまうと、そのサーバーへのリクエストが溜まり続けます。Step 5 では現在のアクティブコネクション数が最も少ないサーバーを選ぶ Least Connections を実装します。

---

## 1. ラウンドロビンとの違い

```
サーバーA: 軽いリクエストを処理 → すぐ終わる
サーバーB: 重いリクエストを処理 → なかなか終わらない

ラウンドロビン:
  → A, B, A, B, A, B と均等に割り当て
  → B は処理中のリクエストが溜まり続ける（キューが伸びる）

Least Connections:
  → A が空いているのでA, A, A, A と集中
  → B の処理が終わったら B, A, B, A とバランスされる
```

---

## 2. アクティブコネクション数のカウント

各アップストリームのアクティブコネクション数をカウンターで管理します。

```
リクエスト受付: connections++ （インクリメント）
レスポンス完了: connections-- （デクリメント）
```

`defer` を使うとリクエスト完了時のデクリメントを確実に行えます。

```go
upstream := selectLeastConnections(upstreams)
atomic.AddInt64(&upstream.connections, 1)
defer atomic.AddInt64(&upstream.connections, -1)
// ... 転送処理 ...
```

---

## 3. 最小コネクション数サーバーの選択

全アップストリームをスキャンして `connections` が最小かつ `alive == true` のサーバーを返します。

```
スキャン: 全アップストリームをループ
    → minConn = math.MaxInt64 で初期化
    → alive かつ connections < minConn なら candidate を更新
    → スキャン完了後の candidate を返す
```

複数台が同じ最小コネクション数の場合は、最初に見つかったサーバーを選びます（tie-breaking はラウンドロビンで行う実装もありますが、Step 5 では最初のサーバーで十分です）。

---

## 4. `int64` と `atomic`

コネクション数はインクリメントもデクリメントもするため、`uint64` ではなく `int64` を使います。

```go
connections int64  // 符号付き整数

atomic.AddInt64(&u.connections, 1)   // インクリメント
atomic.AddInt64(&u.connections, -1)  // デクリメント（負数を加算）
atomic.LoadInt64(&u.connections)     // 読み取り
```

---

## 📌 まとめ：Step 5 のフロー

1. `Accept` でクライアントの接続を受け取る
2. `selectLeastConnections(upstreams)` で alive かつ最小コネクション数のサーバーを選ぶ
3. 選んだサーバーの `connections` をアトミックにインクリメントする
4. `defer atomic.AddInt64(&upstream.connections, -1)` でデクリメントを予約する
5. Step 2 と同じ手順でヘッダーを書き換えて転送する
