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