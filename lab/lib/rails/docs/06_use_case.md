# Step 6: Use Case（3日）

1 つのユーザー操作の全手順をまとめるクラス。Service / Query / Strategy を組み合わせてオーケストレーションする。UseCase 自身はビジネスロジックを持たない。

---

## 実装

```ruby
# app/use_cases/scouts/send_scout.rb
module Scouts
  class SendScout
    Result = Data.define(:success, :errors)

    def initialize(scout:, admin:)
      @scout = scout
      @admin = admin
    end

    def call
      targets = find_targets
      save_histories(targets)         # トランザクション内
      notify_users(targets)           # トランザクション外（外部 API）
      Result.new(success: true, errors: [])
    rescue => e
      Result.new(success: false, errors: [e.message])
    end

    private

    def find_targets
      Scouts::TargetSearchQuery.new(
        occupation: @scout.occupation,
        prefecture: @scout.prefecture
      ).call
    end

    def save_histories(targets)
      ActiveRecord::Base.transaction do
        targets.each { |user| ScoutHistory.create!(scout: @scout, user:) }
      end
    end

    def notify_users(targets)
      targets.each do |user|
        Scouts::SendScout.new(user:, scout: @scout).call
      end
    end
  end
end
```

---

## トランザクション分離の原則

```
トランザクション内  → DB 更新（失敗したらロールバック）
トランザクション外  → 外部 API（SMS・メール・Slack）
```

外部 API をトランザクション内に入れると、DB ロールバック後も SMS が送信済みになるバグが起きる。

---

## Controller から呼ぶ

```ruby
# app/controllers/scouts_controller.rb
def send_scout
  result = Scouts::SendScout.new(scout: @scout, admin: current_admin).call
  if result.success
    render json: { message: 'スカウト送信完了' }
  else
    render json: { errors: result.errors }, status: :unprocessable_entity
  end
end
```

Controller は UseCase を呼ぶだけ。20 行以内に収まる。

---

## 確認ポイント

- [ ] UseCase が Service と Query を組み合わせているだけで、自分ではビジネスロジックを持っていない
- [ ] 外部 API 呼び出しがトランザクションの外にある
- [ ] Controller が 20 行以内に収まった
