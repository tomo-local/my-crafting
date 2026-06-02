# Step 5 実装ガイド：DocumentChange イベントと TargetChange(CURRENT)

## ゴール

```bash
# alice が存在する状態で購読開始すると、まず ADDED が届き、続いて CURRENT が届くこと
curl -N -X POST http://localhost:8080/v1/listen \
  -H "Content-Type: application/json" \
  -d '{"type":"AddTarget","path":"users/alice","targetId":1}'

# 期待する出力:
# data: {"type":"DocumentChange","changeType":"ADDED","path":"users/alice",...}
# data: {"type":"TargetChange","changeType":"CURRENT","targetId":1,"resumeToken":"MQ=="}
# （以降は書き込みがあったときだけ届く）
```

---

## 変更するファイル

```
go/
└── internal/
    ├── store/
    │   └── store.go       （CurrentVersion を追加）
    ├── watcher/
    │   └── watcher.go     （Event に type フィールドを追加）
    └── server/
        └── server.go      （handleListen に初期スナップショット送信を追加）
```

---

## `store.go` への追加

```go
func (s *Store) CurrentVersion() SnapshotVersion {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.version
}
```

---

## `watcher.go` への変更

`Event` に `Type` フィールドを追加して DocumentChange と TargetChange を区別する:

```go
type EventType string

const (
    TypeDocumentChange EventType = "DocumentChange"
    TypeTargetChange   EventType = "TargetChange"
)

type Event struct {
    Type        EventType
    ChangeType  ChangeType
    Path        string
    Document    *store.Document
    Version     store.SnapshotVersion
    TargetID    int
    ResumeToken string // TargetChange(CURRENT) のときのみ
}
```

> **Step 4 との差分**: `Type` と `TargetID`、`ResumeToken` フィールドを追加する。  
> 既存の `Publish` 呼び出しも `Type: TypeDocumentChange` をセットするよう更新する。

---

## `handleListen` への変更

`Subscribe` を呼んだ直後、イベントループの前に以下を挿入する:

### 初期スナップショット送信フェーズ

内部でやること（順番どおり）:

1. `doc, exists := s.store.Get(req.Path)` で存在チェック
2. exists なら `ADDED` イベントを SSE で送り `Flush()`
3. `v := s.store.CurrentVersion()`
4. `resumeToken := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", v)))`
5. TargetChange イベントを JSON にして SSE で送り `Flush()`

```go
currentEvent := watcher.Event{
    Type:        watcher.TypeTargetChange,
    ChangeType:  "CURRENT",
    TargetID:    req.TargetID,
    ResumeToken: resumeToken,
}
```

---

## 実装の確認手順

```bash
go build ./go/...

go run ./go/main.go &

# alice を事前に作成
curl -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -H "Content-Type: application/json" -d '{"name":"Alice"}'

# 購読開始 → ADDED + CURRENT が届くこと
curl -N -X POST http://localhost:8080/v1/listen \
  -H "Content-Type: application/json" \
  -d '{"type":"AddTarget","path":"users/alice","targetId":1}'
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| CURRENT が届かない | 初期スナップショット後の Flush を忘れた | CURRENT イベント送信後に `flusher.Flush()` |
| ADDED が届かない | `store.Get` と `Subscribe` の順序が逆 | `Subscribe` → `Get` の順にする（逆だと書き込みを見逃す） |
| resumeToken が空 | `base64` パッケージを import していない | `encoding/base64` を追加 |
