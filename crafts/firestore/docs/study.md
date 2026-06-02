# ベース知識

Firestore リアルタイム実装に必要な概念をまとめる。

---

## 📁 1. Firestore のデータモデル

### コレクション / ドキュメント階層

Firestore はフラットなキー/バリューではなく **コレクション → ドキュメント** のツリー構造を持つ。

```
/
├── users/          ← コレクション
│   ├── alice       ← ドキュメント（JSON オブジェクト）
│   └── bob
└── posts/
    └── hello-world
```

**パス** でドキュメントを一意に識別する: `users/alice`  
Watch の単位もパスで指定する（ドキュメント単体 or コレクション全体）。

---

## 📸 2. スナップショットバージョン

### なぜバージョンが必要か

Firestore はリアルタイム配信を保証するために、すべての書き込みに **単調増加するバージョン番号** を付与する。

```
write "alice" → version: 1
write "bob"   → version: 2
write "alice" → version: 3
```

クライアントが最後に受け取ったバージョンを覚えておけば、**再接続時に「version > N のイベントだけ送って」** とサーバーに要求できる。これが Resume token の本体。

### 実装上の表現

```
type SnapshotVersion uint64
```

本物の Firestore は `google.protobuf.Timestamp` を使うが、簡易実装では uint64 で十分。

---

## 📡 3. gRPC ストリーミング vs SSE

### gRPC の Listen RPC

本物の Firestore は HTTP/2 上の **双方向ストリーミング RPC** を使う。

```
Client ──ListenRequest──▶ Server
Client ◀──ListenResponse── Server (push, multiple times)
Client ──ListenRequest──▶ Server (AddTarget / RemoveTarget)
```

1本のコネクションで **クライアントとサーバーが同時に複数のメッセージを送り合える**。

### このクラフトでの代替: SSE

gRPC は標準ライブラリで実装が困難なため、**Server-Sent Events（SSE）** で代替する。

```
Client ──POST /v1/listen──▶ Server   (接続確立 + AddTarget を body で送る)
Client ◀──data: {...}\n\n─── Server  (イベントプッシュ、接続は維持)
```

SSE は HTTP の上に乗る一方向プッシュ。双方向性は「CRUD は別の HTTP リクエスト」で代替する。

**SSE のヘッダー**:
```http
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**SSE のメッセージ形式**:
```
data: {"type":"DocumentChange","path":"users/alice","changeType":"MODIFIED","data":{...},"version":3}\n\n
```
`\n\n`（空行 2 つ）がメッセージの区切り。

### `http.Flusher` の必要性

```
ResponseWriter
    └── Flusher （バッファをすぐに送り出すインターフェース）
```

通常の HTTP レスポンスはレスポンス完了時にまとめて送られる。SSE では **書き込むたびに `Flush()` を呼ぶ** ことでリアルタイムにデータが届く。

---

## 🔔 4. Watch 機構（pub/sub）

### Watch target

クライアントが「このパスの変更を教えて」と宣言するもの。

```json
{"type": "AddTarget", "path": "users/alice", "targetId": 1}
```

`targetId` はクライアントが管理する識別子。複数のパスを同時に購読するときに使う。

### pub/sub の構造

```
Watcher
├── subscriptions: map[path]→[]chan Event
└── Publish(path, event)
    └── 該当パスの全チャネルに event を送る
```

チャネルは **goroutine で非同期に読む**。チャネルが詰まると Publish が block するため、バッファサイズや `select` + `default` のドロップ戦略を検討する。

---

## 📦 5. DocumentChange イベント

### イベント種別

| type | 発生条件 |
|---|---|
| `ADDED` | ドキュメントが新規作成された、または購読開始時に既存ドキュメントを初回送信 |
| `MODIFIED` | 既存ドキュメントが更新された |
| `REMOVED` | ドキュメントが削除された |

### TargetChange(CURRENT)

購読開始時、サーバーは「現在の状態をすべて送り終えた」ことを示す特別なイベントを送る。

```json
{"type": "TargetChange", "changeType": "CURRENT", "targetId": 1, "resumeToken": "AAAB"}
```

クライアントはこれを受け取って初めて「初期状態の受信が完了した」と判断できる。

---

## 🔑 6. Resume token

### 仕組み

```
接続確立
  └── TargetChange(CURRENT, resumeToken="AAAB") を受信
        └── クライアントがトークンを保存

切断

再接続
  └── AddTarget(path, resumeToken="AAAB") を送信
        └── サーバーは version > decode("AAAB") のイベントのみ再送
```

### サーバー側の実装方針

- イベントログを **バージョン付きリングバッファ** として保持する
- `resumeToken` = `base64(version)` のような単純なエンコードでよい
- バッファに残っていない古すぎる version が来たら **RESET** を返す（再購読を促す）

---

## ⚠️ 実装上の罠・注意点

| 罠 | 原因 | 対処 |
|---|---|---|
| SSE がまとめて届く | `Flush()` を呼んでいない | `w.(http.Flusher).Flush()` を毎回呼ぶ |
| goroutine リーク | クライアント切断を検知していない | `r.Context().Done()` で購読解除 |
| チャネルが詰まって Publish が block | バッファなしチャネル + 遅い subscriber | バッファ付きチャネルまたは `select` + `default` でドロップ |
| 再接続後に重複配信 | Resume token のバージョン比較が `>=` になっている | `>` にする（token のバージョン自体は既に受信済み） |
| TargetChange(CURRENT) が来ない | 初期スナップショットを送る実装を忘れた | 購読開始時に既存ドキュメントを ADDED で全送信してから CURRENT を送る |
