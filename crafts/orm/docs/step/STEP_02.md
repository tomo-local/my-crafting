# Step 2：INSERT（前提知識）

Step 1 で作ったスキーマ定義を使い、struct の値を読み取ってインメモリストアに保存します。Step 2 で「ORM の書き込み」の基本が揃います。

---

## 1. インメモリストアの設計

データを保存する場所として、テーブル名をキーにしたマップを使います。

```
store: map[string][]Record
  "users" → [
    {"id": 1, "name": "Alice", "age": 30},
    {"id": 2, "name": "Bob",   "age": 25},
  ]
```

`Record` は `map[string]interface{}` で表現します。型を統一するためのコストは後で払い、まずはシンプルに動かします。

---

## 2. フィールド値の読み取り

`reflect.ValueOf(&user).Elem()` で struct の Value を取得し、`Field(i).Interface()` でフィールドの値を `interface{}` として取り出せます。

```
reflect.ValueOf(&user).Elem().Field(0).Interface() → 0 (int)
reflect.ValueOf(&user).Elem().Field(1).Interface() → "Alice" (string)
```

スキーマの `ColumnDef.FieldIndex` がフィールドのインデックスに対応しているので、これを使って「カラム名 → 値」のマップを構築します。

---

## 3. 自動採番（Auto Increment）

PK フィールドが `0`（ゼロ値）の場合、ORM 側でカウンターをインクリメントして ID を自動付与します。

```
tableCounters: map[string]int
  "users" → 2（これまでに2件 insert 済み）

INSERT 時:
  tableCounters["users"]++  → 3
  record["id"] = 3
```

付与した ID を元の struct のポインタに書き戻す必要があります（`db.Insert(&u)` の呼び出し後に `u.ID` が更新されること）。

---

## 4. ポインタへの書き戻し

自動採番した ID を struct の PK フィールドに書き戻すには、`reflect.Value.SetInt` を使います。

```
pkField := reflect.ValueOf(v).Elem().Field(pkIndex)
pkField.SetInt(int64(newID))
```

`SetInt` は `int64` を要求するので型変換が必要です。

---

## 📌 まとめ：Step 2 のフロー

1. `db.Insert(v)` でポインタを受け取る
2. スキーマを取得する（Step 1 のキャッシュを活用）
3. PK フィールドがゼロ値なら自動採番する
4. `reflect.Value.Field(i).Interface()` で全フィールドの値を読み取る
5. `Record（map[string]interface{}）` を構築する
6. `store[tableName]` にレコードを追加する
7. 自動採番した ID を元の struct に書き戻す
