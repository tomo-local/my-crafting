# Step 4 実装ガイド：クエリビルダー（WHERE）

## ゴール

```go
var users []*User
db.Where("age", ">", 20).Where("name", "=", "Alice").Find(&users)
fmt.Println(len(users))      // → 1
fmt.Println(users[0].Name)   // → Alice
```

## 変更するファイル

```
go/
└── internal/
    └── orm/
        ├── orm.go     ← Where() 追加
        ├── schema.go  ← 変更なし
        └── query.go   ← Query・Condition・Find()・matchAll() を新規作成
```

## 実装手順

### 1. `internal/orm/query.go` を新規作成する

内部でやること（この順番で）:

- `Condition` 構造体を定義する（Column・Op・Value）
- `Query` 構造体を定義する（db・conditions []Condition）
- `Where(col, op string, val interface{}) *Query` を実装する:
  - `conditions` に新しい `Condition` を追加してレシーバをコピーして返す
- `Find(out interface{}) error` を実装する:
  1. Step 3 の `FindAll` と同じ型解析で `elemType` とスキーマを取得する
  2. 全レコードをループして `matchAll(record, q.conditions)` が true なものだけ処理する
  3. struct 生成・scanRecord・Append は Step 3 と同じ

- `matchAll(record Record, conds []Condition) bool` ヘルパー:
  1. 全条件をループして1つでも false なら `return false`
  2. `switch cond.Op` で `=`, `!=`, `>`, `<`, `>=`, `<=` を実装する
  3. 数値比較は `toFloat64(v interface{}) float64` ヘルパーで `int`・`float64` を統一する

### 2. `orm.go` に `Where()` を追加する

```go
func (db *DB) Where(col, op string, val interface{}) *Query {
    return &Query{db: db, conditions: []Condition{{col, op, val}}}
}
```

### 3. `main.go` で動作確認する

```go
db.Insert(&User{Name: "Alice", Age: 30})
db.Insert(&User{Name: "Bob", Age: 25})
db.Insert(&User{Name: "Carol", Age: 17})

var adults []*User
db.Where("age", ">", 20).Find(&adults)
fmt.Println(len(adults)) // 2

var alice []*User
db.Where("age", ">", 20).Where("name", "=", "Alice").Find(&alice)
fmt.Println(len(alice))      // 1
fmt.Println(alice[0].Name)   // Alice
```

## 実装の確認手順

```bash
go build ./go/...
go run ./go/main.go
# 2
# 1
# Alice
```

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `">"` 比較が常に false | `int` と `int` を直接比較していて型アサーションが失敗している | `toFloat64()` で両辺を変換してから比較する |
| `Where().Where()` で最初の条件が消える | `conditions` をコピーせずスライスの参照を共有している | `append` は新しいスライスを返すので問題ない（ただし元のスライスの容量に注意） |
| `Find()` と `FindAll()` でコードが重複する | 条件フィルタリングと struct 変換が混在している | フィルタリング後のレコードスライスを `scanRecords(out, records)` に渡す形に切り出す |
