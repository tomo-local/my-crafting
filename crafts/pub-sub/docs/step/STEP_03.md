# Step 3：Subscribe/Unsubscribeのライフサイクル（前提知識）

Step 2 まではサブスクライバーの追加だけを扱っていました。Step 3 では `UNSUB` コマンドと接続切断時のクリーンアップを実装し、「不要になったサブスクライバーをリストから正しく削除する」ライフサイクル管理を完成させます。

---

## 1. サブスクライバーの削除が必要なケース

| ケース | トリガー |
|---|---|
| クライアントが `UNSUB <topic>` コマンドを送る | `readLoop` でコマンドを検出 |
| クライアントのTCP接続が切れる | `scanner.Scan()` が `false` を返す |
| サーバー側でタイムアウト等が発生する | 今回は省略 |

どちらのケースも「サブスクライバーリストからの除去」と「チャネルのクローズ」が必要です。

---

## 2. スライスからの要素削除

Goのスライスに「削除」メソッドはありません。フィルタリングで新しいスライスを作ります。

```go
func removeSubscriber(subs []*Subscriber, target *Subscriber) []*Subscriber {
    result := subs[:0]  // 元のスライスのメモリを再利用
    for _, s := range subs {
        if s != target {
            result = append(result, s)
        }
    }
    return result
}
```

これを `Lock` の中で呼びます。

---

## 3. `close(sub.ch)` のタイミング

`writeLoop` は `for msg := range sub.ch` で動いています。`close(sub.ch)` を呼ぶと `range` が終了し、goroutine が自然に終了します。

```
切断処理の順序:
1. b.Unsubscribe(topic, sub)   ← リストから削除
2. close(sub.ch)               ← writeLoop goroutine を終了させる
3. （conn.Close() は defer で自動）
```

`close` の前にリストから削除することで、`Publish` が削除済みのサブスクライバーに書き込もうとする競合を防ぎます。

---

## 4. 1クライアントが複数トピックを購読する場合

Step 1〜2 では `Subscriber` が1トピックしか持っていませんでした。1クライアントが複数の `SUB` コマンドを送れるよう、`Subscriber` が購読トピックのセットを管理する構造に変えます。

```
Client（1つの接続）
  └── subscriptions: map[string]*Subscription
        ├── "news" → Subscription{ch: chan string}
        └── "events" → Subscription{ch: chan string}
```

各トピックへの購読が独立した `ch` を持つことで、`UNSUB news` しても `events` への配信は続きます。

---

## 📌 まとめ：Step 3 のフロー

**`UNSUB <topic>` を受け取ったとき:**
1. `b.Unsubscribe(topic, sub)` でリストから削除する
2. `sub` に紐づいたそのトピックの `ch` を `close` する
3. `+OK` を返す

**TCP接続が切れたとき（`scanner.Scan()` が false）:**
1. そのクライアントが購読している全トピックに対して `Unsubscribe` を呼ぶ
2. 全チャネルを `close` する
3. `conn.Close()` は `defer` に任せる
