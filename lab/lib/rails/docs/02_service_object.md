# Step 2: Service Object（3日）

1 つの責務を持つ Plain Ruby Object。ビジネスロジックの置き場所。

---

## 実装

```ruby
# app/services/entries/create_entry.rb
module Entries
  class CreateEntry
    Result = Data.define(:success, :entry, :errors)

    def initialize(user:, job:)
      @user = user
      @job  = job
    end

    def call
      return duplicate_error if already_applied?
      entry = Entry.create!(user: @user, job: @job)
      Result.new(success: true, entry: entry, errors: [])
    rescue ActiveRecord::RecordInvalid => e
      Result.new(success: false, entry: nil, errors: e.record.errors.full_messages)
    end

    private

    def already_applied?
      Entry.exists?(user: @user, job: @job)
    end

    def duplicate_error
      Result.new(success: false, entry: nil, errors: ['すでに応募済みです'])
    end
  end
end
```

```ruby
# app/services/scouts/send_scout.rb
module Scouts
  class SendScout
    Result = Data.define(:success, :errors)

    def initialize(user:, scout:)
      @user  = user
      @scout = scout
    end

    def call
      message = "#{@user.name}さん、スカウトが届いています: #{@scout.occupation}"
      Rails.logger.info "SMS to #{@user.phone_number}: #{message}"
      Result.new(success: true, errors: [])
    rescue => e
      Result.new(success: false, errors: [e.message])
    end
  end
end
```

---

## Controller から呼ぶ

```ruby
# app/controllers/entries_controller.rb
class EntriesController < ApplicationController
  def create
    result = Entries::CreateEntry.new(user: current_user, job: @job).call
    if result.success
      render json: result.entry, status: :created
    else
      render json: { errors: result.errors }, status: :unprocessable_entity
    end
  end
end
```

---

## テスト

```ruby
# spec/services/entries/create_entry_spec.rb
RSpec.describe Entries::CreateEntry do
  subject(:result) { described_class.new(user:, job:).call }

  let(:user) { create(:user) }
  let(:job)  { create(:job) }

  it { expect(result.success).to be true }
  it { expect(result.entry).to be_persisted }

  context '同じ求人に2回応募した場合' do
    before { described_class.new(user:, job:).call }
    it { expect(result.success).to be false }
    it { expect(result.errors).to include('すでに応募済みです') }
  end
end
```

---

## 確認ポイント

- [ ] `User` モデルからロジックが消えた
- [ ] `call` の結果が `Result` オブジェクトで一貫している
- [ ] Controller が「Service を呼ぶだけ」になった
- [ ] テストが軽量（外部依存なし）で書けた
