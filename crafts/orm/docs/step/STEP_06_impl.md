# Step 6 実装ガイド：Associations & Eager Loading

## ゴール

```go
// N+1 あり（クエリカウンター: 1 + N 回）
var posts []*Post
db.FindAll(&posts)
for _, p := range posts {
    var users []*User
    db.Where("id", "=", p.UserID).Find(&users)
    p.User = users[0]
}
fmt.Println("queries:", db.QueryCount) // → 4（1件 + 3件分）

// Eager Loading（クエリカウンター: 2 回）
db.QueryCount = 0
db.Preload("User").FindAll(&posts)
fmt.Println("queries:", db.QueryCount) // → 2
for _, p := range posts {
    fmt.Println(p.Title, p.User.Name)
}
```

## 変更するファイル

```
go/
└── internal/
    └── orm/
        ├── orm.go    ← QueryCount フィールド追加・Preload() 追加・"in" 演算子対応
        └── query.go  ← Preload フィールド追加・resolvePreloads() 追加
```

## 実装手順

### 1. クエリカウンターを追加する

`DB` 構造体に `QueryCount int` を追加し、`store[tableName]` にアクセスするたびにインクリメントする。

### 2. `"in"` 演算子を `matchAll` に追加する

`query.go` の `matchAll` の `switch cond.Op` に `"in"` ケースを追加する:

1. `reflect.ValueOf(cond.Value)` の Kind が `Slice` であることを確認する
2. スライスをループして `record[col] == elem` なら true を返す

### 3. `Query` に `preloads []string` フィールドを追加する

`query.go` の `Query` 構造体に `preloads []string` を追加する。

`Preload(field string) *Query` メソッドを追加:

```go
func (q *Query) Preload(field string) *Query {
    q2 := *q
    q2.preloads = append(append([]string{}, q.preloads...), field)
    return &q2
}
```

### 4. `orm.go` に `Preload()` を追加する

```go
func (db *DB) Preload(field string) *Query {
    return &Query{db: db, preloads: []string{field}}
}
```

### 5. `resolvePreloads(records []interface{}, preloads []string)` を実装する

内部でやること（この順番で）:

1. 各 preload フィールド名について:
   1. `records[0]` の struct 型からそのフィールドを `reflect.Type.FieldByName(fieldName)` で探す
   2. タグ（`belongs_to` / `has_many`）から外部キー名と関連テーブルの型を取得する
   3. 全レコードから外部キーの値を収集してユニーク化する（`map[interface{}]bool`）
   4. `db.Where("id", "in", idSlice).Find(&relatedRecords)` で一括取得する
   5. `map[pkValue]*RelatedStruct` を構築する
   6. 各レコードの関連フィールドに対応するポインタを `reflect.Value.Set` でセットする

> **Step 3 との差分**: `FindAll` の後に `resolvePreloads` を呼ぶだけでよい。`FindAll` 本体を変更する必要はない。

### 6. `FindAll` と `Find` で preloads を処理する

`Query.FindAll` と `Query.Find` の末尾で `resolvePreloads` を呼ぶ。

### 7. `main.go` で動作確認する

```go
type Post struct {
    ID     int    `orm:"pk"`
    Title  string `orm:"column:title"`
    UserID int    `orm:"column:user_id"`
    User   *User  `orm:"belongs_to:User,fk:user_id"`
}

db := orm.New()
alice := &User{Name: "Alice", Age: 30}
bob := &User{Name: "Bob", Age: 25}
db.Insert(alice)
db.Insert(bob)
db.Insert(&Post{Title: "Hello", UserID: alice.ID})
db.Insert(&Post{Title: "World", UserID: bob.ID})
db.Insert(&Post{Title: "Again", UserID: alice.ID})

// Eager Loading
var posts []*Post
db.Preload("User").FindAll(&posts)
fmt.Println("queries:", db.QueryCount) // 2
for _, p := range posts {
    fmt.Println(p.Title, "by", p.User.Name)
}
```

## 実装の確認手順

```bash
go build ./go/...
go run ./go/main.go
# queries: 2
# Hello by Alice
# World by Bob
# Again by Alice
```

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `reflect: call of reflect.Value.Set using zero Value` | `FieldByName` が存在しないフィールド名を探している | タグのフィールド名とコード上のフィールド名が一致しているか確認する |
| preload 後も `post.User` が nil | `resolvePreloads` が呼ばれていない、または `Set` の引数の型が合っていない | `reflect.TypeOf(ptr)` と `field.Type()` をログに出して比較する |
| `"in"` 演算子が動かない | `cond.Value` が `[]int` でなく `[]interface{}` になっている | idSlice を構築する際に `[]interface{}` に統一するか、reflect でスライスを動的に生成する |
| QueryCount が期待値より多い | `resolvePreloads` 内で ID ごとに個別クエリを発行している | `"in"` を使った一括取得に切り替えているか確認する |
