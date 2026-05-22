# Step 5 実装ガイド：ワイルドカードサブスクリプション

## ゴール

`news.*` で `news.sports` と `news.tech` 両方が届き、`news.sports.football` は届かないこと。

```bash
# ターミナル B: SUB news.*
nc localhost 4222
SUB news.*

# ターミナル C: SUB news.>
nc localhost 4222
SUB news.>

# ターミナル D: PUB news.sports Goal!
nc localhost 4222
PUB news.sports Goal!
# → B と C 両方に MSG news.sports Goal! が届くこと

# ターミナル D: PUB news.sports.football Touchdown!
PUB news.sports.football Touchdown!
# → C のみに届くこと（B には届かない）

# ターミナル D: PUB news Hello
PUB news Hello
# → B にも C にも届かないこと（* は1セグメント、news は0セグメント追加なのでマッチしない）
```

---

## 変更するファイル

```
go/
└── broker/
    ├── broker.go    ← Publish の全走査に変更
    └── match.go     ← 新規作成（matchTopic 関数）
```

---

## 1. `broker/match.go`

**`matchTopic(pattern, topic string) bool`:**

1. `patSegs := strings.Split(pattern, ".")` でパターンを分割する
2. `topSegs := strings.Split(topic, ".")` でトピックを分割する
3. `for i, pat := range patSegs` でループする:
   - `pat == ">"` → `return true`
   - `i >= len(topSegs)` → `return false`
   - `pat != "*" && pat != topSegs[i]` → `return false`
4. `return len(patSegs) == len(topSegs)`

```go
func matchTopic(pattern, topic string) bool {
    patSegs := strings.Split(pattern, ".")
    topSegs := strings.Split(topic, ".")

    for i, pat := range patSegs {
        if pat == ">" {
            return true
        }
        if i >= len(topSegs) {
            return false
        }
        if pat != "*" && pat != topSegs[i] {
            return false
        }
    }
    return len(patSegs) == len(topSegs)
}
```

---

## 2. `broker/broker.go` の修正

### `Publish` の全走査への変更

> **Step 4 との差分**
> `b.subscribers[topic]` の直接アクセスから、全パターンに対する `matchTopic` 走査に変える。

```go
func (b *Broker) Publish(topic, message string) {
    b.mu.RLock()
    // 全パターンに対してマッチングし、マッチしたサブスクライバーを収集
    var matched []*Subscriber
    seen := make(map[*Subscriber]bool)
    for pattern, subs := range b.subscribers {
        if matchTopic(pattern, topic) {
            for _, sub := range subs {
                if !seen[sub] {
                    seen[sub] = true
                    matched = append(matched, sub)
                }
            }
        }
    }
    b.mu.RUnlock()

    for _, sub := range matched {
        // Step 4 と同じノンブロッキング送信
        sub.mu.Lock()
        // ... 各 subscription に送信 ...
        sub.mu.Unlock()
    }
}
```

> **重複排除の理由**: 1つのサブスクライバーが `news.*` と `news.>` の両方を購読していると、同じサブスクライバーが2回 `matched` に入ります。`seen` マップで重複を排除します。

---

## 実装の確認手順

### ステップ 1: ビルドが通ること

```bash
go build ./...
```

### ステップ 2: `matchTopic` のユニットテスト

```go
// broker/match_test.go
func TestMatchTopic(t *testing.T) {
    cases := []struct{ pattern, topic string; want bool }{
        {"news.sports", "news.sports", true},
        {"news.*", "news.sports", true},
        {"news.*", "news.sports.football", false},
        {"news.>", "news.sports", true},
        {"news.>", "news.sports.football", true},
        {"news.>", "news", false},
        {"*", "news", true},
        {"*", "news.sports", false},
    }
    for _, c := range cases {
        got := matchTopic(c.pattern, c.topic)
        if got != c.want {
            t.Errorf("matchTopic(%q, %q) = %v, want %v", c.pattern, c.topic, got, c.want)
        }
    }
}
```

```bash
go test ./broker/...
```

### ステップ 3: 手動動作確認

上記ゴールのコマンドを実行して、各パターンのマッチ結果を確認する。

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `news.*` に `news.sports.football` が届く | `>` と `*` の処理が逆になっている | `*` は1セグメントのみ、`>` は複数セグメント。アルゴリズムを再確認 |
| 同じメッセージが2回届く | 同じサブスクライバーが複数パターンにマッチして重複収集されている | `seen` マップで重複排除する |
| 完全一致が動かなくなる | `Publish` で `matchTopic` を使うようにしたが完全一致のケースが落ちている | `matchTopic("news", "news")` → `true` であることを確認（`len(patSegs) == len(topSegs)` の条件が機能しているか） |
| パフォーマンスが落ちる | 全パターンを毎 Publish で走査している | Step 5 の台数規模（数百パターン以下）では問題なし。大規模が必要ならトライ木で最適化 |
