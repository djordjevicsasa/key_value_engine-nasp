package sstable

import (
	"encoding/binary"

	"key_value_engine-nasp/block"
	"key_value_engine-nasp/bloom"
	"key_value_engine-nasp/merkle"
	"key_value_engine-nasp/model"
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

// Ucitava Summary strukturu iz fajla
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

// Pretrazuje Index fajl u zadatom opsegu i trazi trazeni kljuc
// Vraca offset i velicinu zapisa u Data fajlu
func SearchIndex(indexPath string, startOffset, endOffset uint64, key string, mgr *block.CachedManager) (uint64, uint32, bool, error) {
	data, err := mgr.ReadFile(indexPath)
	if err != nil {
		return 0, 0, false, err
	}

	start := int(startOffset)
	end := len(data)
	if endOffset > 0 && int(endOffset) < end {
		end = int(endOffset)
	}

	offset := start
	for offset < end {
		if offset+4 > end {
			break
		}

		keySize := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4

		if offset+keySize+12 > len(data) {
			break
		}

		entryKey := string(data[offset : offset+keySize])
		offset += keySize

		dataOffset := binary.LittleEndian.Uint64(data[offset : offset+8])
		offset += 8

		dataSize := binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4

		if entryKey == key {
			return dataOffset, dataSize, true, nil
		}

		// Index je sortiran — ako smo presli trazeni kljuc, prekidamo
		if entryKey > key {
			break
		}
	}

	return 0, 0, false, nil
}

// Cita zapis iz Data fajla na zadatom offset-u sa zadatom velicinom
func ReadDataRecord(dataPath string, offset uint64, size uint32, mgr *block.CachedManager) (*model.Record, error) {
	buf, err := mgr.ReadRaw(dataPath, int64(offset), int(size))
	if err != nil {
		return nil, err
	}

	return model.DeserializeRecord(buf)
}

// Ucitava Merkle stablo iz Metadata fajla
func LoadMetadata(metadataPath string, mgr *block.CachedManager) (*merkle.MerkleTree, error) {
	data, err := mgr.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}
	return merkle.DeserializeTree(data), nil
}

// Verifikuje integritet Data fajla koriscenjem Merkle stabla
func VerifySSTable(sst *SSTable, mgr *block.CachedManager) (bool, error) {
	// Ucitavamo Merkle stablo
	tree, err := LoadMetadata(sst.MetadataPath, mgr)
	if err != nil {
		return false, err
	}

	// Ucitavamo sve zapise iz Data fajla
	records, err := ReadAllData(sst.DataPath, mgr)
	if err != nil {
		return false, err
	}

	// Serijalizujemo zapise za poredjenje
	var dataChunks [][]byte
	for _, rec := range records {
		dataChunks = append(dataChunks, rec.Serialize())
	}

	return merkle.Verify(dataChunks, tree), nil
}

// Cita sve zapise iz Data fajla
func ReadAllData(dataPath string, mgr *block.CachedManager) ([]*model.Record, error) {
	data, err := mgr.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	var records []*model.Record
	offset := 0

	for offset < len(data) {
		recordSize, err := model.GetRecordSize(data[offset:])
		if err != nil || offset+recordSize > len(data) {
			break
		}

		rec, err := model.DeserializeRecord(data[offset : offset+recordSize])
		if err != nil {
			break
		}

		records = append(records, rec)
		offset += recordSize
	}

	return records, nil
}

// Pretrazuje jednu SSTable tabelu za zadatim kljucem — kompletni read path
// Redosled: Bloom Filter → Summary → Index → Data
func SearchSSTable(sst *SSTable, key string, mgr *block.CachedManager) (*model.Record, error) {
	// 1. Provera Bloom Filtera — ako kaze da ne postoji, sigurno ne postoji
	bf, err := LoadFilter(sst.FilterPath, mgr)
	if err != nil {
		return nil, err
	}
	if !bf.Contains([]byte(key)) {
		return nil, nil // kljuc sigurno nije u ovoj tabeli
	}

	// 2. Ucitamo Summary i proverimo da li je kljuc u opsegu
	minKey, maxKey, summaryEntries, err := LoadSummary(sst.SummaryPath, mgr)
	if err != nil {
		return nil, err
	}
	if key < minKey || key > maxKey {
		return nil, nil // kljuc je van opsega ove tabele
	}

	// 3. Pretrazujemo Summary da nadjemo opseg u Index-u
	startOffset, endOffset := SearchSummary(summaryEntries, key)

	// 4. Pretrazujemo Index u tom opsegu
	dataOffset, dataSize, found, err := SearchIndex(sst.IndexPath, startOffset, endOffset, key, mgr)
	if err != nil || !found {
		return nil, err
	}

	// 5. Citamo zapis iz Data fajla
	return ReadDataRecord(sst.DataPath, dataOffset, dataSize, mgr)
}
