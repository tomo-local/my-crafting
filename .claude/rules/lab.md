# Lab ルール

## Submodule の扱い

`lab/lang/` と `lab/lib/` の各ディレクトリは独立した git リポジトリを submodule として参照している。
**このリポジトリからファイルを直接編集しない** — 各 submodule のリポジトリで編集する。

```bash
# 特定の submodule を最新に更新
git submodule update --remote lab/lang/ruby

# 全 submodule を初期化
git submodule update --init --recursive
```

## Ruby (lab/lang/ruby)

mise で Ruby バージョンを管理（`mise.toml`: `ruby = "3.4"`）。

```bash
mise exec -- ruby <file.rb>
```

## Kubernetes (lab/infra/kubernetes)

`build-break-fix/` は意図的に壊れた設定を修正する演習。README の指示に従って進める。
