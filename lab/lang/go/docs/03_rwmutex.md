# sync.RWMutex — 読み書きを分離するロック

## なぜ RWMutex が必要か

`sync.Mutex` は読み込みも書き込みも同じロックで保護する。  
しかし「複数の goroutine が同時に読む」のはデータを壊さない。排他する必要がない。

```
Mutex の場合:
  goroutine A: 読む → Lock() が必要 → 他の読み込みも全部待たされる

RWMutex の場合:
  goroutine A: 読む → RLock() → 他の読み込みは並走できる
  goroutine B: 読む → RLock() → A と同時に実行できる
  goroutine C: 書く → Lock()  → A と B が終わるまで待つ
```

| 操作 | 使うロック | 並走できる相手 |
|---|---|---|
| 読み込み | `RLock()` / `RUnlock()` | 他の読み込み（RLock）はOK |
| 書き込み | `Lock()` / `Unlock()` | 読み込みも書き込みも全部待つ |

---

## 基本的な使い方

```go
// 手を動かす：読み込み専用と書き込み専用を分離する
package main

import (
    "fmt"
    "sync"
    "time"
)

type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()         // 書き込み：排他ロック
    defer c.mu.Unlock()
    c.data[key] = value
}

func (c *Cache) Get(key string) string {
    c.mu.RLock()         // 読み込み：共有ロック
    defer c.mu.RUnlock()
    return c.data[key]
}

func main() {
    cache := &Cache{data: make(map[string]string)}
    cache.Set("name", "Alice")

    var wg sync.WaitGroup

    // 10 goroutine が同時に読む → RLock なので並走できる
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            v := cache.Get("name")
            fmt.Printf("goroutine %d: %s\n", n, v)
        }(i)
    }

    // 書き込みが来たら全読み込みが終わるまで待つ
    time.Sleep(1 * time.Millisecond)
    cache.Set("name", "Bob")

    wg.Wait()
}
```

```bash
go run -race main.go
# DATA RACE なし
```

---

## pub-sub の Broker に当てはめる

```go
func (b *Broker) Subscribe(topic string, sub *Subscriber) {
    b.mu.Lock()         // スライスに append する → 書き込み → 排他ロック
    defer b.mu.Unlock()
    b.subscribers[topic] = append(b.subscribers[topic], sub)
}

func (b *Broker) Publish(topic, message string) {
    b.mu.RLock()        // スライスを読むだけ → 読み込み → 共有ロック
    defer b.mu.RUnlock()
    for _, sub := range b.subscribers[topic] {
        sub.ch <- message
    }
}
```

`Publish` は同時に何本来ても `RLock` 同士は干渉しない。  
`Subscribe` が走っているときだけ `Publish` が待つ。

---

## よくある間違い

### RLock 中に書き込みしてしまう

```go
// NG: RLock しているのに map を変更している
func (b *Broker) BadPublish(topic string) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    b.subscribers[topic] = nil // DATA RACE！
}
```

読み込みロック中は絶対に書き込みしない。

### RUnlock を呼び忘れる

```go
// NG: defer を使わず早期 return で RUnlock を忘れる
func (c *Cache) Get(key string) string {
    c.mu.RLock()
    if _, ok := c.data[key]; !ok {
        return "" // RUnlock されずに抜けてしまう → デッドロック
    }
    v := c.data[key]
    c.mu.RUnlock()
    return v
}

// OK: defer で確実に解放
func (c *Cache) Get(key string) string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}
```

---

## まとめ

| | `sync.Mutex` | `sync.RWMutex` |
|---|---|---|
| 読み込み | `Lock()` / `Unlock()` | `RLock()` / `RUnlock()` |
| 書き込み | `Lock()` / `Unlock()` | `Lock()` / `Unlock()` |
| 読み込みの並走 | できない | できる |
| 使い所 | 読み書きの割合が同程度 | 読み込みが圧倒的に多い |

- 読み込みには `RLock()` / `RUnlock()` を使う
- 書き込みには `Lock()` / `Unlock()` を使う
- `defer RUnlock()` / `defer Unlock()` を `Lock` 直後に書く習慣をつける
