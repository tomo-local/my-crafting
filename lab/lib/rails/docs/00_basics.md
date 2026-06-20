# Phase 0: Rails 基礎

通常の Rails 開発を一通り体験する。アーキテクチャ学習の前提知識を固める。

---

## セットアップ

```bash
rails new sandbox-job-board --api --database=postgresql
cd sandbox-job-board
docker compose up -d
rails db:create
```

---

## 1. モデルとマイグレーション

### Value Object: Address

`User` と `Company` で共有する住所の値オブジェクト。DB カラムは各テーブルに持ち、`composed_of` で Ruby オブジェクトとして扱う。

```ruby
# app/models/address.rb
class Address
  module Country
    JAPAN = "Japan"
    ALL   = [JAPAN].freeze
  end

  attr_reader :country, :postal_code, :prefecture, :city, :street, :building

  def initialize(postal_code:, prefecture:, city:, street:, building: nil, country: Country::JAPAN)
    raise ArgumentError, "Invalid country: #{country}" unless Country::ALL.include?(country)
    @country     = country
    @postal_code = postal_code
    @prefecture  = prefecture
    @city        = city
    @street      = street
    @building    = building
  end
end
```

### migration

```bash
# Company
rails g model Company name:string industry:string \
  country:string postal_code:string prefecture:string city:string street:string building:string

# User
rails g model User name:string email:string occupation:string desired_salary:integer phone_number:string \
  country:string postal_code:string prefecture:string city:string street:string building:string

# Job（Company に属する）
rails g model Job company:references title:string description:text \
  salary:integer occupation:string prefecture:string published_at:datetime

# Entry（重複応募防止に unique index）
rails g model Entry user:references job:references

# Scout
rails g model Scout company:references occupation:string prefecture:string required_experience:integer

rails db:migrate
```

Entry の unique index は migration に手動追加する。

```ruby
# db/migrate/xxx_create_entries.rb
add_index :entries, [:user_id, :job_id], unique: true
```

### Association

```ruby
# app/models/address.rb — 上記参照

# app/models/company.rb
class Company < ApplicationRecord
  composed_of :address,
    class_name: 'Address',
    mapping: { country: :country, postal_code: :postal_code, prefecture: :prefecture,
               city: :city, street: :street, building: :building }

  has_many :jobs
  validates :name, presence: true
end

# app/models/user.rb
class User < ApplicationRecord
  composed_of :address,
    class_name: 'Address',
    mapping: { country: :country, postal_code: :postal_code, prefecture: :prefecture,
               city: :city, street: :street, building: :building }

  has_many :entries
  has_many :jobs, through: :entries
  validates :email, presence: true, uniqueness: true
end

# app/models/job.rb
class Job < ApplicationRecord
  belongs_to :company
  has_many :entries
  has_many :users, through: :entries
  validates :title, presence: true
  scope :published, -> { where.not(published_at: nil) }
end

# app/models/entry.rb
class Entry < ApplicationRecord
  belongs_to :user
  belongs_to :job
end

# app/models/scout.rb
class Scout < ApplicationRecord
  belongs_to :company
end
```

---

## 2. ルーティング

```ruby
# config/routes.rb
Rails.application.routes.draw do
  resources :companies, only: [:index, :show, :create]
  resources :users, only: [:index, :show, :create]
  resources :jobs, only: [:index, :show, :create] do
    resources :entries, only: [:create]
  end
end
```

```bash
rails routes
```

---

## 3. コントローラ（CRUD）

```ruby
# app/controllers/jobs_controller.rb
class JobsController < ApplicationController
  def index
    @jobs = Job.published.order(created_at: :desc)
    render json: @jobs
  end

  def show
    @job = Job.find(params[:id])
    render json: @job
  end

  def create
    @job = Job.new(job_params)
    if @job.save
      render json: @job, status: :created
    else
      render json: { errors: @job.errors.full_messages }, status: :unprocessable_entity
    end
  end

  private

  def job_params
    params.require(:job).permit(:company_id, :title, :description, :salary, :occupation, :prefecture)
  end
end
```

---

## 4. ActiveRecord の主要操作

```ruby
# 検索
User.find(1)
User.find_by(email: 'test@example.com')
User.where(occupation: 'driver').order(created_at: :desc).limit(10)

# 作成
company = Company.create!(name: '株式会社サンプル', industry: '物流')
job = company.jobs.create!(title: 'ドライバー募集', salary: 3_500_000)

# Value Object へのアクセス
company.address.prefecture   #=> "東京都"
company.address.country      #=> "Japan"

# N+1 対策
Entry.includes(:user, :job).all
Job.includes(:company).published
```

---

## 5. 動作確認

```bash
rails server

curl -X POST http://localhost:3000/companies \
  -H 'Content-Type: application/json' \
  -d '{"company": {"name": "株式会社サンプル", "industry": "物流"}}'

curl -X POST http://localhost:3000/jobs \
  -H 'Content-Type: application/json' \
  -d '{"job": {"company_id": 1, "title": "ドライバー募集", "salary": 3500000}}'

curl http://localhost:3000/jobs
```

---

## 確認ポイント

- [ ] `rails new` から `rails server` まで動いた
- [ ] マイグレーションで 5 モデルを作れた
- [ ] `Company` → `Job` → `Entry` ← `User` の関連が機能した
- [ ] `composed_of` で `company.address.prefecture` にアクセスできた
- [ ] CRUD エンドポイントが `curl` で動いた
- [ ] N+1 が `includes` で解消できた
