# 作って、壊して、直して学ぶ Kubernetes 入門

書籍「作って、壊して、直して学ぶ Kubernetes 入門」の学習ログ。

## ディレクトリ構成

```
build-break-fix/
├── README.md
├── part1-build/                # Part 1: つくってみようKubernetes
│   ├── ch01-docker/
│   ├── ch02-cluster/
│   ├── ch03-overview/
│   └── ch04-app/
├── part2-break-fix/            # Part 2: アプリケーションを壊して学ぶKubernetes（メイン）
│   ├── ch05-troubleshoot/
│   ├── ch06-resources/
│   ├── ch07-stateless/
│   └── ch08-final/
└── part3-resilience/           # Part 3: 壊れても動くKubernetes
    ├── ch09-architecture/
    ├── ch10-workflow/
    ├── ch11-observability/
    └── ch12-next/
```

### Part 1: メモ・手順中心

各 chapter に `notes.md` を置き、手順・気づきを記録する。

### Part 2: Break-Fix シナリオ（メイン）

```
ch0X-<name>/
├── notes.md
└── scenario-NN/
    ├── broken.yaml   # 意図的に壊れたマニフェスト
    ├── fixed.yaml    # 修正済みマニフェスト
    └── hint.md       # 原因・修正ポイント
```

### Part 3: 概念理解・観察ログ

各 chapter に `notes.md` を置き、アーキテクチャ理解や実行ログを記録する。

---

## 進捗

### Part 1: つくってみようKubernetes

- [x] Chapter 1 — Dockerコンテナをつくってみる
- [x] Chapter 2 — Kubernetesクラスタをつくってみる
- [x] Chapter 3 — 全体像の説明
- [x] Chapter 4 — アプリケーションをKubernetesクラスタ上につくる

### Part 2: アプリケーションを壊して学ぶKubernetes

- [ ] Chapter 5 — トラブルシューティングガイドとkubectlコマンドの使い方
- [ ] Chapter 6 — Kubernetesリソースをつくって壊そう
- [ ] Chapter 7 — 安全なステートレス・アプリケーションをつくるために
- [ ] Chapter 8 — 総復習：アプリケーションを直そう

### Part 3: 壊れても動くKubernetes

- [ ] Chapter 9  — Kubernetesの仕組み、アーキテクチャーを理解しよう
- [ ] Chapter 10 — Kubernetesの開発ワークフローを理解しよう
- [ ] Chapter 11 — オブザーバビリティとモニタリングに触れてみよう
- [ ] Chapter 12 — この先の歩み方
