package build_database

import (
	"bytes"
	"testing"
)

func TestEntry_Encode(t *testing.T) {
	tests := []struct {
		name string
		ent  *Entry
		want []byte
	}{
		{
			name: "simple entry",
			ent:  &Entry{key: []byte("k1"), val: []byte("v1")},
			want: []byte{2, 0, 0, 0, 2, 0, 0, 0, 'k', '1', 'v', '1'},
		},
		{
			name: "boundary: val length is 0",
			ent:  &Entry{key: []byte("k1"), val: []byte("")},
			want: []byte{2, 0, 0, 0, 0, 0, 0, 0, 'k', '1'},
		},
		{
			name: "boundary: key length is 0",
			ent:  &Entry{key: []byte(""), val: []byte("v1")},
			want: []byte{0, 0, 0, 0, 2, 0, 0, 0, 'v', '1'},
		},
		{
			name: "longer key and val",
			ent:  &Entry{key: []byte("hello"), val: []byte("world")},
			want: []byte{5, 0, 0, 0, 5, 0, 0, 0, 'h', 'e', 'l', 'l', 'o', 'w', 'o', 'r', 'l', 'd'},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ent.Encode()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntry_Decode(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantKey []byte
		wantVal []byte
	}{
		{
			name:    "simple entry",
			data:    []byte{2, 0, 0, 0, 2, 0, 0, 0, 'k', '1', 'v', '1'},
			wantKey: []byte("k1"),
			wantVal: []byte("v1"),
		},
		{
			name:    "boundary: val length is 0",
			data:    []byte{2, 0, 0, 0, 0, 0, 0, 0, 'k', '1'},
			wantKey: []byte("k1"),
			wantVal: []byte{},
		},
		{
			name:    "boundary: key length is 0",
			data:    []byte{0, 0, 0, 0, 2, 0, 0, 0, 'v', '1'},
			wantKey: []byte{},
			wantVal: []byte("v1"),
		},
		{
			name:    "longer key and val",
			data:    []byte{5, 0, 0, 0, 5, 0, 0, 0, 'h', 'e', 'l', 'l', 'o', 'w', 'o', 'r', 'l', 'd'},
			wantKey: []byte("hello"),
			wantVal: []byte("world"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ent := &Entry{}
			if err := ent.Decode(bytes.NewBuffer(tt.data)); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if !bytes.Equal(ent.key, tt.wantKey) {
				t.Errorf("Decode() key = %q, want %q", ent.key, tt.wantKey)
			}
			if !bytes.Equal(ent.val, tt.wantVal) {
				t.Errorf("Decode() val = %q, want %q", ent.val, tt.wantVal)
			}
		})
	}
}
