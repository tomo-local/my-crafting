package build_database

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func newTestKV(t *testing.T) *KV {
	t.Helper()

	path := filepath.Join(t.TempDir(), "kv.log")
	kv := &KV{log: Log{FileName: path}}
	if err := kv.Open(); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { kv.Close() })

	return kv
}

func readLogEntries(t *testing.T, path string) []*Entry {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	var entries []*Entry
	for {
		ent := &Entry{}
		if err := ent.Decode(f); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		entries = append(entries, ent)
	}
	return entries
}

func TestKV_Set(t *testing.T) {
	tests := []struct {
		name        string
		key         []byte
		val         []byte
		wantUpdated bool
	}{
		{name: "new key", key: []byte("k1"), val: []byte("v1"), wantUpdated: true},
		{name: "overwrite existing key", key: []byte("k1"), val: []byte("v2"), wantUpdated: true},
	}

	kv := newTestKV(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := kv.Set(tt.key, tt.val)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}
			if updated != tt.wantUpdated {
				t.Errorf("Set() updated = %v, want %v", updated, tt.wantUpdated)
			}

			got, ok, err := kv.Get(tt.key)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if !ok || string(got) != string(tt.val) {
				t.Errorf("Get() = (%q, %v), want (%q, true)", got, ok, tt.val)
			}

			// Set always appends a new log entry (even on overwrite), so the
			// most recently written entry must reflect this call.
			entries := readLogEntries(t, kv.log.FileName)
			last := entries[len(entries)-1]
			if string(last.key) != string(tt.key) || string(last.val) != string(tt.val) || last.deleted {
				t.Errorf("log last entry = (key=%q, val=%q, deleted=%v), want (key=%q, val=%q, deleted=false)", last.key, last.val, last.deleted, tt.key, tt.val)
			}
		})
	}
}

func TestKV_Get(t *testing.T) {
	kv := newTestKV(t)

	if _, err := kv.Set([]byte("k1"), []byte("v1")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	tests := []struct {
		name    string
		key     []byte
		wantVal []byte
		wantOk  bool
	}{
		{name: "existing key", key: []byte("k1"), wantVal: []byte("v1"), wantOk: true},
		{name: "nonexistent key", key: []byte("xxx"), wantOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := kv.Get(tt.key)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if ok != tt.wantOk {
				t.Errorf("Get() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && string(got) != string(tt.wantVal) {
				t.Errorf("Get() val = %q, want %q", got, tt.wantVal)
			}
		})
	}
}

func TestKV_Del(t *testing.T) {
	kv := newTestKV(t)

	if _, err := kv.Set([]byte("k1"), []byte("v1")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	tests := []struct {
		name        string
		key         []byte
		wantDeleted bool
	}{
		{name: "existing key", key: []byte("k1"), wantDeleted: true},
		{name: "nonexistent key", key: []byte("xxx"), wantDeleted: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the log before Del: a no-op Del (nonexistent key) must not
			// append anything, so we compare against this count afterwards.
			before := len(readLogEntries(t, kv.log.FileName))

			deleted, err := kv.Del(tt.key)
			if err != nil {
				t.Fatalf("Del() error = %v", err)
			}
			if deleted != tt.wantDeleted {
				t.Errorf("Del() = %v, want %v", deleted, tt.wantDeleted)
			}

			if _, ok, err := kv.Get(tt.key); err != nil {
				t.Fatalf("Get() error = %v", err)
			} else if ok {
				t.Errorf("Get() ok = true after Del(), want false")
			}

			entries := readLogEntries(t, kv.log.FileName)
			if tt.wantDeleted {
				// A real delete appends a tombstone entry (deleted=true) for the key.
				last := entries[len(entries)-1]
				if string(last.key) != string(tt.key) || !last.deleted {
					t.Errorf("log last entry = (key=%q, deleted=%v), want (key=%q, deleted=true)", last.key, last.deleted, tt.key)
				}
			} else if len(entries) != before {
				t.Errorf("log has %d entries, want %d (Del on a nonexistent key must not write)", len(entries), before)
			}
		})
	}
}

func TestKV_Open_RecoversFromLog(t *testing.T) {
	type op struct {
		del bool
		key []byte
		val []byte
	}

	tests := []struct {
		name    string
		ops     []op
		key     []byte
		wantVal []byte
		wantOk  bool
	}{
		{
			name:    "set then reopen",
			ops:     []op{{key: []byte("k1"), val: []byte("v1")}},
			key:     []byte("k1"),
			wantVal: []byte("v1"),
			wantOk:  true,
		},
		{
			name: "set twice then reopen keeps latest value",
			ops: []op{
				{key: []byte("k1"), val: []byte("v1")},
				{key: []byte("k1"), val: []byte("v2")},
			},
			key:     []byte("k1"),
			wantVal: []byte("v2"),
			wantOk:  true,
		},
		{
			name: "set then delete then reopen",
			ops: []op{
				{key: []byte("k1"), val: []byte("v1")},
				{del: true, key: []byte("k1")},
			},
			key:    []byte("k1"),
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "kv.log")

			kv := &KV{log: Log{FileName: path}}
			if err := kv.Open(); err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			for _, o := range tt.ops {
				if o.del {
					if _, err := kv.Del(o.key); err != nil {
						t.Fatalf("Del() error = %v", err)
					}
				} else {
					if _, err := kv.Set(o.key, o.val); err != nil {
						t.Fatalf("Set() error = %v", err)
					}
				}
			}
			if err := kv.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}

			reopened := &KV{log: Log{FileName: path}}
			if err := reopened.Open(); err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			defer reopened.Close()

			got, ok, err := reopened.Get(tt.key)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if ok != tt.wantOk {
				t.Errorf("Get() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && string(got) != string(tt.wantVal) {
				t.Errorf("Get() val = %q, want %q", got, tt.wantVal)
			}
		})
	}
}
