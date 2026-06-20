# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema[8.1].define(version: 2026_06_20_081329) do
  # These are extensions that must be enabled in order to support this database
  enable_extension "pg_catalog.plpgsql"

  create_table "companies", force: :cascade do |t|
    t.string "building"
    t.string "city"
    t.string "country"
    t.datetime "created_at", null: false
    t.string "industry"
    t.string "name"
    t.string "postal_code"
    t.string "prefecture"
    t.string "street"
    t.datetime "updated_at", null: false
  end

  create_table "entries", force: :cascade do |t|
    t.datetime "created_at", null: false
    t.bigint "job_id", null: false
    t.datetime "updated_at", null: false
    t.bigint "user_id", null: false
    t.index ["job_id"], name: "index_entries_on_job_id"
    t.index ["user_id"], name: "index_entries_on_user_id"
  end

  create_table "jobs", force: :cascade do |t|
    t.bigint "company_id", null: false
    t.datetime "created_at", null: false
    t.text "description"
    t.string "occupation"
    t.string "prefecture"
    t.datetime "published_at"
    t.integer "salary"
    t.string "title"
    t.datetime "updated_at", null: false
    t.index ["company_id"], name: "index_jobs_on_company_id"
  end

  create_table "scouts", force: :cascade do |t|
    t.bigint "company_id", null: false
    t.datetime "created_at", null: false
    t.string "occupation"
    t.string "prefecture"
    t.integer "required_experience"
    t.datetime "updated_at", null: false
    t.index ["company_id"], name: "index_scouts_on_company_id"
  end

  create_table "users", force: :cascade do |t|
    t.string "building"
    t.string "city"
    t.string "country"
    t.datetime "created_at", null: false
    t.integer "desired_salary"
    t.string "email"
    t.string "name"
    t.string "occupation"
    t.string "phone_number"
    t.string "postal_code"
    t.string "prefecture"
    t.string "street"
    t.datetime "updated_at", null: false
  end

  add_foreign_key "entries", "jobs"
  add_foreign_key "entries", "users"
  add_foreign_key "jobs", "companies"
  add_foreign_key "scouts", "companies"
end
