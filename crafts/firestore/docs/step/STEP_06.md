# Step 6：Resume token（前提知識）

## このステップで何が変わるか

切断・再接続したクライアントが **resume token を送ることで、切断中に見逃したイベントだけを受け取れる** ようにする。full re-fetch が不要になるのが Firestore の重要な特徴。

---

## Resume token の正体

Step 5 で `TargetChange(CURRENT)` に付けた `resumeToken` は、その時点のスナップショットバージョンを base64 エンコードしたもの:

```
resumeToken = base64("3")   →   "Mw=="
```

クライアントはこの値を保存しておき、再接続時の `AddTarget` に含める:

```json
{
  "type": "AddTarget",
  "path": "users/alice",
  "targetId": 1,
  "resumeToken": "Mw=="
}
```

---

## サーバー側の差分再送

サーバーはトークンをデコードして `sinceVersion` を得る:

```
sinceVersion = decode("Mw==")  →  3
```

そして **version > sinceVersion** のイベントだけを再送する。

---

## イベントログ（リングバッファ）

差分再送を実現するには、過去のイベントを一定期間保持する必要がある。  
実装方針: **Store の書き込み履歴をログとして保持する**。

```
EventLog
└── entries: []LogEntry (リングバッファ or スライス)

LogEntry
├── Version  SnapshotVersion
├── Path     string
├── Document *Document  (削除時は nil)
└── Type     ChangeType
```

古すぎるバージョンが来たらログに残っていないので **TargetChange(RESET)** を返す（フルリロードを促す）。

---

## TargetChange(RESET) の意味

```
クライアント: AddTarget(resumeToken="古すぎるトークン")
サーバー: TargetChange(RESET)  ← 「差分が出せない、最初からやり直して」
クライアント: resumeToken なしで再購読
```

---

## 📌 まとめ: Step 6 のフロー

1. `store.Store` に `EventLog` を追加し、`Put` / `Delete` のたびにログを追記する
2. `store.Store` に `EventsSince(version SnapshotVersion) []LogEntry` を追加する
3. `handleListen` の初期スナップショット送信フェーズを拡張する:
   - `req.ResumeToken` が空なら今まで通り `Get` + CURRENT を送る
   - `req.ResumeToken` がある場合は `sinceVersion` をデコードして `EventsSince` を呼ぶ
   - ログにない（古すぎる）バージョンなら `TargetChange(RESET)` を送って return
   - 差分イベントを順番に SSE 送信し、最後に CURRENT を送る
4. `AddTarget` JSON に `resumeToken` フィールドを追加する（Step 5 の実装を更新）
