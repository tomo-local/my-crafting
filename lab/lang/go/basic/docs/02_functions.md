# 関数

## 基本の形

```go
func add(a, b int) int {
    return a + b
}
```

同じ型の引数は `a, b int` とまとめて書ける。

---

## 複数の戻り値

Go の関数は値を複数返せる。エラーハンドリングに使うのが典型的な用途。

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}
```

呼び出し側：

```go
result, err := divide(10, 3)
if err != nil {
    fmt.Println("error:", err)
    return
}
fmt.Printf("%.2f\n", result)
```

### エラーハンドリングの慣用句

1. 戻り値の最後に `error` を置く
2. 呼び出し直後に `if err != nil` でチェックする
3. エラーがあればその場で `return` する

この「エラーが来たらすぐ返す」スタイルが Go らしい書き方。

---

## _ でアンダースコア

使わない戻り値は `_` で捨てる。ただし `error` を `_` で捨てるのは本番では避ける。

```go
result, _ := divide(10, 3) // エラーを無視（学習・プロトタイプ限定）
```

---

## まとめ

- 複数の戻り値はタプルではなく、Go の言語機能として組み込まれている
- エラーは戻り値で返すのが Go の流儀（例外ではない）
- `if err != nil` を書き忘れないことが重要
