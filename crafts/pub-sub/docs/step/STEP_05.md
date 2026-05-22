# Step 5：ワイルドカードサブスクリプション（前提知識）

Step 4 まではトピック名が完全一致するサブスクライバーにのみ配信していました。Step 5 では `news.*` や `events.>` のようなパターンサブスクリプションを実装します。

---

## 1. トピック名の階層構造

トピック名は `.`（ドット）で区切られた階層として扱います。

```
"news.sports.football"
 セグメント: ["news", "sports", "football"]

"events.jp.tokyo"
 セグメント: ["events", "jp", "tokyo"]
```

---

## 2. ワイルドカードの仕様

| 記号 | 意味 | 例 |
|---|---|---|
| `*` | 任意の1セグメント | `news.*` → `news.sports` ○、`news.sports.football` ✗ |
| `>` | 残り全セグメント（1つ以上） | `news.>` → `news.sports` ○、`news.sports.football` ○ |

`>` は最後のセグメントとしてのみ有効です（`news.>.uk` のような途中の使用は無効）。

---

## 3. パターンマッチングアルゴリズム

トピック名とパターンをどちらも `strings.Split(s, ".")` でセグメントに分割し、先頭から比較します。

```
matchTopic(pattern, topic string) bool

手順:
  patSegs := strings.Split(pattern, ".")
  topSegs := strings.Split(topic, ".")

  for i, pat := range patSegs:
    if pat == ">" → return true（残り全てマッチ）
    if i >= len(topSegs) → return false（トピックが短すぎる）
    if pat != "*" && pat != topSegs[i] → return false（不一致）

  return len(patSegs) == len(topSegs)（長さも一致している必要がある）
```

例:

```
pattern: "news.*"  → ["news", "*"]
topic:   "news.sports" → ["news", "sports"]

i=0: "news" == "news" → OK
i=1: "*" → スキップ
終了: len(2) == len(2) → true ✓

pattern: "news.*"  → ["news", "*"]
topic:   "news.sports.football" → ["news", "sports", "football"]

i=0: "news" == "news" → OK
i=1: "*" → スキップ
終了: len(2) != len(3) → false ✗
```

---

## 4. Publish でのパターンマッチング

`Publish("news.sports", msg)` を呼んだとき、ブローカーはすべての購読パターンに対して `matchTopic` を実行します。

```
購読中のパターン:
  "news.sports"   → マッチ ✓
  "news.*"        → マッチ ✓
  "news.>"        → マッチ ✓
  "events.*"      → マッチ ✗
```

完全一致はパターンマッチの特殊ケースとして同じ関数で処理できます。

---

## 5. ブローカーのデータ構造の変更

Step 4 まではキーがトピック名の完全一致（`map[string][]*Subscriber`）でした。ワイルドカードを使う場合、キーが**パターン**になります。

```go
// Step 4 まで: 完全一致
subscribers map[string][]*Subscriber

// Step 5 以降: パターン（ワイルドカードを含む可能性がある）
subscribers map[string][]*Subscriber  // キーはパターン文字列

// Publish 時に全パターンに対してマッチング
for pattern, subs := range b.subscribers {
    if matchTopic(pattern, topic) {
        // ファンアウト
    }
}
```

---

## 📌 まとめ：Step 5 の変更点

1. `matchTopic(pattern, topic string) bool` 関数を追加する
2. `Publish` の中で `b.subscribers` を全走査し、パターンにマッチしたサブスクライバーにファンアウトする
3. それ以外のコードは Step 4 と変わらない
