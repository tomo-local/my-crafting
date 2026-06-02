# Step 5：DocumentChange イベントと TargetChange(CURRENT)（前提知識）

## このステップで何が変わるか

Step 4 では「書き込みが届く」だけだった。このステップで **購読開始時の初期スナップショット送信** と、「初期状態を送り終えた」ことを示す **TargetChange(CURRENT)** を追加する。

---

## 本物の Firestore のイベント順序

購読開始後にサーバーが送る順序:

```
1. 既存ドキュメントを ADDED イベントとして全送信
2. TargetChange(CURRENT, resumeToken=...)  ← 「初期状態を全部送った」印
3. 以降、変更があるたびに DocumentChange (ADDED/MODIFIED/REMOVED) を送る
```

クライアントは TargetChange(CURRENT) を受け取るまで「まだ初期状態の受信中」と判断する。

---

## イベント種別の正確な定義

| ChangeType | 発生タイミング |
|---|---|
| `ADDED` | ① ドキュメントが新規作成された ② 購読開始時に既存ドキュメントを初回送信するとき |
| `MODIFIED` | 既存ドキュメントのフィールドが更新された |
| `REMOVED` | ドキュメントが削除された |

---

## TargetChange(CURRENT) のタイミング

```
Subscribe(path)
    ↓
既存ドキュメントを ADDED で送る（0件でも構わない）
    ↓
TargetChange(CURRENT, resumeToken=<current_version>) を送る
    ↓
以降は変更が来るたびに送る
```

`resumeToken` は現時点のスナップショットバージョンを base64 エンコードしたもの（Step 6 で使う）。

---

## handleListen の拡張

`Subscribe` を呼んだ直後に「初期スナップショット送信フェーズ」を追加する:

```
1. store.List(collection) で現在のドキュメントを全取得
2. 各ドキュメントを ADDED イベントとして SSE で送る
3. store の現在バージョンを取得して TargetChange(CURRENT, resumeToken) を送る
4. Flush()
5. その後、通常のイベントループへ
```

`collection` は `path` を `/` で分割した最初の要素 (`users/alice` → `users`)。  
ただしドキュメントを直接 Watch している場合（`users/alice`）は対象の1件だけ送る。

---

## 📌 まとめ: Step 5 のフロー

1. `watcher.Event` に `TargetChange` 種別を追加するか、別の型として定義する
2. `store.Store` に `CurrentVersion() SnapshotVersion` メソッドを追加する
3. `handleListen` に初期スナップショット送信フェーズを追加する:
   - `store.Get(path)` でドキュメントが存在すれば ADDED イベントを送る
   - `TargetChange(CURRENT)` を送る
4. SSE の JSON フォーマットに `type` フィールドを追加して `DocumentChange` と `TargetChange` を区別できるようにする
