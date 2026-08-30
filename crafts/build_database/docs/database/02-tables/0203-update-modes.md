# 0203: 更新モード

- 元ページ: https://trialofcode.org/database/0203_update_mode/

## このステップで学ぶこと

- KV層の`Set`操作が「挿入」と「更新」を区別せず一括で扱ってしまう問題を理解する
- insert/update/upsertという3つの更新モードを導入し、SQLのCRUD操作をKV操作に正しく対応づける方法を学ぶ

## 要点

- SQLの基本操作はKV操作に素直に対応する（select→get、delete→del）が、KVの`set`は「新規作成」と「既存更新」を区別しないため、insert/updateを個別に扱いたい場合にギャップが生じる
- このギャップを埋めるため`UpdateMode`という列挙型を導入し、3つのモードを定義する
  - `ModeUpsert`（デフォルト）: 存在すれば上書き、存在しなければ新規作成
  - `ModeInsert`: キーが存在しない場合のみ挿入。既に存在する場合は失敗（false）を返す
  - `ModeUpdate`: キーが存在する場合のみ更新。存在しない場合は失敗（false）を返す
- PostgreSQLが明示的な`UPSERT`構文を導入した経緯と同様、insert/update/upsertを区別する設計はよく見られるパターンである

## 実装のポイント / 注意点

- 既存の`Set()`メソッドを直接拡張するのではなく、モードを受け取る新しい`SetEx(key, val, mode)`に処理を委譲し、`Set()`は内部で`ModeUpsert`を渡すだけにすることで後方互換性を保つ
- `SetEx`の戻り値は`(bool, error)`とし、boolは「モードの制約通りに操作が成功したか」を表す（エラーとは別の意味を持つことに注意）
- このモード分岐は次のステップ（CRUD API）でinsert/update/upsertメソッドを実装する際にそのままマッピングされるため、ここでの設計が後続実装の土台になる
