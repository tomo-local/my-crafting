package build_database

import (
	"encoding/binary"
	"io"
)

type Entry struct {
	key []byte
	val []byte
}

func (ent *Entry) Encode() []byte {
	// | key size | val size | key data | val data |
	// | 4 bytes  | 4 bytes  |   ...    |   ...    |
	data := make([]byte, 4+4+len(ent.key)+len(ent.val))
	// A length is never negative, so we store it as an unsigned uint32.
	binary.LittleEndian.PutUint32(data[0:4], uint32(len(ent.key)))
	// Same here: the value length can't be negative, so uint32 is used.
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(ent.val)))

	copy(data[8:], ent.key)
	// Offset 8 skips the two 4-byte length headers, and len(ent.key) skips the key data.
	copy(data[8+len(ent.key):], ent.val)
	return data
}

func (ent *Entry) Decode(r io.Reader) error {
	// header is two 4-byte fields holding the byte lengths of the key and val data that follow.
	// | key size | val size | key data | val data |
	// | 4 bytes  | 4 bytes  |   ...    |   ...    |
	var header [8]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return err
	}

	// key size
	keyLen := binary.LittleEndian.Uint32(header[0:4])
	// val size
	valLen := binary.LittleEndian.Uint32(header[4:8])

	data := make([]byte, keyLen+valLen)
	if _, err := io.ReadFull(r, data); err != nil {
		return err
	}

	ent.key = data[:keyLen]
	ent.val = data[keyLen:]

	return nil
}
