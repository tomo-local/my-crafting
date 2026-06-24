# Step 2：フレームの送受信（前提知識）

---

## このステップで何が変わるか

Step 1 でハンドシェイクが終わった後、TCP 接続に流れるデータは HTTP ではなく WebSocket フレームになります。このステップではフレームのバイナリフォーマットを理解し、クライアントからのメッセージをパースしてターミナルに表示できるようにします。

---

## フレームヘッダーの読み方

最初の 2 バイトから以下の情報を取り出します。

```
バイト 1: [FIN(1bit)] [RSV1(1bit)] [RSV2(1bit)] [RSV3(1bit)] [opcode(4bit)]
バイト 2: [MASK(1bit)] [payload_len(7bit)]
```

**バイト 1 の取り出し方**:
```
byte1  := data[0]
fin    := (byte1 & 0x80) != 0   // 最上位ビット
opcode := byte1 & 0x0F           // 下位 4 ビット
```

**バイト 2 の取り出し方**:
```
byte2      := data[1]
masked     := (byte2 & 0x80) != 0  // 最上位ビット
payloadLen := int(byte2 & 0x7F)    // 下位 7 ビット
```

---

## payload length の 3 段階

`payloadLen`（7bit の値）によって、後続の読み取りが変わります。

| 値 | 意味 | 後続の読み取り |
|---|---|---|
| 0〜125 | そのまま payload 長 | なし |
| 126 | 次の 2 バイトが実際の長さ | `uint16` を big-endian で読む |
| 127 | 次の 8 バイトが実際の長さ | `uint64` を big-endian で読む |

```
payloadLen = 7bit の値
if payloadLen == 126:
    実際の長さ = 次の 2 バイト (uint16, big-endian)
elif payloadLen == 127:
    実際の長さ = 次の 8 バイト (uint64, big-endian)
else:
    実際の長さ = payloadLen そのまま
```

---

## マスキングの解除

クライアントから届くフレームは必ずマスクされています（`MASK=1`）。

**マスキングキーの取得**: `MASK=1` の場合、payload の直前に 4 バイトのキーがあります。

**アンマスク処理**:
```
for i, b := range maskedPayload {
    original[i] = b XOR maskingKey[i % 4]
}
```

XOR は可逆演算なので、同じ処理でマスクもアンマスクも行えます。

---

## 全体の読み取りシーケンス

```
1. 2 バイト読む（FIN, opcode, MASK, payload_len）
2. payload_len に応じて 0/2/8 バイト追加読み取り（実際の長さ）
3. MASK=1 なら 4 バイト読む（masking key）
4. 実際の長さ分のバイトを読む（payload data）
5. MASK=1 なら XOR でアンマスク
```

---

## 📌 まとめ: Step 2 のフロー

1. `io.ReadFull` で 2 バイト読み、FIN/opcode/MASK/payload_len を取り出す
2. payload_len が 126/127 なら追加バイトを読んで実際の長さを確定する
3. MASK が 1 なら masking key を 4 バイト読む
4. payload data を実際の長さ分読む
5. MASK が 1 なら XOR でアンマスクする
6. opcode と payload を返す
