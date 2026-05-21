# Step 7：Connection ヘッダーとグレースフルな接続終了（前提知識）

Step 6 では Keep-Alive ループを実装しましたが、クライアントが `Connection: close` を送ってきたときの処理を省略していました。Step 7 では **`Connection` ヘッダーを見て接続を維持するか切るかを判断する** 処理を追加します。

---

## 1. `Connection` ヘッダーの役割

HTTP/1.1 のデフォルトは Keep-Alive です。しかしクライアントは `Connection: close` を送ることで「このレスポンスを受け取ったら接続を切ってよい」と伝えられます。

```
GET / HTTP/1.1
Host: localhost:8080
Connection: close        ← これが来たらレスポンス後に切断する
```

サーバーはそれに応答して、レスポンスヘッダーにも同じ値を返すのが作法です。

```
HTTP/1.1 200 OK
Content-Length: 8
Connection: close        ← クライアントに「切ります」と伝える
```

---

## 2. HTTP/1.0 との違い

| バージョン | デフォルト | 持続させたい場合 | 切りたい場合 |
|---|---|---|---|
| HTTP/1.0 | 切断 | `Connection: keep-alive` を送る | 何も送らない |
| HTTP/1.1 | Keep-Alive | 何も送らない | `Connection: close` を送る |

実装上は「バージョンが 1.0 なら close をデフォルト、1.1 なら keep-alive をデフォルト」として、`Connection` ヘッダーで上書きするのが正確な挙動です。Step 7 では HTTP/1.1 のみを対象にします。

---

## 3. サーバー側で必要な変更

### (1) リクエストのパース

`Connection` ヘッダーの値を読み取って保持します。

```
Connection: close\r\n
    ↓ ":" で分割・trim
"close"
```

### (2) レスポンスヘッダーの設定

クライアントから来た値をそのまま返します。値が空なら `keep-alive` を返します。

```
Connection: keep-alive\r\n   ← デフォルト
Connection: close\r\n        ← クライアントが close を要求した場合
```

### (3) 接続ループの終了判定

レスポンスを書いた後、`Connection` が `close` ならループを抜けて接続を閉じます。

```
for {
    Read → Parse → Route → Write
    if connection == "close" → break
}
Close
```

---

## 4. `Connection` ヘッダーがない場合

クライアントが `Connection` ヘッダーを送らなかった場合、HTTP/1.1 のデフォルトは Keep-Alive なので接続を維持し続けます。

---

## 📌 まとめ：Step 7 のフロー

1. ヘッダーループ内で `Connection:` を見つけたら値を保持する
2. ルーティング・レスポンス書き込み時に `Connection` ヘッダーを付与する
3. レスポンス書き込み後、値が `close` ならループを抜けて **`Close`** する
4. それ以外はループを継続して次のリクエストを待つ
