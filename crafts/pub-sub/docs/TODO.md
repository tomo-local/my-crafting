# pub-sub Go 実装 復習 TODO

## サマリー

- **理解度**: Step 1〜4 実装経験あり、概念はうっすら残っているが詳細が曖昧
- **時間**: 今日1〜2時間
- **ゴール**: Step 5（ワイルドカード）の実装に進める ＋ 他人に説明できる
- **重点**: goroutine とチャネルの絡み方（readLoop / writeLoop の役割分担）

---

## Phase 1: 全体構造の再把握（20分）

- [ ] `main.go` → `handleClient` → `readLoop` / `writeLoop` の流れをコードで追う
- [ ] `Broker` / `Subscriber` / `Subscription` の役割と関係を頭の中で図にする
- [ ] SUB / PUB / UNSUB コマンドそれぞれが呼び出す関数を確認する

## Phase 2: goroutine とチャネルを重点復習（30分）

- [ ] `SUB news` したとき何が起きるかを口頭で説明できる
  - `broker.Subscribe` → `ch` 取得 → `writeLoop` goroutine 起動
- [ ] `PUB news Hello` したとき何が起きるかを口頭で説明できる
  - `broker.Publish` → 各 Subscriber の `ch` に send → `writeLoop` が受け取り `conn` に書き込み
- [ ] クライアント切断時の流れを追う
  - `scanner.Scan()` が false → `readLoop` 終了 → `UnsubscribeAll` 呼び出し

## Phase 3: RWMutex とバックプレッシャーの確認（20分）

- [ ] `Publish` が `RLock`、`Subscribe` / `Unsubscribe` が `Lock` な理由を説明できる
- [ ] バッファ64 + `select default` でドロップする仕組みを一言で説明できる
- [ ] `dropped` カウントのログが100件ごとな理由を考える

## Phase 4: Step 5 への準備（10分）

- [ ] ワイルドカード（`news.*`）対応で `broker.go` の何が変わるか考える
- [ ] `Publish` のトピックルックアップが完全一致から変わることを確認する

---

## チェックポイント（自問自答）

1. `SUB news` したとき、goroutine はいくつ増えるか？どこで起動されるか？
2. `PUB news Hello` したとき、メッセージはどの経路でサブスクライバーに届くか？
3. クライアントが突然切断したとき、`readLoop` / `writeLoop` / チャネルはどうなるか？
4. `Publish` が `RLock`、`Subscribe` が `Lock` な理由は？
5. バッファ64の `select default` でドロップする仕組みを一言で説明すると？
