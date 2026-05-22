# Step 2：ファンアウト（前提知識）

Step 1 では1サブスクライバーへの配信でした。Step 2 では同一トピックに複数のサブスクライバーがいるとき、全員に同時配信するファンアウトを完成させます。

---

## 1. ファンアウトとは

1つの入力（メッセージ）を複数の出力（サブスクライバー）に同時に届けるパターンです。

```
PUB news "Breaking news!"
            ↓
     Broker.Publish("news", ...)
            ↓
    ┌───────┼───────┐
    ↓       ↓       ↓
  subA.ch subB.ch subC.ch
    ↓       ↓       ↓
  MSG を   MSG を   MSG を
  返す     返す     返す
```

---

## 2. なぜ goroutine per subscriber が重要か

`Publish` の中でサブスクライバーのチャネルに書き込んだとき、チャネルが満杯ならブロックします。サブスクライバーAのチャネルでブロックしている間、サブスクライバーBへの配信も止まります。

```
Publish ループ:
    subA.ch <- msg  ← subA が遅い → ここでブロック
    subB.ch <- msg  ← subB は速いのに subA のせいで届かない
    subC.ch <- msg  ← 同様
```

各サブスクライバーが独立した goroutine でチャネルを読み出しているため、**チャネルへの書き込みさえ成功すれば**次のサブスクライバーに進めます（Step 4 でチャネルへの書き込み自体も non-blocking にする）。

---

## 3. サブスクライバーリストのコピー

`Publish` が `RLock` でリストを読んでいる間、別goroutineが `Lock` でリストを変更しようとするとブロックします。ロック保持時間を最小化するため、**ロック内でリストをコピーしてからロックを解放**してファンアウトする設計も有効です。

```go
b.mu.RLock()
subs := make([]*Subscriber, len(b.subscribers[topic]))
copy(subs, b.subscribers[topic])
b.mu.RUnlock()

for _, sub := range subs {  // ← ロック外でファンアウト
    sub.ch <- message
}
```

トレードオフ: ロック保持時間は短くなるが、メモリコピーが発生する。サブスクライバー数が少ない場合はロック内でファンアウトしても問題ない。

---

## 📌 まとめ：Step 2 のフロー

Step 1 からの変更はほぼありません。`Publish` のループが複数サブスクライバーを正しく処理できているかを確認することが主な作業です。

1. 複数クライアントが同じトピックに `SUB` する
2. `PUB` したとき、`b.subscribers[topic]` に複数の `*Subscriber` が入っている
3. `for _, sub := range subs` で全員のチャネルに書き込む
4. 各サブスクライバーの `writeLoop` goroutine がそれぞれ `MSG` を送る
