# Step 5 実装ガイド：UPDATE / DELETE

## ゴール

```go
// UPDATE
u.Name = "Alicia"
db.Update(u)

var found []*User
db.Where("id", "=", 1).Find(&found)
fmt.Println(found[0].Name) // → Alicia

// DELETE
db.Delete(u)
var all []*User
db.FindAll(&all)
fmt.Println(len(all)) // → 1（Bob のみ）
```

## 変更するファイル

```
go/
└── internal/
    └── orm/
        └── orm.go  ← Update()・Delete()・pkValue() ヘルパー追加
```

## 実装手順

### 1. `pkValue(v interface{}) (interface{}, error)` ヘルパーを追加する

内部でやること:

1. スキーマを取得して PK カラムを探す
2. `reflect.ValueOf(v).Elem().Field(pkIndex).Interface()` で PK 値を取得する
3. `int` 型のゼロ値（`== 0`）チェックを行い、ゼロならエラーを返す

> **Step 2 との差分**: Insert は ゼロ値 PK を「自動採番する」が、Update/Delete はゼロ値 PK を「エラー」とする。

### 2. `Update(v interface{}) error` を実装する

内部でやること（この順番で）:

1. `pkValue(v)` で PK を取得する（エラーなら return）
2. スキーマでテーブル名を確定する
3. `store[tableName]` をループして `record["id"] == pk` なインデックスを探す
4. 見つかったら全カラムを `reflect.ValueOf(v).Elem().Field(i).Interface()` で上書きする
5. 見つからなければ `errors.New("record not found")` を返す

### 3. `Delete(v interface{}) error` を実装する

内部でやること（この順番で）:

1. `pkValue(v)` で PK を取得する（エラーなら return）
2. スキーマでテーブル名を確定する
3. フィルタリング: PK が一致しないレコードだけで新しいスライスを構築する
4. 元のスライス長と比較して変わらなければ `errors.New("record not found")` を返す
5. `store[tableName]` を新しいスライスで置き換える

### 4. `main.go` で動作確認する

```go
u := &User{Name: "Alice", Age: 30}
db.Insert(u)
db.Insert(&User{Name: "Bob", Age: 25})

// UPDATE
u.Name = "Alicia"
db.Update(u)

var found []*User
db.Where("id", "=", u.ID).Find(&found)
fmt.Println(found[0].Name) // Alicia

// DELETE
db.Delete(u)
var all []*User
db.FindAll(&all)
fmt.Println(len(all)) // 1
```

## 実装の確認手順

```bash
go build ./go/...
go run ./go/main.go
# Alicia
# 1
```

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| Update 後も古い値が残る | `store[i] = newRecord` でなく `store[i][col]` を書き換えていない | `Record` は `map[string]interface{}` なのでキーを直接更新する |
| Delete 後も件数が変わらない | フィルタ後のスライスを `store[tableName]` に代入していない | `db.store[tableName] = filtered` の代入を忘れずに |
| PK ゼロ値チェックが int 以外で動かない | `switch v.(type)` で `int` のみ対応している | 将来的に `int64`・`string` PK も考慮するなら `fmt.Sprintf("%v", v) == "0"` は避け、reflect の Kind で判定する |
