# Step 6 実装ガイド：Resume token

## ゴール

```bash
# 1. 購読開始、resumeToken を記録
curl -N -X POST http://localhost:8080/v1/listen \
  -d '{"type":"AddTarget","path":"users/alice","targetId":1}'
# → data: {"type":"TargetChange","changeType":"CURRENT","resumeToken":"MQ=="}

# 2. 接続を切る (Ctrl+C)

# 3. 切断中に書き込む
curl -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -d '{"name":"Alice","age":31}'
curl -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -d '{"name":"Alice","age":32}'

# 4. resumeToken 付きで再接続 → 切断中の変更だけ届くこと
curl -N -X POST http://localhost:8080/v1/listen \
  -d '{"type":"AddTarget","path":"users/alice","targetId":1,"resumeToken":"MQ=="}'
# → data: {"type":"DocumentChange","changeType":"MODIFIED",...,"version":2}
# → data: {"type":"DocumentChange","changeType":"MODIFIED",...,"version":3}
# → data: {"type":"TargetChange","changeType":"CURRENT","resumeToken":"Mw=="}
```

---

## 変更するファイル

```
go/
└── internal/
    └── store/
        └── store.go  （EventLog + EventsSince を追加）
    └── server/
        └── server.go  （handleListen の初期フェーズを拡張）
```

---

## `store.go` への追加

### 1. LogEntry と EventLog

```go
type LogEntry struct {
    Version    SnapshotVersion
    Path       string
    Document   *Document // REMOVED のときは nil
    ChangeType watcher.ChangeType
}
```

`Store` に `log []LogEntry` フィールドを追加する。

### 2. Put / Delete でのログ追記

`Put` の末尾に:
```go
s.log = append(s.log, LogEntry{Version: s.version, Path: path, Document: s.data[path], ChangeType: watcher.Added /* or Modified */})
```

`Delete` の末尾に:
```go
s.log = append(s.log, LogEntry{Version: s.version, Path: path, Document: nil, ChangeType: watcher.Removed})
```

### 3. EventsSince

内部でやること（順番どおり）:
1. `s.mu.RLock()` / `defer s.mu.RUnlock()`
2. `s.log` をループして `entry.Version > since && entry.Path == path` のものを集める
3. スライスを返す

```go
func (s *Store) EventsSince(path string, since SnapshotVersion) ([]LogEntry, bool) {
    // bool は "since が古すぎてログに残っていない" かどうか
```

ログの最古バージョンより `since` が小さい場合は `false` を返す（RESET が必要）。

---

## `handleListen` の拡張

### resumeToken のデコード

```go
func decodeResumeToken(token string) (store.SnapshotVersion, error) {
    b, err := base64.StdEncoding.DecodeString(token)
    if err != nil {
        return 0, err
    }
    v, err := strconv.ParseUint(string(b), 10, 64)
    return store.SnapshotVersion(v), err
}
```

### 初期スナップショット送信フェーズの分岐

```
if req.ResumeToken == "" {
    // Step 5 までの処理（Get + ADDED + CURRENT）
} else {
    sinceVersion = decode(req.ResumeToken)
    entries, ok = store.EventsSince(path, sinceVersion)
    if !ok {
        // TargetChange(RESET) を送って return
    }
    for each entry in entries {
        // DocumentChange を SSE で送る
    }
    // TargetChange(CURRENT, new resumeToken) を送る
}
```

---

## 実装の確認手順

```bash
go build ./go/...
go test ./go/internal/store/...

go run ./go/main.go &

# 購読 → CURRENT の resumeToken を控える
curl -N -X POST http://localhost:8080/v1/listen \
  -d '{"type":"AddTarget","path":"users/alice","targetId":1}' &

curl -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -d '{"name":"Alice","age":30}'
# → resumeToken: "MQ==" などが届く

# 書き込みを追加
curl -X PUT http://localhost:8080/v1/collections/users/documents/alice \
  -d '{"name":"Alice","age":31}'

# 再接続で差分のみ確認
curl -N -X POST http://localhost:8080/v1/listen \
  -d '{"type":"AddTarget","path":"users/alice","targetId":1,"resumeToken":"MQ=="}'
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| 再接続後に全イベントが届く | `>` ではなく `>=` で比較している | `entry.Version > since` にする |
| 再接続後に何も届かない | `path` フィルタを忘れている | `entry.Path == path` も条件に加える |
| RESET が来ない | `EventsSince` が常に `true` を返している | ログの最古バージョンと `since` を比較する |
| base64 デコードエラー | クライアントが token を URL エンコードしてしまった | curl では `-d` で送る、`+` を `%2B` にエスケープしない |
