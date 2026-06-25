package main

import (
	"fmt"
	"time"
)

func main() {
	go fmt.Println("another goroutine")
	// goroutineの起動が少しかかるので、Sleepを追加して待つ
	time.Sleep(10 * time.Microsecond)
	fmt.Println("main")
	time.Sleep(10 * time.Millisecond)
}
