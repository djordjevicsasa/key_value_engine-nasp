package model

import (
	"encoding/binary"
	"hash/crc32"
	"time"
)

type Record struct {
	Key       string
	Value     []byte
	Tombstone bool
	Timestamp uint64
}

func NewRecord(key string, value []byte, tombstone bool) *Record {
	return &Record{
		Key:       key,
		Value:     value,
		Tombstone: tombstone,
		Timestamp: uint64(time.Now().UnixNano()),
	}
}

// Format: [CRC(4)] [Timestamp(8)] [Tombstone(1)] [KeySize(4)] [ValueSize(4)] [Key] [Value]
func (r *Record) Serialize() []byte {
	keyBytes := []byte(r.Key)
	keySize := uint32(len(keyBytes))
	valueSize := uint32(len(r.Value))

	// Izracunamo ukupnu velicinu: CRC + Timestamp + Tombstone + KeySize + ValueSize + Key + Value
	totalSize := 4 + 8 + 1 + 4 + 4 + int(keySize) + int(valueSize)
	buf := make([]byte, totalSize)

	offset := 4

	binary.LittleEndian.PutUint64(buf[offset:], r.Timestamp)
	offset += 8

	if r.Tombstone {
		buf[offset] = 1
	} else {
		buf[offset] = 0
	}
	offset++

	binary.LittleEndian.PutUint32(buf[offset:], keySize)
	offset += 4

	binary.LittleEndian.PutUint32(buf[offset:], valueSize)
	offset += 4

	copy(buf[offset:], keyBytes)
	offset += int(keySize)

	copy(buf[offset:], r.Value)

	crc := crc32.ChecksumIEEE(buf[4:])
	binary.LittleEndian.PutUint32(buf[0:], crc)

	return buf
}
