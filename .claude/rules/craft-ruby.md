# Ruby Craft ルール

- mise でバージョン管理。`ruby/` ディレクトリに `mise.toml` を置く
- 外部 gem は原則追加しない（標準ライブラリのみ）

## ファイル構成

```
ruby/
├── mise.toml   # [tools] ruby = "3.4"
└── main.rb
```

## 実行

```bash
# ruby/ ディレクトリ内から
mise exec -- ruby main.rb
```
