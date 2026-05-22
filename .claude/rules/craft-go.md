# Go Craft ルール

- **標準ライブラリのみ** — `go.mod` に `require` ブロックは追加しない
- **module名**: `github.com/tomo-local/<craft-name>`
- **goバージョン**: 既存クラフトに合わせる（現在 `go 1.25.1`）

## 内部構造

```
go/
├── go.mod
├── main.go       # ハンドラ配線のみ（ロジックは internal/ へ）
└── internal/
    ├── server/   # TCPリスナー・接続ライフサイクル・goroutineディスパッチ
    ├── request/  # バイト列 → 型付き構造体のパース
    └── response/ # レスポンス生成・書き込み
```

- 新しいサブシステムは対応する `internal/` パッケージを追加する
- 並行処理は goroutine ベース（接続1本につき1 goroutine）
- channel・sync.Mutex が必要でも外部ライブラリは使わない

## 実行

```bash
go run ./crafts/<name>/go/main.go
# 動作確認
curl -v http://localhost:<port>/
nc localhost <port>
```
