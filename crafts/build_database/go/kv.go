package build_database

type KV struct {
	log Log
	mem map[string][]byte
}

func (kv *KV) Open() error {
	if err := kv.log.Open(); err != nil {
		return err
	}

	kv.mem = map[string][]byte{}

	for {
		ent := &Entry{}
		eof, err := kv.log.Read(ent)
		if err != nil {
			return err
		}
		if eof {
			break
		}
		if ent.deleted {
			delete(kv.mem, string(ent.key))
			continue
		}

		kv.mem[string(ent.key)] = ent.val
	}

	return nil
}

func (kv *KV) Close() error { return kv.log.Close() }

func (kv *KV) Get(key []byte) (val []byte, ok bool, err error) {
	val, ok = kv.mem[string(key)]
	return val, ok, nil
}
func (kv *KV) Set(key []byte, val []byte) (updated bool, err error) {
	origin, ok := kv.mem[string(key)]
	if ok && string(origin) == string(val) {
		return false, nil
	}

	if err := kv.log.Write(&Entry{key: key, val: val}); err != nil {
		return false, err
	}

	kv.mem[string(key)] = val
	return true, nil
}

func (kv *KV) Del(key []byte) (deleted bool, err error) {
	_, ok := kv.mem[string(key)]
	if !ok {
		return false, nil
	}

	if err := kv.log.Write(&Entry{key: key, deleted: true}); err != nil {
		return false, err
	}

	delete(kv.mem, string(key))
	return true, nil
}
