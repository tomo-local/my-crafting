# Step 4：クエリビルダー（WHERE）（前提知識）

全件取得だけでは実用になりません。`Where()` を連鎖させて条件を絞り込む、メソッドチェーン型のクエリビルダーを実装します。

---

## 1. Builder パターン

クエリビルダーは「条件を積み上げる状態を持つ構造体」です。`Where()` を呼ぶたびに条件を追加し、`Find()` で実行します。

```
Query 構造体:
  db         *DB
  tableName  string
  conditions []Condition

Condition 構造体:
  Column string   // "age"
  Op     string   // ">"
  Value  interface{} // 20
```

`Where()` はレシーバのコピーを返す（`*Query` を返す）ことで、元の `Query` を変更せずにチェーンできます。

---

## 2. 条件の評価

インメモリストアを全件走査し、全ての条件が true のレコードだけを返します。AND 結合です。

```
for _, record := range store["users"] {
    if matchAll(record, conditions) {
        // 結果に追加
    }
}
```

比較演算子ごとの評価ロジック:

| Op | 評価 |
|---|---|
| `"="` | `record[col] == value` |
| `"!="` | `record[col] != value` |
| `">"` | `toFloat(record[col]) > toFloat(value)` |
| `"<"` | `toFloat(record[col]) < toFloat(value)` |

数値比較は `int` と `float64` が混在するため、共通の `float64` に変換してから比較すると統一的に扱えます。

---

## 3. `db.Where()` の入口

`DB.Where()` は新しい `Query` を生成して返します。テーブル名はまだわからないので、後で `Find()` が型情報から解決します。

```
db.Where("age", ">", 20) → &Query{db: db, conditions: [{age, >, 20}]}
  .Where("name", "=", "Alice") → &Query{..., conditions: [{age, >, 20}, {name, =, Alice}]}
  .Find(&users) → 絞り込んでスライスに詰める
```

---

## 📌 まとめ：Step 4 のフロー

1. `DB.Where(col, op, val)` で `Query` を生成して返す
2. `Query.Where(col, op, val)` で条件を追加した新しい `Query` を返す（レシーバコピー）
3. `Query.Find(out)` を呼ぶ:
   1. `out` の型からスキーマを取得してテーブル名を確定する
   2. 全レコードをループして `matchAll(record, conditions)` で絞り込む
   3. 条件に合致したレコードだけを struct に変換してスライスに追加する
