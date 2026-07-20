# Step 3：SELECT / Find（前提知識）

Step 2 で保存したレコードを読み出し、struct のポインタスライスに詰めて返します。「ストアから struct へ」の逆方向のマッピングが核心です。

---

## 1. 出力型の受け取り方

`FindAll(&users)` の引数は `*[]*User` 型です。関数シグネチャは `interface{}` で受け取り、リフレクションで型を特定します。

```
reflect.TypeOf(&users)          → *[]*User  (Ptr)
reflect.TypeOf(&users).Elem()   → []*User   (Slice)
reflect.TypeOf(&users).Elem().Elem() → *User (Ptr)
reflect.TypeOf(&users).Elem().Elem().Elem() → User (Struct)
```

この `User` の型情報からスキーマを取得し、テーブル名を特定します。

---

## 2. レコードから struct への値のセット

ストアの `Record（map[string]interface{}）` から struct のフィールドに値をセットするには、`reflect.Value.Set` 系メソッドを使います。

```
record["name"] = "Alice"
→ rv.Field(nameFieldIndex).SetString("Alice")
```

フィールドの型（`Kind`）によってセットメソッドが変わります。

| Kind | セットメソッド |
|---|---|
| `reflect.Int` | `SetInt(int64)` |
| `reflect.String` | `SetString(string)` |
| `reflect.Bool` | `SetBool(bool)` |
| `reflect.Float64` | `SetFloat(float64)` |

`interface{}` に入った値を型アサーションして正しい型に変換してからセットします。

---

## 3. 新しい struct の動的生成

スライスに追加する struct は毎回新しいゼロ値から作ります。`reflect.New(elemType)` でポインタを生成できます。

```
elemType := reflect.TypeOf(User{})   // User の Type
newPtr := reflect.New(elemType)      // *User（ゼロ値）
newVal := newPtr.Elem()              // User（フィールドに書き込める）
```

---

## 4. スライスへの動的 append

結果スライスへの追加もリフレクション経由で行います。

```
sliceVal := reflect.ValueOf(out).Elem()   // []*User の Value
sliceVal.Set(reflect.Append(sliceVal, newPtr))
```

`out` は `*[]*User` なので、`.Elem()` で `[]*User` の Value を取得し、`reflect.Append` で追加した新しいスライスを `Set` でセットします。

---

## 📌 まとめ：Step 3 のフロー

1. `db.FindAll(out)` でスライスのポインタを受け取る
2. リフレクションで要素型（`User`）を特定してスキーマを取得する
3. `store[tableName]` の全レコードをループする
4. 各レコードに対して `reflect.New(elemType)` で新しい struct を生成する
5. カラム名 → FieldIndex のマップを使って `record[col]` の値を struct にセットする
6. `reflect.Append` で結果スライスに追加する
