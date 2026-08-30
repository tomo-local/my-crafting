package table

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableCell(t *testing.T) {
	cell := Cell{Type: TypeInt64, Int64: -2}
	data := []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	assert.Equal(t, data, cell.Encode(nil))
	decoded := Cell{Type: TypeInt64}
	rest, err := decoded.Decode(data)
	assert.True(t, len(rest) == 0 && err == nil)
	assert.Equal(t, cell, decoded)

	cell = Cell{Type: TypeStr, Str: []byte("asdf")}
	data = []byte{4, 0, 0, 0, 'a', 's', 'd', 'f'}
	assert.Equal(t, data, cell.Encode(nil))
	decoded = Cell{Type: TypeStr}
	rest, err = decoded.Decode(data)
	assert.True(t, len(rest) == 0 && err == nil)
	assert.Equal(t, cell, decoded)
}
