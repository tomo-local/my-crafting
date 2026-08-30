# 0505: SELECT/UPDATEでの式サポート

- 元ページ: https://trialofcode.org/database/0505_select_expr/

## このステップで学ぶこと

- ここまで作った式パーサ・評価器を、実際のSQL実行エンジン（SELECT・UPDATE文）に統合する
- `select a * 4 - b, d + c from t` のように、カラム名だけでなく計算式を選択リストに書けるようにする

## 要点

- `StmtSelect` の `cols` フィールドを `[]string` から `[]interface{}` に変更し、単純な列名・演算式（`ExprBinOp`, `ExprUnOp`）・リテラル（`*Cell`）のいずれも保持できるようにする
- `StmtUpdate` には `ExprAssign` 型を追加し、「列名 = 式」という代入形式（`set a = a - b`）を表現する
- パーサ側の変更
  - `parseSelect()` 内で単純な `tryName()` の代わりに `parseExpr()` を呼ぶ
  - `parseUpdate()` で `parseEqual()` を `parseAssign()` に置き換える
- 実行エンジン側の変更
  - `execSelect()` と `execUpdate()` の両方で `evalExpr()` を呼び出し、パース済みの式を実際の値に計算する

## 実装のポイント / 注意点

- `[]string` → `[]interface{}` への型変更は既存コードの互換性に影響するため、`cols` を参照している箇所を漏れなく洗い出して修正する必要がある
- 例: `select a * 4 - b, d + c from t where d = 123;` / `update t set a = a - b, b = a, c = d + c where d = 123;` のようなクエリが動くようになる
- このステップではWHERE句自体はまだ従来通りの単純な形式（等価条件）のみを想定しており、WHERE句への式統合は次ステップ（0506）以降
