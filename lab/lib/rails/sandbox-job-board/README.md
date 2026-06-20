# sandbox-job-board

A minimal job board API app for practicing Rails architecture patterns.

## Stack

- Ruby 3.4
- Rails 8.1 (API mode)
- PostgreSQL

## Models

- `Company` — company posting jobs
- `User` — job seeker
- `Job` — job posting (belongs to Company)
- `Entry` — application (User → Job)
- `Scout` — scout message

## Setup

```bash
mise install
rails db:create
rails db:migrate
rails server
```

## Learning Steps

See [docs/](../docs/) for the full learning roadmap.
