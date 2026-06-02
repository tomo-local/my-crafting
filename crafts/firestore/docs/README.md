# Firestore 学習ロードマップ

Firestore のリアルタイム同期の仕組みを段階的に実装しながら理解する。  
WebSocket との違いを意識しながら、gRPC Listen ストリームの本質を学ぶ。

## 全体ロードマップ

| Step | テーマ | ゴール |
|---|---|---|
| 1 | In-memory Document Store | コレクション/ドキュメント構造と CRUD、スナップショットバージョンを実装する |
| 2 | Watcher（pub/sub） | パスベースの変更通知を受け取る購読機構を実装する |
| 3 | HTTP サーバー + CRUD エンドポイント | REST API でドキュメントを操作し、書き込みが Watcher に伝わることを確認する |
| 4 | Listen ストリーム（SSE） | 長命接続で `AddTarget` / `RemoveTarget` を受け付け、変更をプッシュする |
| 5 | DocumentChange イベント | `ADDED` / `MODIFIED` / `REMOVED` + `TargetChange(CURRENT)` を正しい順序で配信する |
| 6 | Resume token | 切断再接続時にトークンで差分のみ再送する |

---

## Step 1: In-memory Document Store

### 学習内容
- Firestore のコレクション/ドキュメント階層（パスで一意に識別する）
- スナップショットバージョン（単調増加する uint64）が書き込みに付与される仕組み
- Go の `sync.RWMutex` を使った並行安全な読み書き

### 実験ゴール
```bash
# CRUD が正しく動くことをユニット的に確認（まだ HTTP なし）
go test ./go/internal/store/...
```

---

## Step 2: Watcher（pub/sub）

### 学習内容
- パス → チャネルのマッピング（`map[string][]chan Event`）
- 書き込み時に該当パスの購読者へ `DocumentChange` イベントを送る
- goroutine リークを防ぐ購読解除

### 実験ゴール
```bash
go test ./go/internal/watcher/...
```

---

## Step 3: HTTP サーバー + CRUD エンドポイント

### 学習内容
- `net/http` の `ServeMux` でルーティング
- `PUT /v1/collections/{col}/documents/{doc}` で書き込み → Store + Watcher の両方に伝える
- `GET` / `DELETE` の実装

### 実験ゴール
```bash
curl -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","age":30}'
# → {"status":"ok","version":1}

curl http://localhost:8080/v1/collections/users/documents/alice
# → {"name":"Alice","age":30}
```

---

## Step 4: Listen ストリーム（SSE）

### 学習内容
- gRPC が使えない環境での代替: **Server-Sent Events（SSE）**
- `Transfer-Encoding: chunked` で接続を保持する仕組み
- クライアントが `AddTarget` / `RemoveTarget` を送る JSON プロトコル設計

### 実験ゴール
```bash
# ターミナル 1: SSE で購読開始
curl -N -X POST http://localhost:8080/v1/listen \
  -H "Content-Type: application/json" \
  -d '{"type":"AddTarget","path":"users/alice","targetId":1}'

# ターミナル 2: 書き込む → ターミナル 1 に DocumentChange が届く
curl -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -d '{"name":"Alice","age":31}'
```

---

## Step 5: DocumentChange イベント

### 学習内容
- イベント種別: `ADDED`（新規）/ `MODIFIED`（更新）/ `REMOVED`（削除）
- `TargetChange(CURRENT)` — 「現在の状態を全部送り終えた」マーカー
- イベントの順序保証（バージョン順に届けること）

### 実験ゴール
```bash
# 購読開始直後に既存ドキュメントが ADDED で届くこと
# 更新時に MODIFIED、削除時に REMOVED が届くこと
# TargetChange(CURRENT) が初回配信の末尾に届くこと
```

---

## Step 6: Resume token

### 学習内容
- Resume token = スナップショットバージョンの不透明なラッパー
- 切断後の再接続時に `AddTarget` + `resumeToken` を送ることで差分のみ再送
- イベントログ（バージョン付きリングバッファ）による差分再送実装

### 実験ゴール
```bash
# 1. 購読開始、resume token を記録
# 2. 接続を切る
# 3. いくつか書き込む
# 4. resume token 付きで再接続 → 切断中の変更だけが届くこと
curl -N -X POST http://localhost:8080/v1/listen \
  -d '{"type":"AddTarget","path":"users/alice","targetId":1,"resumeToken":"<token>"}'
```

---

## 学習を進める上でのアドバイス

- **Step 1〜2 はユニットテストで動作確認** する。HTTP を立ち上げる前に Store と Watcher の挙動を固める
- **SSE は `Flusher` インターフェースを使う**。`http.ResponseWriter` を `http.Flusher` にアサートして都度 `Flush()` しないとバッファに溜まったまま届かない
- **goroutine リーク**に注意。クライアント切断を `r.Context().Done()` で検知して購読解除する
- Firestore の本物の実装は Proto3 + gRPC だが、このクラフトでは **JSON over SSE** で動作を模倣する

## 完走後の次のステップ

- コレクションクエリの購読（フィールド条件付き）
- 複数クライアント間の同時書き込みと競合解決
- 本物の gRPC で実装し直す（`google.golang.org/grpc`）
