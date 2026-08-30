# 🗄️ Code a database in 45 steps (Go)

Go でデータベースをゼロから実装していくチュートリアル。

- チュートリアル: https://trialofcode.org/database/#table-of-contents
- 作業ログ（Zenn Scrap）: https://zenn.dev/tomo_local/scraps/e97c4888b194fc
- 学習メモ: [docs/memo.md](docs/memo.md)

## TODO

### 01. Log-based KV
- [x] 0101: KV interfaces
- [x] 0102: Serialization
- [x] 0103: Log storage
- [x] 0104: fsync
- [x] 0105: Checksum

### 02. Tables
- [x] 0201: Datatypes
- [x] 0202: Schemas
- [x] 0203: Update modes
- [x] 0204: CRUD

### 03. Simple SQL
- [ ] 0301: Tokenizer
- [ ] 0302: Values
- [ ] 0303: SELECT
- [ ] 0304: Statements
- [ ] 0305: Execute SQL

### 04. Range query
- [ ] 0401: Sort & search
- [ ] 0402: Iterators
- [ ] 0403: Sort order
- [ ] 0404: Row iterator
- [ ] 0405: Range query

### 05. More SQL
- [ ] 0501: Infix ops
- [ ] 0502: Expr eval
- [ ] 0503: Precedence
- [ ] 0504: Expr parser
- [ ] 0505: Select expr
- [ ] 0506: WHERE
- [ ] 0507: SQL range

### 06. Log + data
- [ ] 0600: Atomic update
- [ ] 0601: Build SSTable
- [ ] 0602: Query SSTable
- [ ] 0603: Refactor code
- [ ] 0604: Merge sort
- [ ] 0605: Log + SSTable

### 07. LSM-Tree
- [ ] 0700: LSM-Tree intro
- [ ] 0701: Atomic store
- [ ] 0702: Store metadata
- [ ] 0703: Multi-levels
- [ ] 0704: Merge levels

### 08. Index
- [ ] See the full book

### 09. Concurrency
- [ ] See the full book
