# ゼロから構築するWebSocketブローカー自作・全体学習プロセス

標準ライブラリの `golang.org/x/net/websocket` を使わず、RFC 6455 に従って HTTP Upgrade ハンドシェイクからフレームパースまで全て手実装することで、WebSocket の仕組みを深く理解していくプロセスです。

---

## 全体ロードマップ

| Step | テーマ | ゴール | 状態 |
|------|--------|--------|------|
| 1 | HTTP Upgrade ハンドシェイク | WebSocket 接続が確立し、`wscat` でサーバーに接続できる | |
| 2 | フレームの送受信 | クライアントからのメッセージをターミナルに表示できる | |
| 3 | エコーサーバー | 送ったメッセージがそのまま返ってくる | |
| 4 | Pub/Sub ルーティング | 複数クライアントがトピック経由でメッセージを交換できる | |
| 5 | Ping/Pong ハートビート | 無通信時に切断を検知してコネクションを管理できる | |

---

## 各ステップの詳細

### Step 1: HTTP Upgrade ハンドシェイク

WebSocket は普通の HTTP リクエストから始まります。クライアントが「WebSocket に切り替えたい」と申し出て、サーバーが同意することで接続が確立します。

**学習内容**
- HTTP/1.1 の `Upgrade` ヘッダーによるプロトコル切り替え
- `Sec-WebSocket-Key` と `Sec-WebSocket-Accept` の計算（SHA-1 + Base64）
- `101 Switching Protocols` レスポンスの構造
- ハンドシェイク後も TCP 接続をそのまま保持する理由

**実験ゴール**

```bash
# サーバー起動
go run main.go

# wscat で接続確認（npm install -g wscat）
wscat -c ws://localhost:8080
# → Connected (press CTRL+C to quit) が表示されれば成功
```

---

### Step 2: フレームの送受信

WebSocket の通信単位は「フレーム」です。HTTP のテキストと違い、バイナリフォーマットで制御情報とペイロードが混在しています。

**学習内容**
- RFC 6455 のフレームフォーマット（FIN/opcode/MASK/payload length）
- クライアント→サーバーは必ずマスクされる仕様とその理由
- payload length の可変長エンコーディング（7bit / 16bit / 64bit）
- opcode の種類（text, binary, close, ping, pong）

**実験ゴール**

```bash
wscat -c ws://localhost:8080
> hello
# サーバー側のターミナルに "Received: hello" が表示されること
```

---

### Step 3: エコーサーバー

受け取ったメッセージをそのまま送り返します。サーバーからクライアントへの送信フォーマットが受信時と異なる点（マスク不要）に注意が必要です。

**学習内容**
- サーバー→クライアント送信はマスクしない（RFC 規定）
- フレームの構築方法（ヘッダーバイトの組み立て）
- Close フレームの送受信と接続終了シーケンス

**実験ゴール**

```bash
wscat -c ws://localhost:8080
> hello
< hello   # 送ったメッセージがそのまま返ってくること
```

---

### Step 4: Pub/Sub ルーティング

複数クライアントが「トピック」を経由してメッセージを交換するブローカーを実装します。

**学習内容**
- ブローカーの役割（送信者と受信者を分離する中継役）
- `Hub` 構造体によるトピック→接続リストの管理
- goroutine per connection と共有状態の `sync.RWMutex` 保護
- メッセージプロトコルの設計（`SUBSCRIBE:topic` / `PUBLISH:topic:message`）

**実験ゴール**

```bash
# ターミナル A（購読者）
wscat -c ws://localhost:8080
> SUBSCRIBE:news

# ターミナル B（発行者）
wscat -c ws://localhost:8080
> PUBLISH:news:hello everyone

# ターミナル A に "hello everyone" が届くこと
```

---

### Step 5: Ping/Pong ハートビート

長時間無通信の接続を放置するとリソースが無駄に消費されます。RFC 6455 の Ping/Pong フレームでクライアントの生存確認を行います。

**学習内容**
- Ping/Pong フレームの仕様（opcode 0x9/0xA）
- `time.Ticker` による定期 Ping 送信
- Pong 未受信でのタイムアウト処理
- goroutine リークを防ぐキャンセル設計

**実験ゴール**

```bash
# wscat で接続後、しばらく放置する
wscat -c ws://localhost:8080
# 一定時間後にサーバー側で "client timed out" ログが出て切断されること
```

---

## 学習を進める上でのアドバイス

1. **`wscat` を使う**
   WebSocket のデバッグには `wscat`（npm package）が便利です。`npm install -g wscat` でインストールできます。

2. **フレームを `tcpdump` で覗く**
   `tcpdump -i lo -w capture.pcap port 8080` でキャプチャして生バイトを確認すると、フレームフォーマットの理解が深まります。

3. **RFC 6455 を手元に置く**
   フレームフォーマットの詳細は [RFC 6455 Section 5](https://www.rfc-editor.org/rfc/rfc6455#section-5) にあります。特に masking の仕様は実装時に必ず参照してください。

4. **goroutine リークに注意する**
   接続が切れた後も goroutine が生き続けるケースがあります。`runtime.NumGoroutine()` をログに出しながら開発すると確認しやすいです。

---

## 完走後の次のステップ

- **バイナリフレームのサポート**: 画像や音声データの中継
- **圧縮拡張（permessage-deflate）**: RFC 7692 に基づくペイロード圧縮
- **TLS（wss://）**: WebSocket over TLS で暗号化通信
- **スケールアウト**: 複数プロセスへのブロードキャスト（Redis Pub/Sub との連携）
