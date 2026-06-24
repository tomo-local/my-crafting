# goroutine と data race

## goroutine とは

`go` キーワードで関数を別スレッドのように並行実行できる。

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    go fmt.Println("別の goroutine から")
    fmt.Println("main から")
    time.Sleep(10 * time.Millisecond) // goroutine の完了を待つ（確認用）
}
```

実行するたびに出力順序が変わる。どちらが先に実行されるかは保証されない。

---

## data race とは

複数の goroutine が同じ変数を「同時に」読み書きすると値が壊れる。これを **data race（データ競合）** と呼ぶ。

```go
// 手を動かす：これを -race フラグ付きで実行する
package main

import (
    "fmt"
    "sync"
)

func main() {
    counter := 0
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++ // 複数 goroutine が同時にここを実行する → 競合
        }()
    }

    wg.Wait()
    fmt.Println(counter) // 1000 にならないことがある
}
```

```bash
go run -race main.go
# DATA RACE が検出される
```

> **`sync.WaitGroup` とは**  
> `wg.Add(n)` で「n 個の goroutine を待つ」、`wg.Done()` で「1つ完了」を通知する。  
> `wg.Wait()` は全員が `Done` を呼ぶまでブロックする。goroutine の完了を確実に待ちたいときに使う。

---

## まとめ

- goroutine は軽量スレッドで `go func()` で起動する
- 複数 goroutine が同じ変数を同時に読み書きすると data race が発生する
- `-race` フラグで検出できる
- 競合を防ぐには **ロック（次のステップ）** が必要
