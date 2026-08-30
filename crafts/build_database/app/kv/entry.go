package kv

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"errors"
)

type Entry struct {
	key []byte
	val []byte
	deleted bool
}

func (ent *Entry) Encode() []byte {
	// 初期のデータ構造を定義 （checksum）4byte+(key)4byte+（val)4byte+(deleted)1byte+keyサイズ+valサイズ
	data := make([]byte, 4+4+4+1+len(ent.key)+len(ent.val))
	// keyの文字数
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(ent.key)))
	// keyの値
	copy(data[4+4+4+1:], ent.key)
	// deletedの値
	// 削除されるので、valの値は設定しない
	if ent.deleted {
		data[4+4+4] = 1
	} else {
		// deletedがfalseなので、値が設定される
		data[4+4+4] = 0
		// valの文字数
		binary.LittleEndian.PutUint32(data[8:12], uint32(len(ent.val)))
		copy(data[4+4+4+1+len(ent.key):], ent.val)
	}
	// ここでchecksumを計算してヘッダーに追加
	binary.LittleEndian.PutUint32(data[0:4], crc32.ChecksumIEEE(data[4:]))
	return data
}

func (ent *Entry) Decode(r io.Reader) error {
	// 最初のkeyとvalのサイズを格納する変数を定義
	// （checksum）4byte+(key)4byte+（val)4byte+(deleted)1byte
	var header [4 + 4 + 4 + 1]byte
	// [13]bytesまで、データを読み込んでねということ
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return err
	}
	// 記録されているChecksumを取得
	recChecksum := binary.LittleEndian.Uint32(header[0:4])
	// keyの文字数
	keyLen := int(binary.LittleEndian.Uint32(header[4:8]))
	// valの文字数
	valLen := int(binary.LittleEndian.Uint32(header[8:12]))
	// deleted
	deleted := header[12]

	data := make([]byte, keyLen+valLen)
	if _, err := io.ReadFull(r, data); err != nil {
		return err
	}

	// Checksumの検証
	h := crc32.NewIEEE()
	// Checksumの計算対象は「ヘッダーの4バイト目以降」＋「データ本体」
	h.Write(header[4:])
	h.Write(data)
	if h.Sum32() != recChecksum {
		return errors.New("checksum mismatch: data is corrupted")
	}

	// 最初からkeyLenの長さまで取得
	ent.key = data[:keyLen]
	if deleted != 0 {
		ent.deleted = true
	} else {
		ent.deleted = false
		// keyLenの長さから最後まで取得
		ent.val = data[keyLen:]
	}

	return nil
}
