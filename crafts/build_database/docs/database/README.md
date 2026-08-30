# データベース自作チュートリアル 学習ノート

- 元サイト: https://trialofcode.org/database/#table-of-contents
- README.md の TODO に対応する、step単位の日本語学習ドキュメント

## 目次

### 01. Log-based KV
- [0101: KV interfaces](01-log-based-kv/0101-kv-interfaces.md)
- [0102: Serialization](01-log-based-kv/0102-serialization.md)
- [0103: Log storage](01-log-based-kv/0103-log-storage.md)
- [0104: fsync](01-log-based-kv/0104-fsync.md)
- [0105: Checksum](01-log-based-kv/0105-checksum.md)

### 02. Tables
- [0201: Datatypes](02-tables/0201-datatypes.md)
- [0202: Schemas](02-tables/0202-schemas.md)
- [0203: Update modes](02-tables/0203-update-modes.md)
- [0204: CRUD](02-tables/0204-crud.md)

### 03. Simple SQL
- [0301: Tokenizer](03-simple-sql/0301-tokenizer.md)
- [0302: Values](03-simple-sql/0302-values.md)
- [0303: SELECT](03-simple-sql/0303-select.md)
- [0304: Statements](03-simple-sql/0304-statements.md)
- [0305: Execute SQL](03-simple-sql/0305-execute-sql.md)

### 04. Range query
- [0401: Sort & search](04-range-query/0401-sort-search.md)
- [0402: Iterators](04-range-query/0402-iterators.md)
- [0403: Sort order](04-range-query/0403-sort-order.md)
- [0404: Row iterator](04-range-query/0404-row-iterator.md)
- [0405: Range query](04-range-query/0405-range-query.md)

### 05. More SQL
- [0501: Infix ops](05-more-sql/0501-infix-ops.md)
- [0502: Expr eval](05-more-sql/0502-expr-eval.md)
- [0503: Precedence](05-more-sql/0503-precedence.md)
- [0504: Expr parser](05-more-sql/0504-expr-parser.md)
- [0505: Select expr](05-more-sql/0505-select-expr.md)
- [0506: WHERE](05-more-sql/0506-where.md)
- [0507: SQL range](05-more-sql/0507-sql-range.md)

### 06. Log + data
- [0600: Atomic update](06-log-plus-data/0600-atomic-update.md)
- [0601: Build SSTable](06-log-plus-data/0601-build-sstable.md)
- [0602: Query SSTable](06-log-plus-data/0602-query-sstable.md)
- [0603: Refactor code](06-log-plus-data/0603-refactor-code.md)
- [0604: Merge sort](06-log-plus-data/0604-merge-sort.md)
- [0605: Log + SSTable](06-log-plus-data/0605-log-plus-sstable.md)

### 07. LSM-Tree
- [0700: LSM-Tree intro](07-lsm-tree/0700-lsm-tree-intro.md)
- [0701: Atomic store](07-lsm-tree/0701-atomic-store.md)
- [0702: Store metadata](07-lsm-tree/0702-store-metadata.md)
- [0703: Multi-levels](07-lsm-tree/0703-multi-levels.md)
- [0704: Merge levels](07-lsm-tree/0704-merge-levels.md)

### 08. Index
- [08: Index（フルブック参照）](08-index.md)

### 09. Concurrency
- [09: Concurrency（フルブック参照）](09-concurrency.md)
