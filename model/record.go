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

func (r *Record) Serialize() []byte {
	keyBytes := []byte(r.Key)
	keySize := uint32(len(keyBytes))
	valueSize := uint32(len(r.Value))

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

func GetRecordSize(data []byte) (int, error) {
	if len(data) < 21 {
		return 0, ErrCorruptedRecord
	}
	keySize := int(binary.LittleEndian.Uint32(data[13:17]))
	valueSize := int(binary.LittleEndian.Uint32(data[17:21]))
	return 21 + keySize + valueSize, nil
}

func DeserializeRecord(data []byte) (*Record, error) {
	if len(data) < 21 {
		return nil, ErrCorruptedRecord
	}

	storedCRC := binary.LittleEndian.Uint32(data[0:4])
	computedCRC := crc32.ChecksumIEEE(data[4:])
	if storedCRC != computedCRC {
		return nil, ErrCRCMismatch
	}

	offset := 4

	timestamp := binary.LittleEndian.Uint64(data[offset:])
	offset += 8

	tombstone := data[offset] == 1
	offset++

	keySize := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	valueSize := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	if len(data) < offset+int(keySize)+int(valueSize) {
		return nil, ErrCorruptedRecord
	}

	key := string(data[offset : offset+int(keySize)])
	offset += int(keySize)

	value := make([]byte, valueSize)
	copy(value, data[offset:offset+int(valueSize)])

	return &Record{
		Key:       key,
		Value:     value,
		Tombstone: tombstone,
		Timestamp: timestamp,
	}, nil
}

func (r *Record) Size() int {
	return 4 + 8 + 1 + 4 + 4 + len(r.Key) + len(r.Value)
}
