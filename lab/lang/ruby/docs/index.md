# Ruby 学習ロードマップ

## 概要

Ruby をゼロから学ぶためのドキュメント群。各ディレクトリの `.keep` ファイルをコードファイルに置き換えながら進める。

## 学習順序

| # | ドキュメント | ディレクトリ | トピック |
|---|---|---|---|
| 1 | [basics](./01_basics.md) | `basics/` | 型・変数・制御フロー・メソッド・ブロック |
| 2 | [oop](./02_oop.md) | `oop/` | クラス・モジュール・ミックスイン・継承 |
| 3 | [meta](./03_meta.md) | `meta/` | メタプログラミング・method_missing・define_method |

## 実行環境

```bash
mise install        # Ruby 3.4 をインストール
mise exec -- ruby <file.rb>
```

## 進め方

1. 各ドキュメントを読んでコンセプトを把握する
2. 対応ディレクトリに `.rb` ファイルを作りながら手を動かす
3. `mise exec -- ruby <file.rb>` で動作を確認する
