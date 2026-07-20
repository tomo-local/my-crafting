# sync.Mutex — Lock / Unlock

## Mutex とは

Mutual Exclusion（相互排除）の略。**一度に1つの goroutine しか入れない部屋** のイメージ。

```
goroutine A: Lock() → 部屋に入る
goroutine B: Lock() → 部屋が空くまで待つ（ブロック）
goroutine A: Unlock() → 部屋を出る
goroutine B: 入れた
```

---

## 基本的な使い方

```go
// 手を動かす：前のステップの counter を Mutex で保護する
package main

import (
    "fmt"
    "sync"
)

func main() {
    var mu sync.Mutex
    counter := 0
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mu.Lock()
            counter++ // この行は一度に1 goroutine しか実行できない
            mu.Unlock()
        }()
    }

    wg.Wait()
    fmt.Println(counter) // 必ず 1000
}
```

```bash
go run -race main.go
# DATA RACE が出なくなる
```

---

## defer Unlock のパターン

`Lock()` したら必ず `Unlock()` しなければいけない。途中で `return` や `panic` が起きても解放されるよう `defer` を使うのが定番。

```go
mu.Lock()
defer mu.Unlock()
// ここの処理が何があっても Unlock は呼ばれる
```

---

## Mutex の制限

Mutex はシンプルだが **読み込みにも排他が発生する** という欠点がある。

```
goroutine A: Lock() → カウンターを読む
goroutine B: Lock() → 待つ（読むだけなのに！）
goroutine C: Lock() → 待つ（読むだけなのに！）
```

複数の goroutine が「読むだけ」なら同時に実行しても競合は起きない。  
これを解決するのが次の **RWMutex**。

---

## まとめ

- `sync.Mutex` は `Lock()` / `Unlock()` で使う
- `defer mu.Unlock()` を `Lock()` の直後に書く
- 読み込みも書き込みも同じロックなので、読み込みが多い場面では非効率
