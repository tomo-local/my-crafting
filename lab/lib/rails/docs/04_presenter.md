# Step 4: Presenter / Decorator（2日）

表示向けのロジックをモデルから分離する。モデルは「データ」、Presenter は「見せ方」。

---

## 実装

```ruby
# app/presenters/job_presenter.rb
class JobPresenter
  def initialize(job)
    @job = job
  end

  def salary_text
    return '応相談' if @job.salary.nil?
    "#{format('%d', @job.salary)}円〜"
  end

  def published_at_text
    @job.published_at&.strftime('%Y年%m月%d日') || '未公開'
  end

  def status_label
    @job.published_at ? '公開中' : '下書き'
  end
end
```

```ruby
# app/presenters/user_presenter.rb
class UserPresenter
  def initialize(user)
    @user = user
  end

  def desired_salary_text
    return '未設定' if @user.desired_salary.nil?
    "#{format('%d', @user.desired_salary)}万円以上"
  end
end
```

---

## Controller から使う

```ruby
# app/controllers/jobs_controller.rb
def show
  job = Job.find(params[:id])
  render json: {
    title: job.title,
    salary: JobPresenter.new(job).salary_text,
    published_at: JobPresenter.new(job).published_at_text
  }
end
```

---

## モデルに書いてはいけない例

```ruby
# Bad: モデルに表示ロジックが混入する
class Job < ApplicationRecord
  def salary_text  # ← ここに書くと Fat Model に逆戻り
    return '応相談' if salary.nil?
    "#{salary.to_s(:delimited)}円〜"
  end
end
```

---

## 確認ポイント

- [ ] モデルに `strftime` や文字列結合が一切なくなった
- [ ] Presenter のテストが DB なしで書けた
- [ ] 用途別に `JobPresenter` / `UserPresenter` に分けられた
