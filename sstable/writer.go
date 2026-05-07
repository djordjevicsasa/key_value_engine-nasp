package sstable

import (
	"encoding/binary"
	"fmt"
	"path/filepath"

	"nasp-kv-engine/block"
	"nasp-kv-engine/bloom"
	"nasp-kv-engine/merkle"
	"nasp-kv-engine/model"
)

// SSTable struktura na disku — svaka tabela ima 5 fajlova:
// Data, Filter (Bloom), Index, Summary, Metadata (Merkle)
type SSTable struct {
	DataPath     string // putanja do Data fajla
	FilterPath   string // putanja do Filter fajla (Bloom)
	IndexPath    string // putanja do Index fajla
	SummaryPath  string // putanja do Summary fajla
	MetadataPath string // putanja do Metadata fajla (Merkle)
	MinKey       string // najmanji kljuc u tabeli
	MaxKey       string // najveci kljuc u tabeli
}

// IndexEntry — jedan zapis u Index strukturi
type IndexEntry struct {
	Key        string
	DataOffset uint64 // pozicija zapisa u Data fajlu
	DataSize   uint32 // velicina zapisa u Data fajlu
}

// SummaryEntry — jedan zapis u Summary strukturi (redje uzorkovanje Index-a)
type SummaryEntry struct {
	Key         string
	IndexOffset uint64 // pozicija u Index fajlu
}

// Generise putanje do svih fajlova SSTable-a u zadatom direktorijumu
func NewSSTablePaths(dir string, id string) *SSTable {
	base := filepath.Join(dir, id)
	return &SSTable{
		DataPath:     base + "_data.db",
		FilterPath:   base + "_filter.db",
		IndexPath:    base + "_index.db",
		SummaryPath:  base + "_summary.db",
		MetadataPath: base + "_metadata.db",
	}
}

// Pisanje SSTable-a na disk od sortiranih zapisa
func WriteSSTable(dir string, id string, records []*model.Record, summaryStep int, bloomExpected int, bloomFPRate float64, mgr *block.CachedManager) (*SSTable, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("nema zapisa za pisanje")
	}

	sst := NewSSTablePaths(dir, id)
	sst.MinKey = records[0].Key
	sst.MaxKey = records[len(records)-1].Key

	// ===== 1. DATA fajl — serijalizovani zapisi jedan za drugim =====
	var dataBytes []byte
	var indexEntries []IndexEntry
	var merkleData [][]byte

	for _, rec := range records {
		serialized := rec.Serialize()
		offset := uint64(len(dataBytes))

		indexEntries = append(indexEntries, IndexEntry{
			Key:        rec.Key,
			DataOffset: offset,
			DataSize:   uint32(len(serialized)),
		})

		dataBytes = append(dataBytes, serialized...)
		merkleData = append(merkleData, serialized)
	}

	if err := mgr.WriteFile(sst.DataPath, dataBytes); err != nil {
		return nil, fmt.Errorf("greska pri pisanju Data fajla: %w", err)
	}

	// ===== 2. FILTER fajl — Bloom Filter svih kljuceva =====
	bf := bloom.NewBloomFilter(bloomExpected, bloomFPRate)
	for _, rec := range records {
		bf.Add([]byte(rec.Key))
	}

	if err := mgr.WriteFile(sst.FilterPath, bf.Serialize()); err != nil {
		return nil, fmt.Errorf("greska pri pisanju Filter fajla: %w", err)
	}

	// ===== 3. INDEX fajl — kljuc + offset + velicina za svaki zapis =====
	indexBytes := serializeIndex(indexEntries)
	if err := mgr.WriteFile(sst.IndexPath, indexBytes); err != nil {
		return nil, fmt.Errorf("greska pri pisanju Index fajla: %w", err)
	}

	// ===== 4. SUMMARY fajl — svaki N-ti kljuc iz Index-a + min/max =====
	summaryBytes := buildSummary(indexEntries, summaryStep)
	if err := mgr.WriteFile(sst.SummaryPath, summaryBytes); err != nil {
		return nil, fmt.Errorf("greska pri pisanju Summary fajla: %w", err)
	}

	// ===== 5. METADATA fajl — Merkle stablo nad podacima =====
	tree := merkle.BuildTree(merkleData)
	if err := mgr.WriteFile(sst.MetadataPath, tree.Serialize()); err != nil {
		return nil, fmt.Errorf("greska pri pisanju Metadata fajla: %w", err)
	}

	return sst, nil
}
