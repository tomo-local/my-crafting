# Step 6：Associations & Eager Loading（前提知識）

Step 1〜5 で CRUD の基盤が整いました。Step 6 では「関連する struct をどう取得するか」と、「N+1 問題をなぜ Eager Loading が解決できるか」を理解します。

---

## 1. アソシエーションの表現

ORM では struct のフィールドが他の struct を指す場合、それが「リレーション」です。

```go
type Post struct {
    ID     int    `orm:"pk"`
    Title  string `orm:"column:title"`
    UserID int    `orm:"column:user_id"`
    User   *User  `orm:"belongs_to:User,fk:user_id"`  // ← アソシエーション
}

type User struct {
    ID    int     `orm:"pk"`
    Name  string  `orm:"column:name"`
    Posts []*Post `orm:"has_many:Post,fk:user_id"`  // ← アソシエーション
}
```

| タグ | 意味 | 外部キーの場所 |
|---|---|---|
| `belongs_to:User,fk:user_id` | Post が User の FK を持つ | Post.UserID |
| `has_many:Post,fk:user_id` | User の ID が Post に参照される | Post.UserID |

---

## 2. N+1 問題の再現

`Preload` を使わない場合、ループの中でクエリが走ります。

```
posts 全件取得（クエリ 1回）
posts[0].UserID = 1 → users WHERE id=1（クエリ 1回）
posts[1].UserID = 2 → users WHERE id=2（クエリ 1回）
posts[2].UserID = 1 → users WHERE id=1（クエリ 1回・重複）
→ 合計 1 + N 回
```

クエリカウンターをストアアクセス時にインクリメントしておくと、実際に何回アクセスが走ったかを計測できます。

---

## 3. Eager Loading の仕組み

Preload は「関連データを 1 度の操作でまとめて取得する」アプローチです。

```
Step 1: posts 全件取得（1回）
  → UserID のセットを収集: {1, 2}

Step 2: users WHERE id IN (1, 2)（1回）
  → map[int]*User を構築: {1: Alice, 2: Bob}

Step 3: 各 post に User をセットする
  posts[0].User = userMap[1]  // ← クエリなし
  posts[1].User = userMap[2]  // ← クエリなし
  posts[2].User = userMap[1]  // ← クエリなし（重複でも 0 コスト）
→ 合計 2回（件数に依存しない）
```

IN 相当の「IDリストで一括取得」は、Step 4 のクエリビルダーで表現すると:

```
db.Where("id", "in", []int{1, 2}).Find(&users)
```

`"in"` 演算子の実装として、`value` がスライスかどうかを `reflect.TypeOf(val).Kind() == reflect.Slice` で判定します。

---

## 4. リフレクションでのポインタフィールドへの注入

`post.User` フィールドに `*User` を動的にセットする:

```
field := reflect.ValueOf(post).Elem().FieldByName("User")
field.Set(reflect.ValueOf(userPtr))
```

ただし `field.Type()` と `reflect.ValueOf(userPtr).Type()` が一致している必要があります。型の不一致でパニックが起きるため、注入前に `CanSet()` と型チェックを行います。

---

## 📌 まとめ：Step 6 のフロー

1. `db.Preload("User").FindAll(&posts)` のインターフェースを設計する
2. `FindAll` 実行後に preload フィールド名を使ってアソシエーションを解決する
3. 解決手順:
   1. `Post.User` フィールドのタグから外部キー（`user_id`）と関連テーブルを特定する
   2. 全 post の `user_id` 値を収集してユニーク化する
   3. `WHERE id IN (...)` 相当で関連レコードをまとめて取得する
   4. `map[pkValue]*User` を構築して各 post に注入する
4. クエリカウンターでアクセス回数が「Preload なし: 1+N 回」→「Preload あり: 2 回」に減ることを確認する
