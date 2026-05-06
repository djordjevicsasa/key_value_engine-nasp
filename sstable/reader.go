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
}