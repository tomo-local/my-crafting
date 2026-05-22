# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

@.claude/rules/craft-common.md
@.claude/rules/craft-go.md
@.claude/rules/craft-ruby.md
@.claude/rules/lab.md

## Commands

```bash
# Go（crafts/<name>/go/ 配下）
go run ./crafts/<name>/go/main.go
go build ./...
go fmt ./...
go vet ./...

# Ruby（crafts/<name>/ruby/ 配下）
mise exec -- ruby main.rb

# Submodule の初期化・更新
git submodule update --init --recursive
git submodule update --remote <path>
```

## 構成

| ディレクトリ | 役割 |
|---|---|
| `crafts/` | ゼロから作るシステム実装 |
| `lab/lang/` | 言語学習リポジトリ（git submodule）|
| `lab/lib/` | ライブラリ/ランタイム学習（git submodule）|
| `lab/infra/` | インフラ演習（Kubernetes break-fix）|

## スキル

| コマンド | 役割 |
|---|---|
| `/new-craft` | 新しいクラフトをスキャフォールド |
| `/craft-learning-plan` | `docs/` に学習ドキュメント群を生成 |
| `/review-and-issue` | コード分析してGitHub Issueを作成 |

## PR ラベリング

`.github/labeler.yml` がファイルパスに基づいてPRへ自動ラベルを付与する。新しいクラフトを追加する際は同ファイルへのエントリ追加も必要。
