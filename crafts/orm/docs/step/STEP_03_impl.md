# Step 3 実装ガイド：SELECT / Find

## ゴール

```go
var users []*User
db.FindAll(&users)
for _, u := range users {
    fmt.Println(u.ID, u.Name, u.Age)
}
// 1 Alice 30
// 2 Bob 25
```

## 変更するファイル

```
go/
└── internal/
    └── orm/
        ├── orm.go    ← FindAll() 追加・scanRecord() ヘルパー追加
        └── schema.go ← colIndex() ヘルパー追加（任意）
```

## 実装手順

### 1. `FindAll(out interface{}) error` を実装する

内部でやること（この順番で）:

1. `reflect.TypeOf(out)` で型を取得し、`*[]*Struct` 形式であることを確認する
2. 要素の struct 型（`User`）を取り出す: `.Elem().Elem().Elem()`
3. `db.Schema()` でスキーマを取得してテーブル名を確定する
4. `db.store[tableName]` が存在しなければ空のまま return する
5. カラム名 → FieldIndex の逆引きマップを構築する（後で `scanRecord` に渡す）
6. 全レコードをループして `reflect.New(elemType)` で struct を生成 → `scanRecord` で値をセット → `reflect.Append` でスライスに追加する
7. 最後に `reflect.ValueOf(out).Elem().Set(sliceVal)` で結果を書き戻す

### 2. `scanRecord(rv reflect.Value, record Record, colMap map[string]int)` ヘルパーを実装する

内部でやること（この順番で）:

1. `record` のキーをループする
2. `colMap[columnName]` でフィールドインデックスを取得する（存在しなければ skip）
3. `rv.Field(idx)` でフィールドの Value を取得する
4. `record[col]` の値を型に応じてセットする:
   - `switch rv.Field(idx).Kind()` で分岐
   - `reflect.Int`: `SetInt(int64(v.(int)))`
   - `reflect.String`: `SetString(v.(string))`

> **型アサーション**: ストアに保存したのと同じ型でアサーションする（`Insert` で `Interface()` して保存したので元の型のまま）。

### 3. `main.go` で動作確認する

```go
db.Insert(&User{Name: "Alice", Age: 30})
db.Insert(&User{Name: "Bob", Age: 25})

var users []*User
if err := db.FindAll(&users); err != nil {
    log.Fatal(err)
}
for _, u := range users {
    fmt.Printf("%d %s %d\n", u.ID, u.Name, u.Age)
}
```

## 実装の確認手順

```bash
go build ./go/...
go run ./go/main.go
# 1 Alice 30
# 2 Bob 25
```

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `reflect.Value.Elem of non-pointer` でパニック | `FindAll(&users)` でなく `FindAll(users)` を渡している | 引数はスライスのポインタ `&users` で渡す |
| フィールドが空のまま（ゼロ値） | `colMap` の逆引きが失敗している（カラム名の大文字/小文字不一致） | Insert 時と FindAll 時で使うカラム名が一致しているか確認する |
| `interface conversion: interface {} is int, not int64` | `SetInt` に渡す前の型アサーションが `int64` になっている | Insert 時に `Interface()` で格納した型（`int`）と一致させる |
| スライスの長さが 0 のまま | `reflect.Append` の結果を `Set` し忘れている | `sliceVal.Set(reflect.Append(...))` は Set を呼ばないと反映されない |
