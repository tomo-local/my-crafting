# 変数・型・定数

## 基本的な型

Go の基本型と、宣言だけしたときに入る**ゼロ値**。

| 型 | 例 | ゼロ値 |
|---|---|---|
| `int` | `42` | `0` |
| `float64` | `3.14` | `0` |
| `string` | `"hello"` | `""` |
| `bool` | `true` | `false` |

```go
var count int    // 0
var flag bool    // false
var label string // ""
```

`nil` になるのはポインタ・スライス・マップ・チャネル・インターフェースだけ。基本型は `nil` にならない。

---

## 変数宣言の3パターン

```go
// 1. var で型を明示
var name string = "Alice"

// 2. var で型推論（初期値から型が決まる）
var age = 30

// 3. 短縮宣言（関数の中だけ使える）
city := "Tokyo"
```

関数内では `:=` を使うのが Go らしいスタイル。  
パッケージレベルや型を明示したいときに `var` を使う。

---

## 定数

`const` で宣言した値は変更できない。型なし定数は使う場所で型が決まる。

```go
const MaxRetry = 3       // 型なし（int として使われる）
const Pi float64 = 3.14  // 型あり
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

各フィールドにもゼロ値が入るので、部分的な初期化も安全にできる。

```go
u := User{Name: "Bob"}  // Age は 0
```

### メソッド

struct に関数を紐づけられる。**値レシーバ**はコピーに対して、**ポインタレシーバ**は元の struct に対して操作する。

```go
// 値レシーバ：struct を変更しない処理
func (u User) Greet() string {
    return "Hello, " + u.Name
}

// ポインタレシーバ：フィールドを変更する処理
func (u *User) Birthday() {
    u.Age++
}

u := User{Name: "Alice", Age: 30}
fmt.Println(u.Greet())  // Hello, Alice
u.Birthday()
fmt.Println(u.Age)      // 31
```

同じ struct のメソッドはどちらかに統一するのが慣例。変更があるメソッドが1つでもあればすべてポインタレシーバに揃える。

---

## ポインタ（`&` と `*`）

Go は関数に値を**コピーして渡す**。大きな struct を毎回コピーするのは非効率で、呼び出し先から元の値を変更することもできない。ポインタはその問題を解決する。

### `&`（アドレス演算子）

変数のアドレスを取得する。結果はポインタ型になる。

```go
x := 42
p := &x   // p は *int 型（int へのポインタ）
```

### `*`（型宣言と間接参照）

`*` には2つの意味がある。文脈で読み分ける。

```go
// 型として使う：「〜へのポインタ」を表す
var p *int   // int へのポインタ型（ゼロ値は nil）

// 式として使う：ポインタが指す値を取り出す（間接参照）
x := 42
p := &x
fmt.Println(*p)  // 42
*p = 100         // x の値が 100 に変わる
fmt.Println(x)   // 100
```

### struct とポインタ

`&` で struct のポインタを取得できる。フィールドアクセスの `.` はポインタでも同じ記法で書ける（自動的に間接参照される）。

```go
p := &User{Name: "Bob", Age: 25}
fmt.Println(p.Name)  // Bob（(*p).Name と書かなくてよい）
p.Age = 26           // 元の struct が変わる
```

関数に struct を渡して変更したいときはポインタを渡す。

```go
func birthday(u *User) {
    u.Age++
}

u := User{Name: "Alice", Age: 30}
birthday(&u)
fmt.Println(u.Age)  // 31
```

### nil に注意

ポインタのゼロ値は `nil`。間接参照するとパニックになる。

```go
var p *int
fmt.Println(p)   // <nil>
fmt.Println(*p)  // パニック：nil pointer dereference
```

---

## まとめ

| トピック | ポイント |
|---|---|
| 基本型 | ゼロ値が自動で入る。`nil` にはならない |
| 変数宣言 | 関数内は `:=`、パッケージレベルは `var` |
| 定数 | `const` で宣言。型なしは使う場所で型が決まる |
| struct | 複数フィールドをまとめる複合型。クラスの代わり |
| メソッド | 変更あり→ポインタレシーバ、変更なし→値レシーバで統一 |
| `&` / `*` | `&` でアドレス取得、`*型` でポインタ型宣言、`*変数` で間接参照 |
