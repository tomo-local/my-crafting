# CRON Task Scheduler

Building a cron-like task scheduler from scratch to understand time-based job scheduling.

## Implementations

- [Go](./go)

## Goals

- Parse cron expressions (e.g. `*/5 * * * *`)
- Schedule and execute jobs at the correct time
- Handle missed jobs (on restart)
- Concurrent job execution with configurable worker pool
