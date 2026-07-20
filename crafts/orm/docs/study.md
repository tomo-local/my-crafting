# 簡易ORM自作に必要な知識の整理

ORMの自作は、大きく分けて **「リフレクションによる型操作」「データモデルの設計」「クエリビルダーのパターン」「アソシエーションとN+1」** の4つの知識が必要です。

---

## 1. Goのリフレクション

### 🔍 reflect パッケージとは

コンパイル時に型が確定しない状況で、実行時に型情報や値を読み書きする仕組みです。ORMは「どんな struct でも受け取れる」汎用関数が必要なため、リフレクションが不可欠です。

```
func Insert(v interface{}) error {
    // v が User でも Post でも、フィールド名と値を動的に取得したい
}
```

### 🧱 Type と Value の2本柱

| | 役割 | 取得方法 |
|---|---|---|
| `reflect.Type` | 型情報（フィールド名・型名・タグ） | `reflect.TypeOf(v)` |
| `reflect.Value` | 実際の値（読み取り・書き込み） | `reflect.ValueOf(v)` |

```
reflect.TypeOf(User{})       → Goの型情報
  .Field(i).Name             → フィールド名 ("ID", "Name", ...)
  .Field(i).Tag.Get("orm")   → struct タグの値

reflect.ValueOf(&user).Elem()
  .Field(i).Int()            → int フィールドの値を読む
  .Field(i).SetInt(42)       → int フィールドに値を書く
```

### 🏷️ struct タグの解析

struct タグは `key:"value"` 形式で複数指定できます。

```go
type User struct {
    ID   int    `orm:"pk"`
    Name string `orm:"column:name"`
}
```

`reflect.StructTag.Get("orm")` で文字列として取得し、自前でパースします。

```
タグ文字列: "column:name,not_null"
→ split(",") → ["column:name", "not_null"]
→ split(":") → key="column", value="name"
```

### ⚠️ ポインタとElem

`interface{}` で受け取った値がポインタの場合、`reflect.ValueOf(v)` は Kind が `Ptr` になります。フィールドにアクセスするには `.Elem()` で間接参照が必要です。

```
reflect.ValueOf(&user)          → Kind: Ptr
reflect.ValueOf(&user).Elem()   → Kind: Struct ← フィールドにアクセスできる
```

---

## 2. ORM のデータモデル設計

### 🗃️ インメモリストアの構造

実データベースを使わない場合、最もシンプルな設計はテーブル名をキーにしたマップです。

```
DB
├── "users"  → []Record
│               ├── Record{1, "Alice", 30}
│               └── Record{2, "Bob",   25}
└── "posts"  → []Record
                ├── Record{1, "Hello", 1}
                └── Record{2, "World", 2}
```

各 `Record` は `map[string]interface{}` で表現すると、カラム名と値を汎用的に扱えます。

### 📐 スキーマ（テーブル定義）

struct の型情報から「どのフィールドがどのカラムに対応するか」を表すスキーマを事前に構築しておきます。

```
Schema for User:
  tableName: "users"
  columns:
    - FieldIndex: 0, ColumnName: "id",   Type: int,    IsPK: true
    - FieldIndex: 1, ColumnName: "name", Type: string, IsPK: false
    - FieldIndex: 2, ColumnName: "age",  Type: int,    IsPK: false
```

スキーマの構築はコスト（リフレクション処理）がかかるため、一度構築したら `sync.Map` などにキャッシュします。

### 🔑 主キー（PK）の扱い

INSERT 時の自動採番と UPDATE / DELETE の対象特定に PK が必要です。

```
テーブルごとの自動採番カウンター:
  "users" → 2（次のIDは3）
  "posts" → 1（次のIDは2）
```

---

## 3. クエリビルダーのパターン

### 🔗 メソッドチェーン（Builder パターン）

`Where()` を呼ぶたびに条件を積み上げ、`Find()` で実行するパターンです。

