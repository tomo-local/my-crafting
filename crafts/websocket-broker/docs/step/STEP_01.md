# Step 1：HTTP Upgrade ハンドシェイク（前提知識）

---

## このステップで何を作るか

HTTP リクエストを受け取り、WebSocket プロトコルへの切り替えを完了するハンドシェイク処理を実装します。このステップが完了すると、`wscat` などの WebSocket クライアントがサーバーに接続できるようになります。

---

## HTTP Upgrade の流れ

通常の HTTP/1.1 接続として始まり、特定のヘッダーが揃っていれば WebSocket に切り替えます。

```
Client                          Server
  │                               │
  │── GET / HTTP/1.1 ────────────→│
  │   Upgrade: websocket          │
  │   Sec-WebSocket-Key: xxx      │
  │                               │ ← ハンドシェイク処理
  │←── HTTP/1.1 101 ─────────────│
  │    Upgrade: websocket         │
  │    Sec-WebSocket-Accept: yyy  │
  │                               │
  │  [WebSocket フレームの通信へ]  │
```

---

## Sec-WebSocket-Accept の計算

クライアントが送ってきた `Sec-WebSocket-Key` をそのまま返しても受け入れられません。RFC 6455 で定められた変換を行います。

```
入力: "dGhlIHNhbXBsZSBub25jZQ=="  (Sec-WebSocket-Key)

1. 固定 GUID を末尾に連結
   "dGhlIHNhbXBsZSBub25jZQ==258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

2. SHA-1 ハッシュ計算
   → b37a4f2cc0624f1690f64606cf385945b2bec4ea (hex)

3. Base64 エンコード
   → "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="  (Sec-WebSocket-Accept)
```

Go では `crypto/sha1` と `encoding/base64` で計算できます。

---

## 101 レスポンスの構造

```
HTTP/1.1 101 Switching Protocols\r\n
Upgrade: websocket\r\n
Connection: Upgrade\r\n
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=\r\n
\r\n
```

101 を返した後、この TCP 接続はもう HTTP としては使いません。同じ `net.Conn` をそのまま WebSocket フレームの読み書きに使い続けます。

---

## アップグレードの判定条件

全てのリクエストを WebSocket として扱うのではなく、以下のヘッダーが揃っているものだけ Upgrade します。

| ヘッダー | 期待値 |
|---|---|
| `Upgrade` | `websocket`（大文字小文字無視） |
| `Connection` | `Upgrade` を含む |
| `Sec-WebSocket-Version` | `13` |
| `Sec-WebSocket-Key` | 存在する |

---

## 📌 まとめ: Step 1 のフロー

1. TCP でリクエストを受け取り、HTTP ヘッダーをパースする
2. `Upgrade: websocket` ヘッダーを確認する
3. `Sec-WebSocket-Key` を取り出す
4. `key + GUID` を SHA-1 してから Base64 エンコードして `Sec-WebSocket-Accept` を生成する
5. `101 Switching Protocols` レスポンスを送る
6. 同じ `net.Conn` を返してフレーム処理へ渡す
