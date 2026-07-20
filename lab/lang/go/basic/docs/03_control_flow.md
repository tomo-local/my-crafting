# 制御フロー

## if

Go の `if` は条件式の前に**初期化文**を書ける。変数のスコープが `if` ブロックに閉じる。

```go
if x := compute(); x > 0 {
    fmt.Println("positive:", x)
} else {
    fmt.Println("non-positive:", x)
}
// x はここでは使えない
```

---

## for

Go にはループ構文が `for` しかない。書き方を変えることで while 相当もできる。

```go
// C スタイル
for i := 0; i < 3; i++ {
    fmt.Println(i)
}

// while 相当
n := 1
for n < 100 {
    n *= 2
}

// 無限ループ
for {
    // break で抜ける
}
```

---

## switch

Go の `switch` は条件に一致した case で自動的に抜ける（`break` 不要）。  
複数の値を1つの `case` に並べられる。

```go
switch day {
case "Saturday", "Sunday":
    fmt.Println("weekend")
case "Monday":
    fmt.Println("start of the week")
default:
    fmt.Println("weekday")
}
```

型スイッチ（インターフェース学習時に登場する）：

```go
switch v := i.(type) {
case int:
    fmt.Println("int:", v)
case string:
    fmt.Println("string:", v)
}
```

---

## defer

`defer` をつけた関数呼び出しは、**囲んでいる関数が終了するときに実行される**。  
複数の `defer` は LIFO（後入れ先出し）で実行される。

```go
func readFile() {
    f, _ := os.Open("file.txt")
    defer f.Close() // 関数が終わったら必ず閉じる

    // ファイルを読む処理
}
```

リソースの後始末（ファイルクローズ・ロック解除）に使うのが典型。

---

## まとめ

- ループは `for` だけ。`while` / `do-while` はない
- `switch` は `break` 不要
- `defer` はリソース後始末のイディオム。`sync.Mutex` の `defer mu.Unlock()` が典型的な例
