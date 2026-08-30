## 学習メモ

### osのconstantについて
これはアクセスする際のモードを指定するための定数

-	O_RDONLY int = syscall.O_RDONLY // 読み取り専用
-	O_WRONLY int = syscall.O_WRONLY // 書き込み専用
-	O_RDWR   int = syscall.O_RDWR   // 読み書き専用

これは動作を制御する追加フラグ（複数の組み合わせ可能）

-	O_APPEND int = syscall.O_APPEND // 追記モード。書き込みを常にファイルの最後尾から行う。
-	O_CREATE int = syscall.O_CREAT  // ファイルが存在しない場合に、新しく作成する。
-	O_EXCL   int = syscall.O_EXCL   // すでにファイルが作成されている場合はエラーにする。O_CREATEと一緒に使用する。
-	O_SYNC   int = syscall.O_SYNC   // 同期モード。書き込みの命令を出した時に、OSのキャッシュに留めず、物理的なディスクに書き込まれるまで待機します。データの安全性を優先するので、速度が遅くなる。
-	O_TRUNC  int = syscall.O_TRUNC  // ファイルを開いたときに、中身をからにします。上書きして作成し直すときにつかう。

### 使い方例
```go
// 読み書き | 作成 | 追記
flags := os.O_RDWR | os.O_CREATE | os.O_APPEND
```


## Effective Go と Bare Return

Bare Returnの書き方
```go
//                ここに戻り値の名前（val, ok, err）が定義されている
func (kv *KV) Get(key []byte) (val []byte, ok bool, err error) {
  val, ok =  kv.mem[string(key)]
  return // 何も書かなくても、現在の val, ok, err の値が返される
}
```

Effective Goの書き方
```go
func (kv *KV) Get(key []byte) ([]byte, bool, error) {
    // 戻り値に名前を付けず、直接 return します
    val, ok := kv.mem[string(key)]
    return val, ok, nil
}
```
