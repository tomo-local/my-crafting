# Step 7: 統合確認（3日）

Step 2〜6 の全パターンが連携していることを確認する。

---

## 最終課題

「求職者がスカウトに応募する機能」を全パターンを使って実装する。

```
1. Controller が UseCase を呼ぶ
2. UseCase が Service・Query・Strategy を組み合わせる
3. DB への保存はトランザクション内
4. メール・SMS 送信はトランザクション外
5. 各層が独立してテスト可能
```

---

## 目標ディレクトリ構造

```
app/
  use_cases/
    scouts/
      send_scout.rb           # Step 6
  services/
    entries/
      create_entry.rb         # Step 2
    scouts/
      send_scout.rb           # Step 2
      matching_strategies/
        base_strategy.rb      # Step 5
        driver_strategy.rb    # Step 5
        factory_strategy.rb   # Step 5
  queries/
    jobs/
      recommended_jobs_query.rb    # Step 3
      not_applied_jobs_query.rb    # Step 3
    scouts/
      target_search_query.rb       # Step 3
  presenters/
    job_presenter.rb          # Step 4
    user_presenter.rb         # Step 4
  models/
    user.rb    # 薄い（Association・Validation のみ）
    job.rb     # 薄い
    entry.rb
    scout.rb
```

---

## 習得の判断基準

| パターン | 習得の基準 |
|---|---|
| Service Object | ロジックを Service に移す判断がすぐできる |
| Query Object | `scope` と Query Object の使い分けを即答できる |
| Presenter | どの表示ロジックをどの Presenter に移すか即答できる |
| Strategy | コピペを見て、Strategy への統合手順を説明できる |
| Use Case | 複雑な操作の責務分割を設計できる |

---

## 確認ポイント

- [ ] 新しい職種を追加するとき、変更するファイルが 1 つだけ
- [ ] バグを修正するとき、どのファイルを開けばいいかすぐ分かる
- [ ] `User` / `Job` モデルが Association と Validation だけになっている
- [ ] Controller が「UseCase を呼ぶだけ」になっている
