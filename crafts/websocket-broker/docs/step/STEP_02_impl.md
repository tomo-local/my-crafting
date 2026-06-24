# Step 2 実装ガイド：フレームの送受信

---

## ゴール

```bash
wscat -c ws://localhost:8080
> hello
# サーバー側ターミナルに "Received: hello" が表示されること
```

---

## 変更するファイル

```
websocket-broker/go/
├── main.go
└── internal/
    └── ws/
        ├── handshake.go   # Step 1 から変更なし
        └── frame.go       # 新規追加
```

---

## 実装手順

### 1. `internal/ws/frame.go`

**Frame 構造体**:
```go
type Frame struct {
    FIN     bool
    Opcode  byte
    Payload []byte
}
```

**ReadFrame(r io.Reader) (Frame, error) の内部でやること（順番どおり）**:

1. `io.ReadFull(r, buf[:2])` で 2 バイト読む
2. バイト 1 から `fin`（`buf[0] & 0x80`）と `opcode`（`buf[0] & 0x0F`）を取り出す
3. バイト 2 から `masked`（`buf[1] & 0x80`）と `payloadLen`（`buf[1] & 0x7F`）を取り出す
4. `payloadLen` に応じて実際の長さを確定する:
   - `126`: `io.ReadFull` で 2 バイト読み `binary.BigEndian.Uint16` で変換
   - `127`: `io.ReadFull` で 8 バイト読み `binary.BigEndian.Uint64` で変換
5. `masked=true` なら `io.ReadFull` で 4 バイトの masking key を読む
6. `io.ReadFull` で payload を実際の長さ分読む
7. `masked=true` なら XOR でアンマスクする:
   ```go
   for i := range payload {
       payload[i] ^= maskKey[i%4]
   }
   ```
8. `Frame{FIN: fin, Opcode: opcode, Payload: payload}` を返す

必要な import:
```go
import (
    "encoding/binary"
    "io"
)
```

---

### 2. `main.go` の handleConn を更新

**handleConn() の内部でやること**:

1. `ws.Upgrade(conn)` でハンドシェイク
2. ループに入る
3. `ws.ReadFrame(conn)` でフレームを読む
4. `frame.Opcode == 0x8`（Close）なら break
5. `fmt.Printf("Received: %s\n", frame.Payload)` でログ出力

---

## 実装の確認手順

```bash
go build ./...
go run main.go

# 別ターミナル
wscat -c ws://localhost:8080
> hello world
# サーバー側: Received: hello world
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| 文字化けした文字列が表示される | アンマスク処理のインデックスが `i % 4` でない | `maskKey[i%4]` を確認 |
| パースが止まる / short read | `conn.Read` が指定バイト数を保証しない | `io.ReadFull` に変える |
| `payloadLen` の読み取り後にクラッシュ | 126/127 の分岐で byte をそのまま使っている | `encoding/binary` の `BigEndian.Uint16` / `Uint64` を使う |
