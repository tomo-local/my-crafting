# Step 1：型マッピングとスキーマ定義（前提知識）

ORM の出発点は「struct の型情報を読み取り、テーブルとカラムの定義を生成する」ことです。Step 1 ではデータの読み書きはなし、スキーマ解析と定義生成だけを実装します。

---

## 1. struct タグとは

struct タグは Go の `reflect.StructTag` 型で、フィールドに任意のメタデータを付与できます。

```go
type User struct {
    ID   int    `orm:"pk"`
    Name string `orm:"column:name"`
    Age  int    `orm:"column:age"`
}
```

`reflect.TypeOf(User{}).Field(0).Tag.Get("orm")` で `"pk"` という文字列が取れます。ORM はこの文字列を自前でパースして、PK かどうか・カラム名は何かを判断します。

---

## 2. テーブル名の決め方

多くのORMはテーブル名を「struct 名を小文字にして複数形にしたもの」とします（`User` → `users`）。

シンプルな実装では「struct 名を小文字にして `s` を付けるだけ」で十分です。

```
User  → "users"
Post  → "posts"
```

struct 名は `reflect.TypeOf(v).Name()` で取得し、`strings.ToLower` で小文字に変換します。

---

## 3. ColumnDef（カラム定義）構造体

スキーマを表すために、フィールドごとのメタデータを保持する構造体を定義します。

```
ColumnDef:
  FieldIndex  int     // struct の何番目のフィールドか
  ColumnName  string  // DB上のカラム名
  IsPK        bool    // 主キーかどうか
```

全フィールドをループして ColumnDef のスライスを構築するのが、スキーマ生成の核心です。

---

## 4. Schema キャッシュ

スキーマ生成はリフレクションを使うため、毎回呼ぶとコストがかかります。一度生成したスキーマは `map[reflect.Type]*Schema` にキャッシュします。

```
初回 Insert(&User{}):
  → TypeOf(User) でキャッシュを検索 → 存在しない
  → スキーマを生成してキャッシュに保存
  → 以後はキャッシュから取り出す
```

---

## 📌 まとめ：Step 1 のフロー

1. `reflect.TypeOf(v)` で struct の型情報を取得する
2. struct 名からテーブル名を生成する（小文字 + "s"）
3. 全フィールドをループして `Tag.Get("orm")` を解析する
4. `ColumnDef` のスライスを構築する（IsPK・ColumnName を設定）
5. `Schema` 構造体（tableName + []ColumnDef）をキャッシュに保存する
6. `db.Schema(User{})` で結果を出力して確認する
