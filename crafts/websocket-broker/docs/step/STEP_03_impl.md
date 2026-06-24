# Step 3 実装ガイド：エコーサーバー

---

## ゴール

```bash
wscat -c ws://localhost:8080
> hello
< hello   # 送ったメッセージがそのまま返ってくること
> goodbye
< goodbye
```

---

## 変更するファイル

```
websocket-broker/go/
├── main.go            # handleConn を更新
└── internal/
    └── ws/
        ├── handshake.go
        └── frame.go   # WriteFrame を追加
```

---

## 実装手順

### 1. `frame.go` に WriteFrame を追加

**WriteFrame(w io.Writer, f Frame) error の内部でやること**:

1. バイト 1 を組み立てる:
   ```go
   b0 := byte(0x80) | f.Opcode  // FIN=1 + opcode
   ```
2. payload 長に応じてヘッダーバイトを決める（まず 125 バイト以下のみ対応）:
   ```go
   b1 := byte(len(f.Payload))  // MASK=0、長さをそのまま
   ```
3. `w.Write([]byte{b0, b1})` でヘッダーを書く
4. `w.Write(f.Payload)` でペイロードを書く

> **Close フレーム用**: `WriteFrame(w, Frame{FIN: true, Opcode: 0x8})` で payload なし Close を送れる（`len(nil) == 0`）。

---

### 2. `main.go` の handleConn を更新

**handleConn() の内部でやること**:

1. `ws.Upgrade(conn)` でハンドシェイク
2. ループに入る
3. `ws.ReadFrame(conn)` でフレームを読む
4. `frame.Opcode == 0x8`（Close）なら:
   - `ws.WriteFrame(conn, ws.Frame{FIN: true, Opcode: 0x8})` を返す
   - break
5. `frame.Opcode == 0x1` または `0x2` なら:
   - `ws.WriteFrame(conn, ws.Frame{FIN: true, Opcode: frame.Opcode, Payload: frame.Payload})` でエコー

---

## 実装の確認手順

```bash
go build ./...
go run main.go

wscat -c ws://localhost:8080
> hello
< hello
> 日本語も送れる
< 日本語も送れる
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| wscat が受信後すぐ切断する | Close フレームを返していない | opcode==0x8 の分岐で Close を送り返してから break する |
| 送り返した文字が文字化けする | 送信時に payload を再マスクしている | サーバー→クライアントはマスク不要（b1 の最上位ビットは 0 のまま） |
| 長いメッセージが途切れる | payload length を 125 バイト以下に制限している | WriteFrame に 126/127 の分岐を追加する |
