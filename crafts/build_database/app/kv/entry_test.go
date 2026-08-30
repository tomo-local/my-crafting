package kv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryEncode(t *testing.T) {
	ent := Entry{key: []byte("k1"), val: []byte("xxx")}
	// Checksum(4) + klen(4) + vlen(4) + deleted(1) + k1 + xxx
	data := []byte{0xe9, 0xec, 0x4d, 0x9e, 2, 0, 0, 0, 3, 0, 0, 0, 0, 'k', '1', 'x', 'x', 'x'}

	assert.Equal(t, data, ent.Encode())

	decoded := Entry{}
	err := decoded.Decode(bytes.NewBuffer(data))
	assert.Nil(t, err)
	assert.Equal(t, ent, decoded)

	ent = Entry{key: []byte("k1"), deleted: true}
	// Checksum(4) + klen(4) + vlen(4) + deleted(1) + k1
	data = []byte{0x4c, 0xd0, 0xfe, 0xe5, 2, 0, 0, 0, 0, 0, 0, 0, 1, 'k', '1'}
	assert.Equal(t, data, ent.Encode())

	decoded = Entry{}
	err = decoded.Decode(bytes.NewBuffer(data))
	assert.Nil(t, err)
	assert.Equal(t, ent, decoded)
}
