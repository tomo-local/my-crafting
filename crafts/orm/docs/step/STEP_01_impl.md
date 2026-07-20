# Step 1 実装ガイド：型マッピングとスキーマ定義

## ゴール

```go
type User struct {
    ID   int    `orm:"pk"`
    Name string `orm:"column:name"`
    Age  int    `orm:"column:age"`
}

db := orm.New()
schema := db.Schema(User{})
fmt.Printf("table: %s\n", schema.TableName)
for _, col := range schema.Columns {
    fmt.Printf("  column: %s, pk: %v\n", col.ColumnName, col.IsPK)
}
// table: users
//   column: id, pk: true
//   column: name, pk: false
//   column: age, pk: false
```

## 変更するファイル

```
go/
├── main.go
└── internal/
    └── orm/
        ├── orm.go      ← DB 構造体・New()・Schema()
        └── schema.go   ← Schema・ColumnDef・スキーマ生成ロジック
```

## 実装手順

### 1. `internal/orm/schema.go` を作る

内部でやること（この順番で）:

- `ColumnDef` 構造体を定義する（`FieldIndex int`・`ColumnName string`・`IsPK bool`）
- `Schema` 構造体を定義する（`TableName string`・`Columns []ColumnDef`）
- `buildSchema(t reflect.Type) *Schema` 関数を実装する:
  1. `strings.ToLower(t.Name()) + "s"` でテーブル名を生成する
  2. `for i := 0; i < t.NumField(); i++` で全フィールドをループする
  3. `t.Field(i).Tag.Get("orm")` でタグ文字列を取得する
  4. タグを `,` で split して各部分を解析する（`pk` なら IsPK=true、`column:xxx` なら ColumnName="xxx"）
  5. ColumnName が未指定なら `strings.ToLower(フィールド名)` をデフォルトにする
  6. `ColumnDef` を構築して `Schema.Columns` に追加する

### 2. `internal/orm/orm.go` を作る

内部でやること:

- `DB` 構造体を定義する（`schemaCache map[reflect.Type]*Schema` を持つ）
- `New() *DB` で初期化して返す
- `Schema(v interface{}) *Schema` を実装する:
  1. `reflect.TypeOf(v)` で型を取得する（ポインタなら `.Elem()`）
  2. キャッシュに存在すれば返す
  3. `buildSchema(t)` で生成してキャッシュに保存してから返す

### 3. `main.go` で動作確認する

```go
db := orm.New()
schema := db.Schema(User{})
fmt.Printf("table: %s\n", schema.TableName)
for _, col := range schema.Columns {
    fmt.Printf("  column: %-10s pk: %v\n", col.ColumnName, col.IsPK)
}
```

## 実装の確認手順

```bash
# ビルド確認
go build ./go/...

# 実行
go run ./go/main.go
# table: users
#   column: id         pk: true
#   column: name       pk: false
#   column: age        pk: false
```

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `reflect.TypeOf(v).Name()` が空文字になる | ポインタを渡していて `Elem()` していない | `if t.Kind() == reflect.Ptr { t = t.Elem() }` を先頭に入れる |
| タグが空文字のフィールドがある | `orm:""` や タグなしのフィールド | タグが空なら `strings.ToLower(フィールド名)` をカラム名にする |
| `pk` と `column:id` を両方書きたいが片方しか解析されない | `,` で split した後のループが途中で break している | `continue` ではなく全部ループして両方のキーを処理する |
