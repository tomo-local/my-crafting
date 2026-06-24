# Step 5：Ping/Pong ハートビート（前提知識）

---

## このステップで何が変わるか

接続したまま無通信が続いた場合、クライアントが実際に生きているか確認する仕組みを追加します。RFC 6455 の Ping/Pong フレームを使い、応答がなければ接続を切断します。

---

## Ping/Pong フレームの仕様

| opcode | 名前 | 送信方向 | 意味 |
|---|---|---|---|
| 0x9 | Ping | サーバー → クライアント | 「生きてる？」 |
| 0xA | Pong | クライアント → サーバー | 「生きてる」 |

RFC 6455 の規定:
- Ping を受け取ったクライアントはできるだけ早く Pong を返す
- Pong の payload には Ping と同じデータを入れる（省略可）
- サーバーはいつでも Ping を送れる

---

## ハートビートの設計

```
Server                        Client
  │                             │
  │──── Ping ──────────────────→│  (定期送信)
  │                             │
  │←─── Pong ──────────────────│  (応答)
  │                             │
  │  ←タイムアウト→             │  (Pong が来なかった)
  │──── Close ─────────────────→│  (切断)
```

タイムアウト管理のパターン:
1. `time.Ticker` で定期的に Ping を送る
2. `time.Timer` でタイムアウトを設定する
3. Pong を受け取ったら Timer をリセットする
4. Timer が発火したら（Pong が来なかった）接続を閉じる

---

## time.Timer と time.Ticker の違い

| | Timer | Ticker |
|---|---|---|
| 発火回数 | 1回 | 繰り返し |
| リセット | `Reset(d)` で可能 | なし（Stop して作り直す） |
| 用途 | タイムアウト検知 | 定期実行 |

ハートビートでの使い方:
```
pingInterval := 10 * time.Second   // Ping の送信間隔
pongTimeout  :=  5 * time.Second   // Pong 待ちのタイムアウト

ticker := time.NewTicker(pingInterval)
timer  := time.NewTimer(pingInterval + pongTimeout)  // 最初の Ping 送信後に Pong を待つ時間
```

---

## goroutine の構成

Ping/Pong を追加すると goroutine が増えます。

```
接続ごとに:
  goroutine 1（受信）: ReadFrame ループ + Pong 検知 → timer.Reset()
  goroutine 2（送信）: send チャネルからフレームを書く
  goroutine 3（Ping）: Ticker で定期 Ping → send チャネルに投入
  goroutine 4（タイムアウト）: timer.C を監視 → 発火で conn.Close()
```

終了シグナルの伝播:
- `done chan struct{}` で goroutine 3・4 に終了を通知する
- `close(done)` で全 goroutine が `select` の `<-done` を受信できる

---

## 📌 まとめ: Step 5 のフロー

1. 接続確立後に `Ticker` と `Timer` を作る
2. Ping goroutine を起動: Ticker のたびに Ping フレームを `send` チャネルへ
3. タイムアウト goroutine を起動: `timer.C` を待ち、発火したら `conn.Close()`
4. 受信ループで Pong (opcode=0xA) を検知したら `timer.Reset(...)`
5. 接続終了時に `ticker.Stop()` / `timer.Stop()` / `close(done)`
