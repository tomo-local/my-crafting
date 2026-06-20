# Step 3: Query Object（3日）

複雑な DB 検索をクラスに封じ込める。`scope` では表現しにくい複数条件の検索に使う。

---

## scope との使い分け

| | scope | Query Object |
|---|---|---|
| 条件が | 単純（1〜2条件） | 複雑（複数テーブル・サブクエリ） |
| 再利用 | モデル内で完結 | 複数箇所から呼ぶ |
| テスト | モデルテストで十分 | 独立してテストしたい |

---

## 実装

```ruby
# app/queries/jobs/recommended_jobs_query.rb
module Jobs
  class RecommendedJobsQuery
    def initialize(user:, relation: Job.all)
      @user     = user
      @relation = relation
    end

    def call
      @relation
        .where(occupation: @user.occupation)
        .where('salary >= ?', @user.desired_salary.to_i)
        .order(created_at: :desc)
        .limit(20)
    end
  end
end
```

```ruby
# app/queries/jobs/not_applied_jobs_query.rb
module Jobs
  class NotAppliedJobsQuery
    def initialize(user:, relation: Job.all)
      @user     = user
      @relation = relation
    end

    def call
      applied_job_ids = Entry.where(user: @user)
                             .where('created_at >= ?', 30.days.ago)
                             .select(:job_id)
      @relation.where.not(id: applied_job_ids)
    end
  end
end
```

---

## 組み合わせて使う

```ruby
# Controller や UseCase から
base = Job.published
base = Jobs::RecommendedJobsQuery.new(user: current_user, relation: base).call
base = Jobs::NotAppliedJobsQuery.new(user: current_user, relation: base).call
render json: base
```

`relation:` を外から渡すことでチェーンが可能になる。

---

## スカウト対象検索（ERB SQL の置き換え）

```ruby
# app/queries/scouts/target_search_query.rb
module Scouts
  class TargetSearchQuery
    def initialize(occupation:, prefecture:)
      @occupation = occupation
      @prefecture = prefecture
    end

    def call
      User
        .where(occupation: @occupation)
        .where.not(id: recently_scouted_ids)
    end

    private

    def recently_scouted_ids
      Scout.where(created_at: 30.days.ago..).select(:user_id)
    end
  end
end
```

---

## 確認ポイント

- [ ] `scope` と Query Object の使い分けが判断できる
- [ ] `relation:` を外から渡して Query Object を合成できた
- [ ] Query Object にパラメータを渡して単体テストできた
