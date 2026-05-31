# Meta — メタプログラミング・method_missing・define_method

## メタプログラミングとは

実行時にクラス・メソッド・変数を動的に操作する技術。Ruby はクラスが実行時に開いており、あらゆるオブジェクトを実行中に変更できる（オープンクラス）。

## オープンクラス

既存のクラスに後からメソッドを追加できる（モンキーパッチ）。

```ruby
class String
  def palindrome?
    self == self.reverse
  end
end

"racecar".palindrome?   # => true
"hello".palindrome?     # => false
```

> 注意: 標準クラスへのモンキーパッチは名前衝突のリスクがある。Refinements（後述）で安全にスコープを限定できる。

## respond_to? と send

```ruby
class Robot
  def greet
    "Hello!"
  end

  private

  def secret
    "hidden"
  end
end

r = Robot.new

r.respond_to?(:greet)    # => true
r.respond_to?(:secret)   # => false（private は false）
r.respond_to?(:secret, true)   # => true（第2引数 true で private も含む）

# send はプライベートメソッドも呼べる（テスト等での利用に限定すること）
r.send(:greet)    # => "Hello!"
r.send(:secret)   # => "hidden"

# public_send はパブリックのみ（推奨）
r.public_send(:greet)   # => "Hello!"
```

## method_missing

未定義メソッドが呼ばれたときに実行されるフック。DSL やプロキシオブジェクトで多用される。

```ruby
class DynamicProxy
  def initialize(target)
    @target = target
  end

  def method_missing(name, *args, &block)
    if @target.respond_to?(name)
      puts "[proxy] #{name} called"
      @target.send(name, *args, &block)
    else
      super   # 処理できない場合は必ず super を呼ぶ
    end
  end

  # method_missing を定義するときは respond_to_missing? もセットで定義する
  def respond_to_missing?(name, include_private = false)
    @target.respond_to?(name, include_private) || super
  end
end

proxy = DynamicProxy.new("hello")
proxy.upcase   # => [proxy] upcase called => "HELLO"
```

`method_missing` だけでは `respond_to?` が正しく動かない。**必ず `respond_to_missing?` を一緒に定義する**。

## define_method

実行時にメソッドをプログラム的に定義する。繰り返しパターンを除去するのに有効。

```ruby
class Status
  STATES = %w[pending processing completed failed].freeze

  STATES.each do |state|
    # "pending?" "processing?" ... を動的に定義
    define_method("#{state}?") do
      @state == state
    end
  end

  def initialize(state)
    @state = state
  end
end

s = Status.new("completed")
s.completed?   # => true
s.pending?     # => false
```

クロージャを使ってローカル変数をキャプチャできる点が `def` との違い。

```ruby
multipliers = {}
[2, 3, 5].each do |n|
  multipliers[n] = ->(x) { x * n }
end
# define_method も同様にクロージャをキャプチャする
```

## class_eval / module_eval

文字列またはブロックでクラスのコンテキストで評価する。

```ruby
class Person
  attr_reader :name

  def initialize(name)
    @name = name
  end
end

# ブロック形式（推奨）
Person.class_eval do
  def shout
    name.upcase + "!!!"
  end
end

Person.new("alice").shout   # => "ALICE!!!"

# 文字列形式（eval なのでセキュリティに注意）
Person.class_eval("def whisper; name.downcase; end")
```

## instance_variable_get / set

インスタンス変数に名前でアクセスする。

```ruby
class Config
  def initialize
    @host = "localhost"
    @port = 8080
  end
end

c = Config.new
c.instance_variable_get(:@host)         # => "localhost"
c.instance_variable_set(:@host, "prod") # ミュート
c.instance_variables                    # => [:@host, :@port]
```

## const_get / const_set

定数を動的に参照・定義する。ファクトリパターンなどで利用。

```ruby
class Dog; end
class Cat; end

def create_animal(type)
  Object.const_get(type).new
end

create_animal("Dog")   # => #<Dog:...>
```

## Refinements（安全なモンキーパッチ）

`using` したスコープ内だけに変更を限定できる。

```ruby
module StringExtensions
  refine String do
    def palindrome?
      self == self.reverse
    end
  end
end

# このファイル内でだけ有効
using StringExtensions
"racecar".palindrome?   # => true
```

## フック（コールバック）

クラスの変化に反応するフックメソッド群。

```ruby
module Hooks
  def self.included(base)
    puts "#{name} included into #{base}"
  end
end

class MyClass
  include Hooks   # => "Hooks included into MyClass"
end
```

主要フック:

| フック | タイミング |
|--------|-----------|
| `included(base)` | モジュールが include されたとき |
| `extended(base)` | モジュールが extend されたとき |
| `inherited(subclass)` | クラスが継承されたとき |
| `method_added(name)` | インスタンスメソッドが定義されたとき |
| `prepended(base)` | モジュールが prepend されたとき |

## まとめ: 使い分け指針

| 手法 | 適切な用途 |
|------|-----------|
| `define_method` | 繰り返しメソッドの動的生成 |
| `method_missing` | 未定義メソッドへの委譲・DSL |
| `class_eval` | クラスへの後付け定義 |
| `send` | メソッド名が動的に決まる呼び出し |
| Refinements | 安全なクラス拡張（スコープ限定） |
