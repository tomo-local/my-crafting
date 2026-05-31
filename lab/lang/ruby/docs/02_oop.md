# OOP — クラス・モジュール・ミックスイン・継承

## クラス

```ruby
class User
  # attr_accessor でゲッター/セッターを自動生成
  # attr_reader → 読み取りのみ、attr_writer → 書き込みのみ
  attr_accessor :name, :age
  attr_reader   :id

  # クラス変数（インスタンス間で共有）
  @@count = 0

  def initialize(id, name, age)
    @id   = id      # インスタンス変数
    @name = name
    @age  = age
    @@count += 1
  end

  # クラスメソッド（self. プレフィックス）
  def self.count
    @@count
  end

  # インスタンスメソッド
  def adult?
    @age >= 18
  end

  def to_s
    "User(#{@id}, #{@name})"
  end
end

alice = User.new(1, "Alice", 30)
alice.name         # => "Alice"
alice.age = 31     # セッター
alice.adult?       # => true
User.count         # => 1
```

## 継承

```ruby
class Animal
  attr_reader :name

  def initialize(name)
    @name = name
  end

  def speak
    raise NotImplementedError, "#{self.class} must implement speak"
  end

  def to_s
    "#{self.class}(#{@name})"
  end
end

class Dog < Animal
  def speak
    "Woof!"
  end
end

class Cat < Animal
  def speak
    "Meow!"
  end
end

# super でスーパークラスのメソッドを呼ぶ
class Puppy < Dog
  def speak
    "#{super} (tiny)"   # => "Woof! (tiny)"
  end
end
```

## アクセス制御

```ruby
class BankAccount
  def initialize(balance)
    @balance = balance
  end

  # public（デフォルト）
  def deposit(amount)
    @balance += validate(amount)
  end

  # protected: サブクラスや同クラスのインスタンス間で呼べる
  protected

  def balance
    @balance
  end

  # private: クラス内部からのみ呼べる（レシーバを指定できない）
  private

  def validate(amount)
    raise ArgumentError if amount <= 0
    amount
  end
end
```

## モジュール

モジュールには2つの用途がある。

### 名前空間

```ruby
module Payments
  class Invoice
    # ...
  end

  class Receipt
    # ...
  end
end

Payments::Invoice.new
```

### ミックスイン（Mixin）

Ruby は単一継承だが、モジュールを `include` することで複数の振る舞いを合成できる。

```ruby
module Greetable
  def greet
    "Hello, I'm #{name}"   # name はインクルード先で定義されていると想定
  end
end

module Farewell
  def bye
    "Goodbye from #{name}"
  end
end

class Person
  include Greetable
  include Farewell

  attr_reader :name

  def initialize(name)
    @name = name
  end
end

Person.new("Alice").greet   # => "Hello, I'm Alice"
Person.new("Alice").bye     # => "Goodbye from Alice"
```

### extend vs include vs prepend

| 方式 | 効果 |
|------|------|
| `include` | モジュールのメソッドをインスタンスメソッドとして追加（継承チェーンの後方に挿入） |
| `extend` | モジュールのメソッドをクラスメソッドとして追加（または特定オブジェクトに追加） |
| `prepend` | 継承チェーンの前方に挿入（元メソッドより先に呼ばれる） |

```ruby
module Logging
  def greet
    puts "[LOG] calling greet"
    super
  end
end

class Person
  prepend Logging

  def greet
    "Hello!"
  end
end

Person.new.greet
# => [LOG] calling greet
# => "Hello!"
```

## Comparable と Enumerable

標準モジュールを mixin することで演算子やイテレータを自動実装できる。

```ruby
class Temperature
  include Comparable

  attr_reader :degrees

  def initialize(degrees)
    @degrees = degrees
  end

  # <=> を定義するだけで <, <=, >, >=, between?, clamp が使える
  def <=>(other)
    degrees <=> other.degrees
  end
end

temps = [Temperature.new(30), Temperature.new(20), Temperature.new(25)]
temps.min.degrees   # => 20
temps.sort.map(&:degrees)   # => [20, 25, 30]
```

## Method Resolution Order (MRO)

Ruby は C3 線形化でメソッドを探索する。`ancestors` で確認できる。

```ruby
module A; end
module B; end
class C
  include A
  include B
end

C.ancestors   # => [C, B, A, Object, Kernel, BasicObject]
# 後から include したものが先に探索される
```

## Struct

軽量なデータクラスを素早く定義できる。

```ruby
Point = Struct.new(:x, :y) do
  def distance_to(other)
    Math.sqrt((x - other.x)**2 + (y - other.y)**2)
  end
end

p = Point.new(0, 0)
q = Point.new(3, 4)
p.distance_to(q)   # => 5.0
p == Point.new(0, 0)   # => true（値比較が自動実装される）
```
