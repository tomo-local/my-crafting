# Step 2 実装ガイド：INSERT

## ゴール

```go
u1 := &User{Name: "Alice", Age: 30}
db.Insert(u1)
fmt.Println(u1.ID) // → 1

u2 := &User{Name: "Bob", Age: 25}
db.Insert(u2)
fmt.Println(u2.ID) // → 2
```

## 変更するファイル

```
go/
└── internal/
    └── orm/
        ├── orm.go      ← store・counters フィールド追加・Insert() 実装
        └── schema.go   ← 変更なし
```

## 実装手順

### 1. `DB` 構造体にストアを追加する

`orm.go` の `DB` に以下のフィールドを追加する:

- `store map[string][]Record`（`Record` は `map[string]interface{}` の type alias）
- `counters map[string]int`（テーブルごとの自動採番カウンター）

`New()` でゼロ値ではなく `make()` で初期化することを忘れずに。

### 2. `Insert(v interface{}) error` を実装する

内部でやること（この順番で）:

1. `reflect.ValueOf(v).Elem()` で struct の Value を取得する（ポインタ前提）
2. スキーマを取得する（`db.Schema(v)` を流用）
3. PK フィールドの値を確認する。ゼロ値なら:
   - `db.counters[tableName]++` でカウンターを増やす
   - `rv.Field(pkIndex).SetInt(int64(newID))` で struct に書き戻す
4. 全カラムをループして `record[col.ColumnName] = rv.Field(col.FieldIndex).Interface()` で Record を構築する
5. `db.store[tableName] = append(db.store[tableName], record)` で保存する

> **スキーマとの連携**: `db.Schema(v)` は `interface{}` を受け取れるよう、内部でポインタ→Elem の変換を行っておく。

### 3. `main.go` で動作確認する

```go
db := orm.New()
u1 := &User{Name: "Alice", Age: 30}
if err := db.Insert(u1); err != nil {
    log.Fatal(err)
}
fmt.Println(u1.ID) // 1

u2 := &User{Name: "Bob", Age: 25}
db.Insert(u2)
fmt.Println(u2.ID) // 2
```

## 実装の確認手順

```bash
# ビルド確認
go build ./go/...

# 実行
go run ./go/main.go
# 1
# 2
```

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `reflect: call of reflect.Value.SetInt on zero Value` | `reflect.ValueOf(v)` に nil を渡している | Insert に非ポインタを渡していないか確認する |
| ID が毎回 0 のまま | ポインタでなく値を渡している（書き戻しが反映されない） | `db.Insert(&user)` とポインタで渡す |
| `store` がゼロ値 map でパニック | `DB.store` を `make()` していない | `New()` 内で `store: make(map[string][]Record)` する |
| 同じ struct で 2 回 Insert すると ID が重複する | PK がゼロ値かどうかのチェックが抜けている | `rv.Field(pkIndex).IsZero()` で判定する |
