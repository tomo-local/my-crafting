# new-craft

`crafts/<name>/` に新しいクラフトのスキャフォールドを作成するスキルです。

## 実行手順

引数 `$ARGUMENTS` からクラフト名と言語を読み取ってください。
言語が指定されていない場合は Go を使用します。

例:
- `rate-limiter` → Go
- `rate-limiter go` → Go
- `rate-limiter ruby` → Ruby

### Step 1: 共通ファイルの作成

言語に関わらず以下を作成する：

```
crafts/<name>/
├── README.md
└── docs/
    └── images/
        └── .gitkeep
```

**`README.md`**
```markdown
# <Name>

<一言説明>

## Status

🚧 実装中

## 実行方法

（言語ごとの実行コマンドを記載）
```

### Step 2: 言語別のスキャフォールド

#### Go の場合

```
crafts/<name>/go/
├── go.mod
├── main.go
└── internal/
    ├── server/
    │   └── server.go
    ├── request/
    │   └── request.go
    └── response/
        └── response.go
```

**`go/go.mod`**
```
module github.com/tomo-local/<name>

go 1.25.1
```

**`go/main.go`**
```go
package main

import (
	"log/slog"
	"os"

	"github.com/tomo-local/<name>/internal/server"
)

func main() {
	srv := server.NewServer(":8080")
	slog.Info("starting server", "addr", ":8080")
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
```

**`go/internal/server/server.go`**
```go
package server

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) ListenAndServe() error {
	// TODO: implement
	return nil
}
```

**`go/internal/request/request.go`**
```go
package request

type Request struct {
	// TODO: define fields
}
```

**`go/internal/response/response.go`**
```go
package response

// TODO: implement
```

README の実行方法セクション：
```bash
go run ./go/main.go
```

#### Ruby の場合

```
crafts/<name>/ruby/
├── mise.toml
└── main.rb
```

**`ruby/mise.toml`**
```toml
[tools]
ruby = "3.4"
```

**`ruby/main.rb`**
```ruby
# TODO: implement
```

README の実行方法セクション：
```bash
cd ruby
mise exec -- ruby main.rb
```

#### その他の言語

クラフト名のサブディレクトリ（例: `ts/`）を作成し、その言語の慣習に従った最小構成を置く。
`mise.toml` でバージョン管理できる言語はそこに記載する。

### Step 3: `.github/labeler.yml` への追記

既存の `labeler.yml` を読み取り、他のクラフトに倣って新しいクラフトのラベルルールを追記してください。

### Step 4: 完了報告

作成したファイル一覧と、次のステップ（`/craft-learning-plan` でドキュメントを作成する）を案内してください。

---

引数: $ARGUMENTS
