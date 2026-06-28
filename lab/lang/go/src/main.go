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
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *Cache) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.data[key]
	if !ok {
		return ""
	}
	return value
}

func main() {
	cache := &Cache{data: make(map[string]string)}
	cache.Set("name", "Alice")

	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			v := cache.Get("name")
			fmt.Println(v)
			fmt.Printf("goroutine %d: %s\n", n, v)
		}(i)
	}

	time.Sleep(1 * time.Millisecond)
	cache.Set("name", "Bob")

	wg.Wait()
	fmt.Println("done")
}

// ## goroutine
func outputTiming() {
	go fmt.Println("another goroutine")
	// goroutineの起動が少しかかるので、Sleepを追加して待つ
	time.Sleep(10 * time.Microsecond)
	fmt.Println("main")
	time.Sleep(10 * time.Millisecond)
}

// ## data race
func waitGroup() {
	counter := 0
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		// １つ待つための処理を追加する
		wg.Add(1)
		go func() {
			// 処理が完了したことをpushしている
			defer wg.Done()
			counter++

		}()
	}

	wg.Wait()
	fmt.Println(counter)
}

func waitGroupWithLock() {
	var mu sync.Mutex
	counter := 0
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()
	fmt.Println(counter)
}
