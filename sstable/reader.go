package sstable

import (
	"encoding/binary"

	"nasp-kv-engine/block"
	"nasp-kv-engine/bloom"
	"nasp-kv-engine/merkle"
	"nasp-kv-engine/model"
)

// Ucitava Bloom Filter iz Filter fajla — koristi se za brzu proveru pre pretrage
func LoadFilter(filterPath string, mgr *block.CachedManager) (*bloom.BloomFilter, error) {
	data, err := mgr.ReadFile(filterPath)
	if err != nil {
		return nil, err
	}
	bf := bloom.DeserializeBloomFilter(data)
	if bf == nil {
		return nil, model.ErrCorruptedRecord
	}
	return bf, nil
}

/ Ucitava Summary strukturu iz fajla
// Vraca min kljuc, max kljuc i listu summary zapisa
func LoadSummary(summaryPath string, mgr *block.CachedManager) (string, string, []SummaryEntry, error) {
	data, err := mgr.ReadFile(summaryPath)
	if err != nil {
		return "", "", nil, err
	}

	if len(data) < 12 {
		return "", "", nil, model.ErrCorruptedRecord
	}

	offset := 0

	// Citamo min kljuc
	if offset+4 > len(data) {
		return "", "", nil, model.ErrCorruptedRecord
	}
	minKeySize := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if offset+minKeySize > len(data) {
		return "", "", nil, model.ErrCorruptedRecord
	}
	minKey := string(data[offset : offset+minKeySize])
	offset += minKeySize

	// Max kljuc
	if offset+4 > len(data) {
		return "", "", nil, model.ErrCorruptedRecord
	}
	maxKeySize := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if offset+maxKeySize > len(data) {
		return "", "", nil, model.ErrCorruptedRecord
	}
	maxKey := string(data[offset : offset+maxKeySize])
	offset += maxKeySize

	// Broj summary zapisa
	if offset+4 > len(data) {
		return "", "", nil, model.ErrCorruptedRecord
	}
	count := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	entries := make([]SummaryEntry, 0, count)
	for i := 0; i < count && offset < len(data); i++ {
		if offset+4 > len(data) {
			return "", "", nil, model.ErrCorruptedRecord
		}
		keySize := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4

		if offset+keySize > len(data) {
			return "", "", nil, model.ErrCorruptedRecord
		}
		key := string(data[offset : offset+keySize])
		offset += keySize

		if offset+8 > len(data) {
			return "", "", nil, model.ErrCorruptedRecord
		}
		indexOffset := binary.LittleEndian.Uint64(data[offset : offset+8])
		offset += 8

		entries = append(entries, SummaryEntry{
			Key:         key,
			IndexOffset: indexOffset,
		})
	}

	return minKey, maxKey, entries, nil




}

// Pretrazuje Summary da pronadje opseg u Index fajlu gde kljuc moze biti
// Vraca pocetni i krajnji offset u Index fajlu
func SearchSummary(entries []SummaryEntry, key string) (uint64, uint64) {
	if len(entries) == 0 {
		return 0, 0
	}

	startOffset := entries[0].IndexOffset
	endOffset := uint64(0) // 0 znaci "do kraja fajla"

	for i := 0; i < len(entries); i++ {
		if entries[i].Key <= key {
			startOffset = entries[i].IndexOffset
			if i+1 < len(entries) {
				endOffset = entries[i+1].IndexOffset
			} else {
				endOffset = 0
			}
		} else {
			break
		}
	}

	return startOffset, endOffset
}