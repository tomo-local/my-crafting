# Step 1：最小Pub/Sub（前提知識）

Pub/Subの本質は「トピックという名前の郵便受けを介した間接通信」です。Step 1 ではサブスクライバー1人・トピック1つという最小構成で、プロトコルのパース・チャネルによる非同期配信・TCPへの書き戻しを実装します。

---

## 1. プロトコル概要

1行1コマンドのテキストプロトコルです。行末は `\r\n`（CRLF）。

```
クライアント→ブローカー:
  SUB news\r\n        ← "news" トピックを購読する
  PUB news Hello\r\n  ← "news" トピックにメッセージを発行する

ブローカー→クライアント:
  +OK\r\n             ← コマンド受理
  MSG news Hello\r\n  ← メッセージ配信
```

---

## 2. ブローカーのデータ構造

Step 1 では1トピック固定でよいですが、Step 3 で拡張することを見越して最初からマップで管理します。

```
Broker
  └── subscribers: map[string][]*Subscriber
        └── "news" → []*Subscriber{ subA }
```

各 `Subscriber` はTCP接続と受信チャネルを持ちます。

```
Subscriber
  ├── conn  net.Conn      ← クライアントとのTCP接続
  └── ch    chan string   ← ブローカーからのメッセージ受け取り口
```

---

## 3. 接続処理の分離

1つのTCP接続で「コマンドの読み取り」と「メッセージの書き込み」が同時に必要です。

```
goroutine A（readLoop）: コマンドを読んで SUB / PUB を処理する
goroutine B（writeLoop）: sub.ch からメッセージを受け取って conn に書く
```

この2つを同時に走らせることで、PUB コマンドの処理中に MSG を送ることができます。

---

## 4. PUB コマンドの処理フロー

```
PUB news Hello が届く
    ↓
topic = "news", message = "Hello" を取り出す
    ↓
broker.subscribers["news"] のリストを取得
    ↓
for _, sub := range subscribers {
    sub.ch <- message   ← チャネルに入れるだけ（ブロックしない設計は Step 4 で）
}
```

---

## 📌 まとめ：Step 1 のフロー

**起動時:**
1. `net.Listen` でブローカーを起動する
2. `Accept` ループで接続を受け付ける
3. 接続ごとに `go handleClient(conn)` を起動する

**`handleClient` の中:**
1. `Subscriber{conn, make(chan string, 64)}` を作る
2. `go writeLoop(sub)` を起動する（チャネルを読んでTCPに書く）
3. `readLoop(sub)` でコマンドを読み続ける（SUB, PUB を処理する）

**`writeLoop` の中:**
1. `for msg := range sub.ch` でメッセージを待つ
2. `fmt.Fprintf(conn, "MSG <topic> %s\r\n", msg)` で書き出す
