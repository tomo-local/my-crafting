package main

import (
	"fmt"
	"sync"
	"time"
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
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()
	fmt.Println(counter)
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
