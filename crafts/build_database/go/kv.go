package build_database

type KV struct {
	mem map[string][]byte
}

func (kv *KV) Open() error {
	kv.mem = map[string][]byte{}
	return nil
}

func (kv *KV) Close() error { return nil }

func (kv *KV) Get(key []byte) (val []byte, ok bool, err error) {
	return kv.mem[string(key)], true, nil
}
func (kv *KV) Set(key []byte, val []byte) (updated bool, err error) {
	return true, nil
}
func (kv *KV) Del(key []byte) (deleted bool, err error) {
	return true, nil
}
