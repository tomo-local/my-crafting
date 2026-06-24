# WebSocketブローカー 基礎知識

---

## 📡 WebSocket とは

HTTP はクライアントがリクエストを送り、サーバーがレスポンスを返したら通信が完了する「半二重」プロトコルです。WebSocket は同じ TCP 接続を使いながら、両端が自由なタイミングでメッセージを送れる「全二重」プロトコルです。

```
HTTP:
  Client → Server: GET /
  Server → Client: 200 OK + body
  [接続を閉じる or 次のリクエストを待つ]

WebSocket:
  Client → Server: upgrade request
  Server → Client: 101 Switching Protocols
  [同じ TCP 接続のまま...]
  Client → Server: "hello"
  Server → Client: "world"
  Client → Server: "ping"
  ...（どちらからでも自由に送れる）
```

---

## 🤝 HTTP Upgrade メカニズム

WebSocket 接続は通常の HTTP/1.1 リクエストから始まります。

### クライアントのリクエスト

```
GET / HTTP/1.1
Host: localhost:8080
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13
```

| ヘッダー | 意味 |
|---|---|
| `Upgrade: websocket` | プロトコル切り替えの要求 |
| `Connection: Upgrade` | この接続を Upgrade に使うことを明示 |
| `Sec-WebSocket-Key` | ランダムな 16 バイトを Base64 エンコードした値 |
| `Sec-WebSocket-Version: 13` | RFC 6455 のバージョン番号 |

### サーバーのレスポンス

```
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

`Sec-WebSocket-Accept` の計算:

```
1. Sec-WebSocket-Key を取得
   "dGhlIHNhbXBsZSBub25jZQ=="

2. 固定 GUID を末尾に連結
   "dGhlIHNhbXBsZSBub25jZQ==258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

3. SHA-1 ハッシュ計算
   → b37a4f2cc0624f1690f64606cf385945b2bec4ea

4. Base64 エンコード
   → "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
```

---

## 🔲 WebSocket フレームフォーマット

ハンドシェイク後のデータは「フレーム」という単位でやり取りされます。

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-------+-+-------------+-------------------------------+
|F|R|R|R| opcode|M| Payload len |    Extended payload length    |
|I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
|N|V|V|V|       |S|             |   (if payload len==126/127)   |
| |1|2|3|       |K|             |                               |
+-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - -+
|     Extended payload length continued, if payload len == 127  |
+ - - - - - - - - - - - - - - -+-------------------------------+
|                               |  Masking-key, if MASK set to 1|
+-------------------------------+-------------------------------+
| Masking-key (continued)       |          Payload Data         |
+-------------------------------- - - - - - - - - - - - - - - -+
:                     Payload Data continued ...                :
+---------------------------------------------------------------+
```

### 各フィールドの意味

| フィールド | ビット数 | 説明 |
|---|---|---|
| FIN | 1 | このフレームがメッセージの最終フレームなら 1 |
| RSV1-3 | 各1 | 拡張用（通常は 0） |
| opcode | 4 | フレームの種類 |
| MASK | 1 | ペイロードがマスクされているなら 1 |
| Payload len | 7 | 0〜125: そのまま / 126: 次の2バイト / 127: 次の8バイト |
| Masking-key | 32 | マスクキー（MASK=1 の場合のみ存在） |
| Payload Data | 可変 | 実際のデータ |

### opcode の種類

| 値 | 名前 | 説明 |
|---|---|---|
| 0x0 | Continuation | 前フレームの継続 |
| 0x1 | Text | UTF-8 テキスト |
| 0x2 | Binary | バイナリデータ |
| 0x8 | Close | 接続終了 |
| 0x9 | Ping | 生存確認 |
| 0xA | Pong | Ping への応答 |

---

## 🔑 マスキング

クライアント → サーバー方向のフレームは**必ずマスクされます**（RFC 6455 の強制仕様）。サーバー → クライアント方向はマスクしてはいけません。

マスクの適用方法:
```
masked_byte[i] = original_byte[i] XOR masking_key[i % 4]
```

アンマスクも同じ XOR 演算で行えます（XOR は可逆）。

**なぜマスクが必要か**: 中間プロキシのキャッシュポイズニング攻撃を防ぐためです（RFC 6455 Section 10.3）。

---

## 🔁 Pub/Sub パターン

ブローカーは送信者（Publisher）と受信者（Subscriber）を切り離す中継役です。

```
Publisher ──PUBLISH:news:hello──→ [ Broker / Hub ]
                                         │
                               ┌─────────┼─────────┐
                               ↓         ↓         ↓
                          Subscriber1 Sub2      Sub3
                          (news)     (news)    (sports)
                             ← hello  ← hello    ✗
```

Hub の内部構造:
```
topics: map[string][]*Client
{
  "news":   [client1, client2],
  "sports": [client3],
}
```

---

## ⚠️ 実装上の罠・注意点

1. **マスキングキーは 4 バイト固定**
   `MASK=1` の場合、payload の前に必ず 4 バイトのキーが来ます。読み飛ばさないよう注意。

2. **payload length は 3 パターン**
   `0〜125: そのまま` / `126: 次の 2 バイト（uint16）` / `127: 次の 8 バイト（uint64）` — 分岐を正しく実装しないとパース失敗。

3. **サーバーはマスクしない**
   送信時に `MASK bit = 0` にする。クライアント（ブラウザ等）はサーバーからマスクされたフレームを受け取ると切断します。

4. **Close フレームは双方向**
   一方が Close フレームを送ったら、もう一方も Close フレームを返してから TCP を閉じるのが RFC 6455 の規定です。

5. **goroutine の管理**
   接続ごとに goroutine を立てる設計では、接続終了時に Hub からの登録解除と goroutine の終了を確実に行わないとリークします。
