package build_database

import "testing"

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

	kv := KV{}
	if err := kv.Open(); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer kv.Close()

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
		})
	}
}

func TestKV_Get(t *testing.T) {
	kv := KV{}
	if err := kv.Open(); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer kv.Close()

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
	kv := KV{}
	if err := kv.Open(); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer kv.Close()

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
		})
	}
}
