package table

import (
	"errors"
	"slices"
)

type Schema struct {
	Table string   // テーブル名
	Cols  []Column // カラムのリスト
	PKey  []int    // どのカラムが主キーか？（インデックス番号の配列）
}

type Column struct {
	Name string
	Type CellType
}

type Row []Cell

func (schema *Schema) NewRow() Row {
	return make(Row, len(schema.Cols))
}

func (row Row) EncodeKey(schema *Schema) []byte {
	key := []byte{}
	key = append(key, []byte(schema.Table)...)
	key = append(key, 0x00)

	if len(row) != len(schema.Cols) {
		panic("col count error")
	}

	for idx, value := range row {
		if value.Type != schema.Cols[idx].Type {
			panic("type error")
		}

		if slices.Contains(schema.PKey, idx) {
			key = row[idx].Encode(key)
		}
	}

	return key
}

func (row Row) EncodeVal(schema *Schema) []byte {
	if len(row) != len(schema.Cols) {
		panic("col count error")
	}

	val := []byte{}
	for idx, value := range row {
		if value.Type != schema.Cols[idx].Type {
			panic("type error")
		}

		if !slices.Contains(schema.PKey, idx) {
			val = row[idx].Encode(val)
		}
	}
	return val
}

func (row Row) DecodeKey(schema *Schema, key []byte) error {
	if len(row) != len(schema.Cols) {
		panic("col count error")
	}

	if len(key) < len(schema.Table)+1 {
		return errors.New("bad key")
	}
	if string(key[:len(schema.Table)+1]) != schema.Table+"\x00" {
		return errors.New("bad key")
	}
	key = key[len(schema.Table)+1:]

	for idx, col := range schema.Cols {
		if !slices.Contains(schema.PKey, idx) {
			continue
		}
		row[idx] = Cell{Type: col.Type}
		decodedKey, err := row[idx].Decode(key)
		if err != nil {
			return err
		}
		key = decodedKey
	}

	if len(key) != 0 {
		return errors.New("trailing garbage")
	}
	return nil

}

func (row Row) DecodeVal(schema *Schema, val []byte) error {
	if len(row) != len(schema.Cols) {
		panic("col count error")
	}

	for idx, col := range schema.Cols {
		if slices.Contains(schema.PKey, idx) {
			continue
		}
		row[idx] = Cell{Type: col.Type}
		decodedVal, err := row[idx].Decode(val)
		if err != nil {
			return err
		}
		val = decodedVal
	}

	if len(val) != 0 {
		return errors.New("trailing garbage")
	}
	return nil
}
