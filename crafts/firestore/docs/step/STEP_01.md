# Step 1：In-memory Document Store（前提知識）

## このステップで何を作るか

HTTP もリアルタイムもまだない。ドキュメントの **読み書き削除** と、各操作に **スナップショットバージョン** を付与するストレージ層を作る。

---

## コレクション / ドキュメントのデータ構造

Firestore のパスは `/collections/{col}/documents/{doc}` 形式。このクラフトでは省略して `{col}/{doc}` と表現する。

```
Store
└── data: map[path]→Document
         ├── "users/alice" → {fields: {...}, version: 3}
         └── "users/bob"  → {fields: {...}, version: 1}
```

**Document 型のフィールド**:
- `Fields map[string]interface{}` — JSON として受け取った任意のフィールド
- `Version SnapshotVersion` — 書き込み時に付与される単調増加番号

---

## スナップショットバージョンの役割

```
書き込み 1: version = 1
書き込み 2: version = 2
書き込み 3: version = 3
```

Store はグローバルなカウンターを持ち、**書き込みのたびにインクリメントして Document に付与する**。

この値は後の Step 6（Resume token）で「どこまで受信したか」の基準として使う。

---

## 並行安全性

HTTP サーバーを立てると複数の goroutine が Store に同時アクセスする。  
`sync.RWMutex` で保護する:

- **Read 系**（Get, List）: `RLock` / `RUnlock`
- **Write 系**（Put, Delete）: `Lock` / `Unlock`

---

## 📌 まとめ: Step 1 のフロー

1. `SnapshotVersion` 型（`uint64`）を定義する
2. `Document` 構造体（`Fields`, `Version`）を定義する
3. `Store` 構造体（`data map[string]*Document`, `version SnapshotVersion`, `mu sync.RWMutex`）を定義する
4. `New() *Store` — 初期化
5. `Put(path string, fields map[string]interface{}) SnapshotVersion` — バージョンをインクリメントして保存し、付与したバージョンを返す
6. `Get(path string) (*Document, bool)` — ドキュメントを返す
7. `Delete(path string) (SnapshotVersion, bool)` — 削除してバージョンを返す
8. `List(collection string) []*Document` — コレクション内の全ドキュメントを返す
