# スライス・マップ

## スライス

Go の配列は固定長で使いにくいため、実際は**スライス**を使う。

```go
// リテラルで作る
nums := []int{1, 2, 3}

// make で作る（長さ0、容量5）
s := make([]int, 0, 5)
```

### append

スライスへの追加は `append`。元のスライスに追加するのではなく、新しいスライスを返す。

```go
nums = append(nums, 4, 5)
fmt.Println(nums) // [1 2 3 4 5]
```

### スライシング

`[low:high]` で部分スライスを作る。元のメモリを共有する（コピーではない）。

```go
fmt.Println(nums[1:3]) // [2 3]（インデックス1以上3未満）
fmt.Println(nums[:2])  // [1 2]
fmt.Println(nums[3:])  // [4 5]
```

### range

`range` でインデックスと値を同時に取得できる。

```go
for i, v := range nums {
    fmt.Printf("nums[%d] = %d\n", i, v)
}

// インデックスだけ
for i := range nums { ... }

// 値だけ
for _, v := range nums { ... }
```

---

## マップ

```go
// リテラルで作る
scores := map[string]int{
    "Alice": 90,
    "Bob":   75,
}

// make で作る
m := make(map[string]int)
```

### 追加・更新・削除

```go
scores["Carol"] = 85  // 追加
scores["Alice"] = 95  // 更新
delete(scores, "Bob") // 削除
```

### 存在確認

マップから値を取るとき、**2つ目の戻り値（ok）で存在確認**をする。  
キーがなくてもゼロ値が返るので、`ok` を確認しないと存在するのかゼロ値なのか区別できない。

```go
if v, ok := scores["Bob"]; ok {
    fmt.Println("Bob:", v)
} else {
    fmt.Println("Bob not found")
}
```

---

## まとめ

- 配列ではなくスライスを使う（`append` / `range` がセット）
- スライシングは元のメモリを共有する点に注意
- マップの値取得は必ず `v, ok := m[key]` で存在確認する
