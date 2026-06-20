# Rails Learning Roadmap

## Sandbox App

A small job board app similar to plexjob-api, used to practice all steps.

```bash
rails new sandbox-job-board --api --database=postgresql
```

Models: `User` (job seeker) / `Job` (job posting) / `Entry` (application) / `Scout`

---

## Phase 0: Rails Basics

| Doc | Topics |
|---|---|
| [00_basics](./docs/00_basics.md) | Setup, MVC, routing, CRUD, ActiveRecord |

## Phase 1: Architecture Patterns

| Step | Doc | Duration |
|---|---|---|
| Step 1 | [01_fat_model](./docs/01_fat_model.md) | 1 day |
| Step 2 | [02_service_object](./docs/02_service_object.md) | 3 days |
| Step 3 | [03_query_object](./docs/03_query_object.md) | 3 days |
| Step 4 | [04_presenter](./docs/04_presenter.md) | 2 days |
| Step 5 | [05_strategy](./docs/05_strategy.md) | 3 days |
| Step 6 | [06_use_case](./docs/06_use_case.md) | 3 days |
| Step 7 | [07_integration](./docs/07_integration.md) | 3 days |
