# Basics — 型・変数・制御フロー・メソッド・ブロック

## 型

Ruby はすべてがオブジェクト。リテラルから直接メソッドを呼べる。

```ruby
# 数値
42.class        # => Integer
3.14.class      # => Float
42.even?        # => true
-5.abs          # => 5

# 文字列
"hello".upcase  # => "HELLO"
"hello".length  # => 5
"hello #{42}"   # => "hello 42"  （式展開はダブルクォートのみ）

# シンボル（軽量な不変文字列。ハッシュのキーに多用）
:name.class     # => Symbol
:name.to_s      # => "name"

# 配列
[1, 2, 3].first   # => 1
[1, 2, 3].last    # => 3
[1, 2, 3].length  # => 3

# ハッシュ
{ name: "Alice", age: 30 }   # シンボルキー（推奨）
{ "name" => "Alice" }        # 文字列キー

# nil・真偽値（nil と false だけが偽。0 や "" は真）
nil.nil?    # => true
false.nil?  # => false
```

## 変数

```ruby
# ローカル変数（スネークケース）
user_name = "Alice"

# 定数（大文字始まり）
MAX_RETRY = 3

# グローバル変数（$ プレフィックス。実用では避ける）
$global = "avoid this"

# インスタンス変数・クラス変数は OOP セクションで扱う
```

## 制御フロー

### 条件分岐

```ruby
# if / elsif / else
if score >= 90
  "A"
elsif score >= 70
  "B"
else
  "C"
end

# 後置 if（条件が単純なとき）
puts "ok" if valid?

# unless（if not の糖衣構文）
unless error?
  proceed
end

# 三項演算子
label = admin? ? "admin" : "user"

# case / when（型・範囲・正規表現も使える）
case score
when 90..100 then "A"
when 70..89  then "B"
else              "C"
end
```

### ループ

```ruby
# times
3.times { |i| puts i }

# upto / downto
1.upto(5) { |i| puts i }

# each（配列・ハッシュの反復）
[1, 2, 3].each { |n| puts n }
{ a: 1, b: 2 }.each { |k, v| puts "#{k}: #{v}" }

# while
while condition
  # ...
end

# loop + break（明示的な無限ループ）
loop do
  input = gets.chomp
  break if input == "quit"
end

# next（continue 相当）、break（途中脱出）
[1, 2, 3, 4].each do |n|
  next if n.even?
  puts n   # => 1, 3
end
```

## メソッド

```ruby
# 基本定義
def greet(name)
  "Hello, #{name}!"   # 最後の式が戻り値（return 省略可）
end

# デフォルト引数
def greet(name = "World")
  "Hello, #{name}!"
end

# キーワード引数（呼び出し側が引数名を明示できる）
def connect(host:, port: 80)
  "#{host}:#{port}"
end
connect(host: "localhost")

# 可変長引数
def sum(*numbers)
  numbers.sum
end

# ? と ! の慣習
# ? → 真偽値を返す（破壊的でない）
# ! → 破壊的操作 or 失敗時に例外を投げる
"hello".empty?        # => false
[3, 1, 2].sort        # 新しい配列を返す（元は不変）
[3, 1, 2].sort!       # 元の配列を破壊的にソート
```

## ブロック

ブロックはメソッドに渡す処理のかたまり。`{ }` か `do...end` で書く。

```ruby
# { } は一行、do...end は複数行が慣習
[1, 2, 3].map { |n| n * 2 }       # => [2, 4, 6]

[1, 2, 3].each do |n|
  puts n * 2
end

# yield でブロックを呼び出す
def repeat(n)
  n.times { yield }
end
repeat(3) { puts "hi" }

# block_given? でブロックの有無を確認
def maybe_yield
  yield if block_given?
end

# & でブロックを Proc として受け取る
def capture(&block)
  block.call(42)
end
capture { |n| puts n }   # => 42

# Proc と lambda
double = Proc.new { |n| n * 2 }
double.call(5)   # => 10

triple = ->(n) { n * 3 }   # lambda（引数チェックが厳密）
triple.call(5)   # => 15

# Symbol#to_proc（& で簡潔に書ける）
["hello", "world"].map(&:upcase)   # => ["HELLO", "WORLD"]
```

## よく使う Enumerable メソッド

```ruby
nums = [1, 2, 3, 4, 5]

nums.map    { |n| n * 2 }       # => [2, 4, 6, 8, 10]  変換
nums.select { |n| n.odd? }      # => [1, 3, 5]          絞り込み
nums.reject { |n| n.odd? }      # => [2, 4]             除外
nums.find   { |n| n > 3 }       # => 4                  最初の一件
nums.reduce(0) { |sum, n| sum + n }  # => 15            畳み込み
nums.any?   { |n| n > 4 }       # => true
nums.all?   { |n| n > 0 }       # => true
nums.none?  { |n| n > 10 }      # => true
nums.count  { |n| n.even? }     # => 2
nums.sort_by { |n| -n }         # => [5, 4, 3, 2, 1]
```
