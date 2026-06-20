# Step 5: Strategy パターン（3日）

「やり方が違うが目的が同じ」処理を差し替え可能なクラス群で表現する。
職種別スカウトマッチングのコピペを 1 つに統合するのが目標。

---

## コピペ状態を再現する（問題の確認）

```ruby
# Bad: 職種ごとにほぼ同じコードが増殖する
def match_driver_targets(scout)
  User.where(occupation: :driver).where('experience_years >= ?', scout.required_experience)
end

def match_factory_targets(scout)
  User.where(occupation: :factory).where('experience_years >= ?', scout.required_experience)
end

def match_nurse_targets(scout)
  User.where(occupation: :nurse).where('license_required <= ?', scout.license_level)
end
```

---

## Strategy パターンで統合

```ruby
# app/services/scouts/matching_strategies/base_strategy.rb
module Scouts
  module MatchingStrategies
    class BaseStrategy
      def build_targets(scout)
        raise NotImplementedError, "#{self.class}#build_targets を実装してください"
      end
    end
  end
end
```

```ruby
# app/services/scouts/matching_strategies/driver_strategy.rb
module Scouts
  module MatchingStrategies
    class DriverStrategy < BaseStrategy
      def build_targets(scout)
        User.where(occupation: :driver)
            .where('experience_years >= ?', scout.required_experience)
      end
    end
  end
end
```

```ruby
# app/services/scouts/matching_strategies/factory_strategy.rb
module Scouts
  module MatchingStrategies
    class FactoryStrategy < BaseStrategy
      def build_targets(scout)
        User.where(occupation: :factory)
            .where('experience_years >= ?', scout.required_experience)
      end
    end
  end
end
```

---

## 動的な切り替え

```ruby
# app/services/scouts/matching_strategies.rb
module Scouts
  module MatchingStrategies
    STRATEGIES = {
      'driver'  => DriverStrategy,
      'factory' => FactoryStrategy,
      'nurse'   => NurseStrategy
    }.freeze

    def self.for(occupation_type)
      STRATEGIES.fetch(occupation_type) { raise ArgumentError, "未対応の職種: #{occupation_type}" }
        .new
    end
  end
end

# 呼び出し側
strategy = Scouts::MatchingStrategies.for(scout.occupation)
targets  = strategy.build_targets(scout)
```

---

## 新しい職種を追加するとき

`STRATEGIES` に 1 行追加 + 新クラス 1 ファイルだけ。既存コードは一切変更しない。

---

## 確認ポイント

- [ ] 職種を追加するときに `if/elsif` を増やさなくて済んだ
- [ ] `BaseStrategy` の `build_targets` を実装しないと `NotImplementedError` になる
- [ ] `Scouts::MatchingStrategies.for(type)` で動的に切り替えられた
