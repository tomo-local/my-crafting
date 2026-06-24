# Step 5 実装ガイド：Ping/Pong ハートビート

---

## ゴール

```bash
wscat -c ws://localhost:8080

# 何も送らず放置する
# 15 秒後にサーバー側のログに以下が出て切断されること:
# "client timed out: 127.0.0.1:xxxxx"
```

---

## 変更するファイル

```
websocket-broker/go/
├── main.go            # handleConn にハートビートを追加
└── internal/
    └── ws/
        ├── handshake.go
        ├── frame.go
        └── hub.go
```

---

## 実装手順

### 1. `main.go` に定数を追加

```go
const (
    pingInterval = 10 * time.Second
    pongTimeout  =  5 * time.Second
)
```

---

### 2. handleConn にハートビートを追加

**追加する処理（Step 4 の handleConn に加える）**:

1. Ticker と Timer を作る:
   ```go
   ticker := time.NewTicker(pingInterval)
   timer  := time.NewTimer(pingInterval + pongTimeout)
   done   := make(chan struct{})
   ```

2. defer に停止処理を追加:
   ```go
   defer func() {
       ticker.Stop()
       timer.Stop()
       close(done)
       // ... 既存の Unsubscribe / close(client.send) / conn.Close()
   }()
   ```

3. Ping 送信 goroutine を起動:
   ```go
   go func() {
       for {
           select {
           case <-ticker.C:
               client.send <- nil  // nil を Ping のシグナルとして使う
           case <-done:
               return
           }
       }
   }()
   ```

4. タイムアウト goroutine を起動:
   ```go
   go func() {
       select {
       case <-timer.C:
           slog.Info("client timed out", "addr", conn.RemoteAddr())
           conn.Close()
       case <-done:
       }
   }()
   ```

5. 受信ループで Pong を処理:
   ```go
   if frame.Opcode == 0xA {  // Pong
       timer.Reset(pingInterval + pongTimeout)
       continue
   }
   ```

---

### 3. 送信 goroutine で nil（Ping）を処理

Step 4 の送信 goroutine を更新:

```go
go func() {
    for msg := range client.send {
        var f ws.Frame
        if msg == nil {
            f = ws.Frame{FIN: true, Opcode: 0x9}  // Ping
        } else {
            f = ws.Frame{FIN: true, Opcode: 0x1, Payload: msg}
        }
        if err := ws.WriteFrame(conn, f); err != nil {
            return
        }
    }
}()
```

---

## 実装の確認手順

```bash
go build ./...
go run main.go

# ターミナル B（何も送らずに待つ）
wscat -c ws://localhost:8080
# 15 秒後: サーバーに "client timed out" が出て wscat が切断される

# wscat は自動で Pong を返すため、通常の使用では切断されないことも確認できる
```

---

## よくあるハマりポイント

| 症状 | 原因 | 対処 |
|---|---|---|
| タイムアウトしない | timer が毎回 Reset されている | Pong opcode の分岐（0xA）を確認 |
| 接続後すぐ切断される | Timer の初期値が短すぎる | `pingInterval + pongTimeout` を初期値に設定する |
| goroutine が残り続ける | `close(done)` を呼んでいない | defer で確実に呼ぶ |
| send on closed channel でパニック | close(done) 後に send チャネルへ書き込んでいる | Ping goroutine が `done` を先に受信するよう select の順序を確認 |
