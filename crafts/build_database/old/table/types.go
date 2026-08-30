package table

import (
	"encoding/binary"
	"errors"
	"slices"
)

type CellType uint8

const (
	TypeInt64 CellType = 1
	TypeStr CellType = 2
)

type Cell struct {
	Type CellType
	Int64 int64
	Str []byte
}

func (cell *Cell) Encode(toAppend []byte) []byte {
	switch cell.Type {
	case TypeInt64:
		return binary.LittleEndian.AppendUint64(toAppend, uint64(cell.Int64))
	case TypeStr:
		// Strの場合は、最初にlengthを4byteで指定してから、dataを設定するようにします。
		toAppend = binary.LittleEndian.AppendUint32(toAppend, uint32(len(cell.Str)))
		return append(toAppend, cell.Str...)
	default:
		panic("unreachable")
	}
}

func (cell *Cell) Decode(data []byte) (rest []byte, err error) {
	switch cell.Type {
	case TypeInt64:
		// 64bit = 8bytesなので、8より小さい場合はエラー
		if len(data) < 8 {
			return data, errors.New("expect more data")
		}
		cell.Int64 = int64(binary.LittleEndian.Uint64(data[0:8]))
		return data[8:], nil
	case TypeStr:
		// 最初に文字数を4bytesで指定しているので、4より小さい場合はエラー
		if len(data) < 4 {
			return data, errors.New("expect more data")
		}
		// 最初の文字数を数値に変換して、4bytesと文字数の値がdataの長さより小さくないか確認
		size := int(binary.LittleEndian.Uint32(data[0:4]))
		if len(data) < 4+size {
			return data, errors.New("expect more data")
		}
		// 新しいメモリ場所に値をコピーしている
		cell.Str = slices.Clone(data[4 : 4+size])
		return data[4+size:], nil
	default:
		panic("unreachable")
	}
}
