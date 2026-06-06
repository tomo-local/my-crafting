# Step 2：プロキシヘッダーの付加（前提知識）

Step 1 では生のTCPストリームをそのまま転送していました。Step 2 ではHTTPリクエストを一度パースしてヘッダーを書き換えてから転送します。これによりアップストリームはクライアントの本来のIPを知れるようになります。

---

## 1. なぜヘッダーを書き換えるのか

Step 1 のままだと、アップストリームが受け取るリクエストには以下の問題があります。

| 問題 | 影響 |
|---|---|
| `Host: localhost:8080`（プロキシのアドレス）のまま | アップストリームが正しいバーチャルホストを判断できない |
| クライアントのIPがわからない | アクセスログにプロキシのIPしか残らない |
| プロキシを経由したことがわからない | 多段プロキシの経路が追えない |

---

## 2. 書き換えが必要な3つのヘッダー

### `Host`

```
書き換え前: Host: localhost:8080   （プロキシのアドレス）
書き換え後: Host: localhost:9001   （アップストリームのアドレス）
```

`Host` ヘッダーはHTTP/1.1では必須フィールドです。アップストリームはこれをもとにバーチャルホストのルーティングを行うため、プロキシのアドレスのままでは正しく動作しない場合があります。

### `X-Forwarded-For`

クライアントの本来のIPアドレスを記録します。

```
新規付加:     X-Forwarded-For: 203.0.113.42
既存がある場合: X-Forwarded-For: 10.0.0.1, 203.0.113.42  （末尾に追加）
```

### `Via`

リクエストが経由したプロキシを記録します（RFC 9110 Section 7.6.3）。

```
Via: 1.1 my-reverse-proxy
```

形式は `プロトコルバージョン スペース プロキシ識別子` です。

---

## 3. HTTPリクエストのパース戦略

Step 1 の `io.Copy` は「中身を見ずにコピー」でした。ヘッダーを書き換えるには一度リクエストを**読んで解析してから再構築**する必要があります。

Goの `net/http` パッケージには `http.ReadRequest(bufio.NewReader(conn))` という関数があり、TCPコネクションから HTTP リクエストを構造体として読み取れます。

```
TCPコネクション
    ↓ bufio.NewReader でバッファリング
    ↓ http.ReadRequest で構造体に変換
http.Request{
    Method: "GET",
    URL: "/path",
    Header: map[string][]string{
        "Host": ["localhost:8080"],
        ...
    },
    Body: io.ReadCloser（ボディのストリーム）
}
```

読み取った `*http.Request` のヘッダーを直接書き換えて、`req.Write(upstreamConn)` で再シリアライズして送信します。

---

## 4. hop-by-hop ヘッダーの除去

「この接続の中だけで有効」なヘッダーは、転送先に送ってはいけません（RFC 9110 Section 7.6.1）。

| 除去すべきヘッダー |
|---|
| `Connection` |
| `Keep-Alive` |
| `Proxy-Authenticate` |
| `Proxy-Authorization` |
| `TE` |
| `Trailers` |
| `Transfer-Encoding` |
| `Upgrade` |

`Connection` ヘッダーの値に列挙されているヘッダーも除去対象です。

---

## 5. ヘッダー操作の一覧

リクエストを転送するまでに行うヘッダー操作をまとめます。

| タイミング | 対象ヘッダー | 操作 | 理由 |
|---|---|---|---|
| 最初 | `Connection` が指定するヘッダー群 | **削除** | hop-by-hop なので転送しない |
| 最初 | `Connection` / `Keep-Alive` / `Transfer-Encoding` など | **削除** | hop-by-hop なので転送しない |
| 次 | `Host` | **上書き** | アップストリームのアドレスに変える |
| 次 | `X-Forwarded-For` | **追加**（既存があれば末尾に連結） | クライアントの本来のIPを伝える |
| 次 | `Via` | **追加** | 経由したプロキシを記録する |

### 操作の順番が重要な理由

hop-by-hop の削除を**最初**に行うのは、`Connection` ヘッダーの値に「削除すべき追加ヘッダー名」が含まれている場合があるためです。先に削除してから他のヘッダーを付加することで、誤って付加したヘッダーを消してしまうリスクをなくします。

### `Host` の書き換えに関する注意

`req.Header.Set("Host", ...)` では書き換わりません。`Host` はヘッダーマップとは別のフィールドとして管理されているため、`req.Host = upstream` と直接フィールドに代入する必要があります。

---

## 📌 まとめ：Step 2 のフロー

1. `Accept` でクライアントの接続を受け取る
2. `http.ReadRequest(bufio.NewReader(clientConn))` でリクエストを解析する
3. hop-by-hop ヘッダーを削除する
4. `Host` ヘッダーをアップストリームのアドレスに書き換える
5. `X-Forwarded-For` にクライアントIPを追加する
6. `Via` ヘッダーを追加する
7. `net.Dial` でアップストリームに接続する
8. `req.Write(upstreamConn)` で書き換えたリクエストを送信する
9. `io.Copy(clientConn, upstreamConn)` でレスポンスをクライアントに返す
