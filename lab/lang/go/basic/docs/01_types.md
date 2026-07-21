# 変数・型・定数

## 変数宣言の3パターン

Go の変数宣言には3つの書き方がある。

```go
// 1. var で明示的に型を指定
var name string = "Alice"

// 2. var で型推論（初期値から型を決める）
var age = 30

// 3. 短縮宣言（関数の中だけで使える）
city := "Tokyo"
```

実際には関数内では `:=` を使うことがほとんど。  
パッケージレベルの変数や型を明示したいときに `var` を使う。

---

## 基本的な型

| 型 | 例 | ゼロ値 |
|---|---|---|
| `int` | `42` | `0` |
| `float64` | `3.14` | `0` |
| `string` | `"hello"` | `""` |
| `bool` | `true` | `false` |

### ゼロ値

Go では初期化しなくても**ゼロ値**が自動で入る。`nil` になるケースは少ない。

```go
var count int    // 0
var flag bool    // false
var label string // ""
```

---

## 定数

`const` で宣言した値は変更できない。型なし定数は使う場所で型が決まる。

```go
const MaxRetry = 3       // 型なし（int として使われる）
const Pi float64 = 3.14  // 型あり
```

---

## ポインタ（`&` と `*`）

Go のポインタは「変数のメモリアドレスを保持する値」。

### `&`（アドレス演算子）

変数のアドレスを取得する。結果はポインタ型になる。

```go
x := 42
p := &x   // p は *int 型（int へのポインタ）
```

### `*`（型宣言と間接参照）

`*` には2つの使い方がある。

```go
// 型として使う：「〜へのポインタ」を表す
var p *int   // int へのポインタ型の変数 p（初期値は nil）

// 式として使う：ポインタが指す値を取り出す（間接参照・デリファレンス）
x := 42
p := &x
fmt.Println(*p)  // 42
*p = 100         // x の値が 100 に変わる
```

### なぜポインタが必要か

Go は関数に値を**コピーして渡す**。ポインタを渡せば、呼び出し先から元の変数を変更できる。

```go
func double(n *int) {
    *n = *n * 2
}

x := 5
double(&x)
fmt.Println(x)  // 10
```

### ゼロ値と nil

ポインタのゼロ値は `nil`（アドレスがない状態）。`nil` ポインタを間接参照するとパニックになる。

```go
var p *int
fmt.Println(p)   // <nil>
fmt.Println(*p)  // パニック：nil pointer dereference
```

---

## struct（構造体）

複数のフィールドをまとめた複合型。Go にはクラスがなく、struct がその代わりになる。

### 型の定義とインスタンス生成

```go
type User struct {
    Name string
    Age  int
}

// フィールド名を指定して初期化（推奨）
u := User{Name: "Alice", Age: 30}

// フィールドへのアクセス
fmt.Println(u.Name)  // Alice
u.Age = 31
```

### ゼロ値

初期化しなければ各フィールドにゼロ値が入る。

```go
var u User
fmt.Println(u.Name)  // ""
fmt.Println(u.Age)   // 0
```

### struct へのポインタ

`&` でポインタを取得できる。フィールドアクセスの `.` はポインタでも同じ（自動的に間接参照される）。

```go
p := &User{Name: "Bob", Age: 25}
fmt.Println(p.Name)  // Bob（(*p).Name と同じ）
```

### メソッド

struct に関数を紐づけられる。**値レシーバ**はコピーに対して、**ポインタレシーバ**は元の struct に対して操作する。

```go
// 値レシーバ：struct を変更しない処理に使う
func (u User) Greet() string {
    return "Hello, " + u.Name
}

// ポインタレシーバ：フィールドを変更する処理に使う
func (u *User) Birthday() {
    u.Age++
}

u := User{Name: "Alice", Age: 30}
fmt.Println(u.Greet())  // Hello, Alice
u.Birthday()
fmt.Println(u.Age)      // 31
```

> **原則**: フィールドを変更するならポインタレシーバ、変更しないなら値レシーバ。  
> ただし同じ struct のメソッドはどちらかに統一するのが慣例。

---

## まとめ

- 関数内では `:=` で書くのが Go らしいスタイル
- 宣言しただけでもゼロ値が入るので `nil` 参照は起きにくい
- `const` は変更されない値に使う
- `&` でアドレス取得、`*型` でポインタ型宣言、`*変数` で間接参照
- struct はフィールドをまとめる複合型。メソッドを生やせる
- メソッドは変更あり→ポインタレシーバ、変更なし→値レシーバで統一する
