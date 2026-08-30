package kv

import (
	"bytes"
)

type KV struct {
	Log Log
	mem map[string][]byte
}

func (kv *KV) Open() error {
	if err := kv.Log.Open(); err != nil {
		return err
	}

	kv.mem = map[string][]byte{}
	for {
		ent := Entry{}
		eof, err := kv.Log.Read(&ent)
		if err != nil {
			return err
		}

		if eof {
			break
		}

		if ent.deleted {
			delete(kv.mem, string(ent.key))
		} else {
			kv.mem[string(ent.key)] = ent.val
		}
	}
	return nil
}

func (kv *KV) Close() error { return kv.Log.Close() }

// 取得
func (kv *KV) Get(key []byte) ([]byte, bool, error) {
	val, ok := kv.mem[string(key)]
	return val, ok, nil
}

type UpdateMode int

const (
	ModeUpsert UpdateMode = 0 // 挿入または更新 (Upsert)
	ModeInsert UpdateMode = 1 // 新規挿入のみ (Insert new)
	ModeUpdate UpdateMode = 2 // 既存の更新のみ (Update existing)
)

func (kv *KV) SetEx(key []byte, val []byte, mode UpdateMode) (bool, error) {
	prev, exist := kv.mem[string(key)]

	var updated bool
	switch mode {
	case ModeUpsert:
		updated = !exist || !bytes.Equal(prev, val)
	case ModeInsert:
		updated = !exist
	case ModeUpdate:
		updated = exist
	default:
		panic("unreachable")
	}

	if updated {
		if err := kv.Log.Write(&Entry{key: key, val: val}); err != nil {
			return false, err
		}
		kv.mem[string(key)] = val
	}

	return updated, nil
}

// 保存
func (kv *KV) Set(key []byte, val []byte) (bool, error) {
	return kv.SetEx(key, val, ModeUpsert)
}

// 削除
func (kv *KV) Del(key []byte) (bool, error) {
	_, deleted := kv.mem[string(key)]
	if deleted {
		if err := kv.Log.Write(&Entry{key: key, deleted: true}); err != nil {
			return false, err
		}
		delete(kv.mem, string(key))
	}
	return deleted, nil
}
