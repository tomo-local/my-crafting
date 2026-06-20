# Step 1: 問題の理解 — Fat Model（1日）

Fat Model が何を引き起こすかを手で体験する。

---

## やること

1. `User` モデルにエントリー作成・スカウト送信・メール通知を全部書く
2. テストを書いてみる（書きにくさを体感する）
3. 新機能（LINE 通知を追加）を実装する（変更の難しさを体感する）

```ruby
# app/models/user.rb — 意図的に太らせる
class User < ApplicationRecord
  has_many :entries
  has_many :jobs, through: :entries

  def apply_to_job(job)
    return false if entries.exists?(job: job)
    entry = entries.create!(job: job)
    send_application_email(entry)
    notify_slack(entry)
    entry
  end

  def send_scout_notification(scout)
    # SMS 送信ロジック
    phone = phone_number
    message = "#{name}さん、スカウトが届いています: #{scout.occupation}"
    # SmsClient.send(phone, message) # 外部APIが直接ここに
    Rails.logger.info "SMS sent to #{phone}: #{message}"
  end

  private

  def send_application_email(entry)
    # メール送信ロジックが直接ここに
    Rails.logger.info "Email sent to #{email} for job #{entry.job.title}"
  end

  def notify_slack(entry)
    Rails.logger.info "Slack notified for #{name}"
  end
end
```

---

## 体感するべき問題

### テストが書きにくい

```ruby
# User のロジック1つをテストするために DB + Rails スタック全体が必要
user = User.create!(name: 'Taro', email: 'taro@example.com')
job  = Job.create!(title: 'ドライバー')
user.apply_to_job(job)
# SMS / Slack のモックも必要になる
```

### 新機能追加で既存コードが壊れる恐怖

「LINE 通知も追加して」→ `apply_to_job` を直接編集する → 既存 SMS/メールが壊れないか不安

---

## 確認ポイント

- [ ] モデルが 100 行を超えて、どこに何があるか分からなくなった
- [ ] テストに DB と Rails スタック全体が必要なことを確認した
- [ ] 新機能追加でモデルを壊す恐怖を感じた
