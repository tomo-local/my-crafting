package kv

import (
	"testing"
	"os"

	"github.com/stretchr/testify/assert"
)

func TestKVBasic(t *testing.T) {
	kv := KV{}
	kv.Log.FileName = ".test_db"
	defer os.Remove(kv.Log.FileName)

	os.Remove(kv.Log.FileName)
	err := kv.Open()
	assert.Nil(t, err)
	defer kv.Close()

  // データの保存
	updated, err := kv.Set([]byte("k1"), []byte("v1"))
	assert.True(t, updated && err == nil)

	// データの取得
	val, ok, err := kv.Get([]byte("k1"))
	assert.True(t, string(val) == "v1" && ok && err == nil)

	// 存在しないデータの取得
	_, ok, err = kv.Get([]byte("xxx"))
	assert.True(t, !ok && err == nil)

	// 存在しないデータの削除
	updated, err = kv.Del([]byte("xxx"))
	assert.True(t, !updated && err == nil)

	// データの削除
	updated, err = kv.Del([]byte("k1"))
	assert.True(t, updated && err == nil)

	// データの削除の確認
	_, ok, err = kv.Get([]byte("xxx"))
	assert.True(t, !ok && err == nil)

	updated, err = kv.Set([]byte("k2"), []byte("v2"))
	assert.True(t, updated && err == nil)

	// 再度開く
	kv.Close()
	err = kv.Open()
	assert.Nil(t, err)

	_, ok, err = kv.Get([]byte("k1"))
	assert.True(t, !ok && err == nil)
	val, ok, err = kv.Get([]byte("k2"))
	assert.True(t, string(val) == "v2" && ok && err == nil)
}