```
db.Where("age", ">", 20).Where("name", "=", "Alice").Find(&users)

内部状態（Query構造体）:
  tableName: "users"
  conditions: [
    {column: "age",  op: ">", value: 20},
    {column: "name", op: "=", value: "Alice"},
  ]
```

`Where()` はレシーバ（Builder 構造体）のコピーを返すことで、チェーンが元の状態を汚染しません。

### 🧮 条件の評価

インメモリストアの場合、全レコードを走査して条件に合致するものだけ返します。

```
全レコードをループ:
  record := Record{"id": 1, "name": "Alice", "age": 30}

  条件1: record["age"] > 20  → true
  条件2: record["name"] == "Alice" → true

  → 全条件 AND → このレコードを結果に含める
```

比較演算子（`>`、`<`、`=`、`!=`）は `switch` 文で分岐して実装します。

---

## 4. N+1 問題とEager Loading

### 🐢 N+1 問題とは

関連データを個別に取得するループがN回のクエリを発生させる問題です。

```
posts を全件取得: クエリ1回
  → posts[0].UserID = 1 → users WHERE id=1: クエリ1回
  → posts[1].UserID = 2 → users WHERE id=2: クエリ1回
  → posts[2].UserID = 1 → users WHERE id=1: クエリ1回（重複）
  合計: 1 + N 回
```

投稿が100件あれば101回のクエリが発生します。

### ⚡ Eager Loading の仕組み

関連する外部キーをまとめて収集し、IN 相当の1回のクエリで取得します。

```
Step 1: posts を全件取得（1回）
  → UserID の集合: {1, 2}

Step 2: users WHERE id IN (1, 2)（1回）
  → map[id]*User を構築

Step 3: 各 post の User フィールドにセット
  posts[0].User = userMap[1]
  posts[1].User = userMap[2]

合計: 2回（何件あっても）
```

### 🔗 アソシエーションの種類

| 種類 | 意味 | struct タグ例 |
|---|---|---|
| belongs_to | 外部キーを自分が持つ | `orm:"belongs_to:User,fk:user_id"` |
| has_many | 相手が外部キーを持つ | `orm:"has_many:Post,fk:user_id"` |

リフレクションで struct のフィールド型を確認し、ポインタなら belongs_to、スライスなら has_many と判断できます。

---

## 5. 【実装上の罠】ORM自作で必ずハマるポイント

### 🛑 reflect.Value の Settable チェック

`reflect.Value.Set` 系メソッドは、値が "settable"（ポインタ経由でアクセスされている）でないとパニックします。

```
reflect.ValueOf(user).Field(0).SetInt(1)    // パニック！
reflect.ValueOf(&user).Elem().Field(0).SetInt(1) // OK
```

**対策**: 引数を受け取る時点で必ずポインタを要求し、`Elem()` してから操作する。

### 🛑 型アサーションの失敗

インメモリストアで `interface{}` に `int` を入れ、後で `int64` として取り出すと型アサーションが失敗します。

```
record["age"] = 30          // int として格納
v := record["age"].(int64) // パニック！（int と int64 は別型）
```

**対策**: 格納時に型を統一するか、`fmt.Sprintf` や `reflect` 経由で比較する。

### 🛑 スキーマ未登録でのアクセス

`db.Insert(&user)` を呼ぶ前にスキーマが登録されていない場合のハンドリングを忘れると、`nil` アクセスでパニックします。

**対策**: スキーマのキャッシュ参照時に存在チェックを入れ、未登録なら自動登録するか明示的なエラーを返す。

### 🛑 Eager Loading のポインタ注入

`Post.User` フィールドに `*User` をセットするとき、`reflect.Value.Set` に渡す値の型が正確に一致している必要があります。型が合わない場合はパニックではなくエラーが `reflect.Value.CanSet()` で検出できます。

**対策**: セット前に `reflect.Value.Type()` と注入する値の型が一致するか確認する。
