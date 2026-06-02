# Step 1 実装ガイド：In-memory Document Store

## ゴール

```bash
go test ./go/internal/store/...
# ok  github.com/tomo-local/firestore/internal/store
```

---

## 変更するファイル

```
go/
├── go.mod
└── internal/
    └── store/
        ├── store.go
        └── store_test.go
```

---

## `store.go` の実装手順

### 1. 型定義

```go
type SnapshotVersion uint64

type Document struct {
    Fields  map[string]interface{}
    Version SnapshotVersion
}
```

### 2. Store 構造体

```go
type Store struct {
    mu      sync.RWMutex
    data    map[string]*Document
    version SnapshotVersion
}

func New() *Store {
    return &Store{data: make(map[string]*Document)}
}
```

### 3. Put

内部でやること（順番どおり）:
1. `s.mu.Lock()` / `defer s.mu.Unlock()`
2. `s.version++`
3. `s.data[path] = &Document{Fields: fields, Version: s.version}`
4. `return s.version`

### 4. Get

内部でやること:
1. `s.mu.RLock()` / `defer s.mu.RUnlock()`
2. `doc, ok := s.data[path]`
3. `return doc, ok`

### 5. Delete

内部でやること:
1. `s.mu.Lock()` / `defer s.mu.Unlock()`
2. 存在チェック（なければ `return 0, false`）
3. `s.version++`
4. `delete(s.data, path)`
5. `return s.version, true`

### 6. List

内部でやること:
1. `s.mu.RLock()` / `defer s.mu.RUnlock()`
2. `prefix := collection + "/"`
3. path が prefix で始まるものをスライスに集めて返す

---

## `store_test.go` の実装手順

以下の4ケースをテストする:

1. `Put` → `Get` で同じフィールドが返ること
2. `Put` を2回呼んでバージョンが単調増加すること
3. `Delete` 後に `Get` が `false` を返すこと
4. `List` でコレクション内のドキュメントだけ返ること

---

## 実装の確認手順

```bash
cd /path/to/crafts/firestore
go build ./go/...
go test ./go/internal/store/...
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `Put` の返り値がゼロ | インクリメント前に version を返している | `s.version++` の後に `return s.version` |
| `List` で他コレクションのドキュメントも返る | prefix チェックがない | `strings.HasPrefix(path, collection+"/")` |
| データ競合で `-race` が落ちる | RLock と Lock の使い分けが逆 | Read 系は `RLock`、Write 系は `Lock` |
