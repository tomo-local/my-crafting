# Go 並行処理 学習ロードマップ

## 目的

`sync.RWMutex`（RLock / Unlock）を中心に、Go の並行処理の仕組みを手を動かして理解する。  
pub-sub の `Broker` 実装で出てきた「なぜ Subscribe は Lock で Publish は RLock なのか」を自分の言葉で説明できるようになることがゴール。

## 学習順序

| # | ドキュメント | トピック |
|---|---|---|
| 1 | [goroutine と data race](./01_goroutine.md) | goroutine の基本、競合状態とは何か |
| 2 | [sync.Mutex](./02_mutex.md) | Lock / Unlock で競合を防ぐ |
| 3 | [sync.RWMutex](./03_rwmutex.md) | 読み書きを分離するロック |

## 進め方

1. ドキュメントを読んでコンセプトを把握する
2. `// 手を動かす` のコードスニペットを `.go` ファイルに写して実行する
3. `-race` フラグで data race を検出しながら確認する

```bash
go run -race main.go
```
