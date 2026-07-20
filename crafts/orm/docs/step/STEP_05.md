# Step 5：UPDATE / DELETE（前提知識）

PK を使って特定のレコードを更新・削除します。INSERT・SELECT と比べて「既存レコードの特定」という新しい操作が加わります。

---

## 1. PK によるレコード特定

UPDATE と DELETE はどちらも「PK が一致するレコード」を対象にします。

```
UPDATE: store["users"] の各レコードを走査して
  record["id"] == u.ID なら → そのレコードを新しい値で置き換える

DELETE: store["users"] の各レコードを走査して
  record["id"] == u.ID なら → そのレコードをスライスから除去する
```

スライスからの除去は、対象以外のレコードで新しいスライスを作る手法（filter）が最もシンプルです。

---

## 2. UPDATE の実装方針

struct の全フィールドを再度ストアに反映します。「変更されたフィールドだけ更新する（Dirty Tracking）」は複雑なため、Step 5 では全フィールドを上書きします。

```
1. PK の値を取得する
2. store[tableName] をループして PK 一致を探す
3. 一致したらレコードを全フィールド分上書きする
4. 更新されなかった場合はエラーを返す
```

---

## 3. DELETE の実装方針

PK が一致しないレコードだけを集めた新しいスライスで置き換えます。

```
filtered := []Record{}
for _, record := range store[tableName] {
    if record["id"] != pkValue {
        filtered = append(filtered, record)
    }
}
store[tableName] = filtered
```

削除前後でスライスの長さが変わらなければ、対象レコードが存在しなかったことになります。

---

## 4. PK が未設定の場合

`u.ID == 0` のまま Update・Delete を呼ぶと全件に影響するなどの危険があります。PK のゼロ値チェックを最初に行い、エラーを返します。

```
if pkValue == 0 {
    return errors.New("cannot update/delete: primary key is not set")
}
```

---

## 📌 まとめ：Step 5 のフロー

**UPDATE**:
1. 引数 struct から PK 値を取得する（ゼロ値ならエラー）
2. store[tableName] をループして PK 一致レコードを探す
3. 一致したら全フィールドを新しい値で上書きする
4. 一致がなければ「レコードが見つからない」エラーを返す

**DELETE**:
1. 引数 struct から PK 値を取得する（ゼロ値ならエラー）
2. PK が一致しないレコードだけで新しいスライスを作る
3. 元のレコード数と新しいスライスの長さを比較し、変わらなければエラーを返す
4. store[tableName] を新しいスライスで置き換える
