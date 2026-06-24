# Step 1 実装ガイド：HTTP Upgrade ハンドシェイク

---

## ゴール

```bash
go run main.go

# 別ターミナルで
wscat -c ws://localhost:8080
# → Connected (press CTRL+C to quit)
```

---

## 変更するファイル

```
websocket-broker/go/
├── go.mod
├── main.go
└── internal/
    └── ws/
        └── handshake.go
```

---

## 実装手順

### 1. `go.mod` の作成

```
module websocket-broker

go 1.25.1
```

### 2. `internal/ws/handshake.go`

**Upgrade(conn net.Conn) error の内部でやること（順番どおり）**:

1. `bufio.NewReader(conn)` でリクエストを読む
2. リクエストラインを読む（`GET / HTTP/1.1\r\n`）
3. ヘッダーを `\r\n` まで 1 行ずつ読む
4. `Sec-WebSocket-Key` の値を取り出す
5. `Upgrade: websocket` が含まれているかチェックする（なければエラーを返す）
6. `Sec-WebSocket-Accept` を計算する:
   ```go
   const guid = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
   h := sha1.Sum([]byte(key + guid))
   accept := base64.StdEncoding.EncodeToString(h[:])
   ```
7. 101 レスポンスを書く:
   ```go
   fmt.Fprintf(conn,
       "HTTP/1.1 101 Switching Protocols\r\n"+
       "Upgrade: websocket\r\n"+
       "Connection: Upgrade\r\n"+
       "Sec-WebSocket-Accept: %s\r\n"+
       "\r\n", accept)
   ```
8. nil を返す（エラーなし）

必要な import:
```go
import (
    "bufio"
    "crypto/sha1"
    "encoding/base64"
    "fmt"
    "net"
    "strings"
)
```

### 3. `main.go`

**main() の内部でやること**:

1. `net.Listen("tcp", ":8080")` でポートを開く
2. ループで `listener.Accept()` する
3. 各接続を `go handleConn(conn)` に渡す

**handleConn(conn net.Conn) の内部でやること**:

1. `defer conn.Close()`
2. `ws.Upgrade(conn)` を呼ぶ
3. エラーなら return
4. `fmt.Println("WebSocket connected")` でログを出す（フレーム処理は Step 2 で追加）

---

## 実装の確認手順

```bash
# ビルド確認
go build ./...

# 起動
go run main.go

# 接続確認（wscat がなければ npm install -g wscat）
wscat -c ws://localhost:8080
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| `Error: Unexpected server response: 400` | Sec-WebSocket-Accept が正しくない | GUID の文字列をコピペし直す（スペルミスが多い） |
| `Error: Unexpected server response: 500` | ヘッダーパース中にパニック | `ReadString('\n')` 後に `strings.TrimRight(line, "\r\n")` する |
| 接続後すぐ切断される | 101 レスポンスの末尾が `\r\n\r\n` になっていない | `Sec-WebSocket-Accept: xxx\r\n` の後に `\r\n` が必要 |
