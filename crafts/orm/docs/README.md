# ゼロから構築する簡易ORM・全体学習プロセス

`database/sql` ドライバや外部ライブラリを使わず、Goのリフレクション（`reflect` パッケージ）と標準ライブラリのみを使って、struct ↔ データストアのマッピング・クエリ生成・eager loading までを実装することで、ORM の仕組みを深く理解していくプロセスです。

---

## 全体ロードマップ

| Step | テーマ | ゴール | 状態 |
|------|--------|--------|------|
| 1 | 型マッピングとスキーマ定義 | struct タグを解析してテーブル定義を生成できること | 🚧 |
| 2 | INSERT | struct の値を読み取ってレコードを保存できること | 🚧 |
| 3 | SELECT / Find | テーブルから読み出して struct に値をマッピングできること | 🚧 |
| 4 | クエリビルダー（WHERE） | 条件付きで絞り込み検索できること | 🚧 |
| 5 | UPDATE / DELETE | PK を使って特定レコードを更新・削除できること | 🚧 |
| 6 | Associations & Eager Loading | N+1 なしで関連レコードをまとめて取得できること | 🚧 |

---

## 各ステップの詳細

### Step 1: 型マッピングとスキーマ定義

ORM の核心は「Go の型とデータストアの型をどう対応させるか」です。struct タグ（`orm:"column:name,pk"`）を解析してテーブル定義を動的に生成する仕組みを作ります。

**学習内容**
- `reflect.TypeOf` / `reflect.ValueOf` の基本
- struct タグ（`reflect.StructTag`）の解析方法
- テーブル名・カラム名・型の対応表を作る設計
- インメモリストアの基本設計（`map[string][]Record`）

**実験ゴール**

```go
type User struct {
    ID   int    `orm:"pk"`
    Name string `orm:"column:name"`
    Age  int    `orm:"column:age"`
}

db := orm.New()
schema := db.Schema(User{})
fmt.Println(schema) // table: users, columns: [id, name, age]
```

---

### Step 2: INSERT

struct フィールドの値をリフレクションで読み取り、インメモリストアに保存します。

**学習内容**
- `reflect.Value.Field(i)` でフィールド値を取得する方法
- 自動採番（Auto Increment）の実装
- `interface{}` でフィールド値を汎用的に扱う
- テーブルが存在しない場合のエラーハンドリング

**実験ゴール**

```go
u := &User{Name: "Alice", Age: 30}
err := db.Insert(u)
fmt.Println(u.ID) // → 1（自動採番で ID が埋まること）

u2 := &User{Name: "Bob", Age: 25}
err = db.Insert(u2)
fmt.Println(u2.ID) // → 2
```

---

### Step 3: SELECT / Find

テーブルの全レコードを走査し、struct のポインタに値をスキャンして返します。

**学習内容**
- `reflect.Value.Set` / `reflect.Value.SetInt` 等でフィールドに値をセットする方法
- `reflect.New` でゼロ値の struct を動的に生成する
- スライス（`[]*User`）へのリフレクション経由での append
- カラム名 → フィールドのインデックスを逆引きするマップ

**実験ゴール**

```go
var users []*User
err := db.FindAll(&users)
for _, u := range users {
    fmt.Println(u.ID, u.Name) // → 1 Alice, 2 Bob
}
```

---

### Step 4: クエリビルダー（WHERE）

`Where()` を連鎖させて条件を積み上げ、最終的に `Find()` で実行するビルダーパターンを実装します。

**学習内容**
- メソッドチェーン可能なビルダー構造体の設計
- 条件（`Condition`）の表現方法（カラム名・演算子・値）
- インメモリストアを走査して条件に合致するレコードを返すフィルタリング
- 複数条件の AND 結合

**実験ゴール**

```go
var users []*User
err := db.Where("age", ">", 20).Where("name", "=", "Alice").Find(&users)
fmt.Println(len(users)) // → 1
fmt.Println(users[0].Name) // → Alice
```

---

### Step 5: UPDATE / DELETE

PK（主キー）を基準に特定のレコードを更新・削除します。

**学習内容**
- PK フィールドをタグから特定する方法
- UPDATE: フィールド値の変更をストアに反映する
- DELETE: ストアからレコードを除去する
- PK が未設定（ゼロ値）の場合のエラーハンドリング

**実験ゴール**

```go
// UPDATE
u.Name = "Alicia"
err := db.Update(u)

// 確認
var found []*User
db.Where("id", "=", 1).Find(&found)
fmt.Println(found[0].Name) // → Alicia

// DELETE
err = db.Delete(u)
db.FindAll(&found)
fmt.Println(len(found)) // → 1（Bob のみ）
```

---

### Step 6: Associations & Eager Loading

`belongs_to` / `has_many` の関連定義を struct タグで表現し、N+1 なしでまとめて取得します。

**学習内容**
- 関連の種類（has_many / belongs_to）と外部キーの仕組み
- N+1 問題の再現と計測（クエリ回数のカウント）
- Preload の実装: 関連 ID をまとめて IN 相当で取得し、ポインタに注入する
- リフレクションで struct のフィールドにスライスをセットする

**実験ゴール**

```go
type Post struct {
    ID     int    `orm:"pk"`
    Title  string `orm:"column:title"`
    UserID int    `orm:"column:user_id"`
}

// N+1 あり（悪い例）
db.FindAll(&posts)
for _, p := range posts {
    db.Where("id", "=", p.UserID).Find(&p.User) // ← N回クエリが走る
}

// Eager Loading（良い例）
db.Preload("User").FindAll(&posts)
// → User の取得が1回にまとまること
// → クエリカウンターで確認できること
```

---

## 学習を進める上でのアドバイス

1. **リフレクションは最初に全部理解しなくてよい**
   Step 1 で `reflect.TypeOf` の基本さえ掴めれば、後のステップで `Set`・`Field`・`New` を都度調べながら進められます。Go の公式ドキュメント（[The Laws of Reflection](https://go.dev/blog/laws-of-reflection)）を手元に置いておくと便利です。

2. **クエリカウンターを早めに仕込む**
   Step 6 で N+1 を確認するために、ストアへのアクセスを計測できる仕組み（シンプルなカウンター変数）を Step 2 の段階で入れておくと、後で「本当に減ったか」が検証しやすいです。

3. **型アサーションと reflect のどちらを使うか意識する**
   `int` 型のフィールドに値をセットするとき、`reflect.Value.SetInt()` を使う方法と `interface{}` にキャストして比較する方法があります。両者を混在させると混乱するので、方針を決めてから進めましょう。

4. **実際のORMの実装を後で読む**
   完走後に [GORM](https://github.com/go-gorm/gorm) や [ent](https://github.com/ent/ent) のコードを読むと「自分が実装したことと何が違うか」が明確になります。特に GORM の `Statement` 構造体はクエリビルダーの参考になります。

---

## 完走後の次のステップ

- **トランザクション**: 複数の操作をアトミックに実行し、失敗時にロールバックする仕組み
- **マイグレーション**: スキーマのバージョン管理と差分適用
- **フック (Callbacks)**: BeforeCreate / AfterUpdate などのライフサイクルフック
- **バリデーション**: Insert/Update 前の値検証
- **実データベースへの接続**: `database/sql` の `driver.Driver` インターフェースを実装して SQLite に繋ぐ
