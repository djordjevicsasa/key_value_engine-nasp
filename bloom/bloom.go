package bloom

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"math"
)

// Bloom Filter — probabilisticka struktura za brzu proveru pripadnosti
type BloomFilter struct {
	bitset    []bool
	size      uint // velicina bit niza
	hashCount uint // broj hash funkcija
}

// Kreira novi Bloom Filter na osnovu ocekivanog broja elemenata i zeljene stope laznih pozitivnih
func NewBloomFilter(expectedItems int, falsePositiveRate float64) *BloomFilter {
	// Optimalna velicina bit niza: m = -(n * ln(p)) / (ln(2))^2
	m := uint(math.Ceil(-float64(expectedItems) * math.Log(falsePositiveRate) / (math.Log(2) * math.Log(2))))
	if m == 0 {
		m = 1
	}

	// Optimalan broj hash funkcija: k = (m/n) * ln(2)
	k := uint(math.Ceil(float64(m) / float64(expectedItems) * math.Log(2)))
	if k == 0 {
		k = 1
	}

	return &BloomFilter{
		bitset:    make([]bool, m),
		size:      m,
		hashCount: k,
	}
}

// Dodaje element u filter
func (bf *BloomFilter) Add(key []byte) {
	for _, pos := range bf.getPositions(key) {
		bf.bitset[pos] = true
	}
}

// Proverava da li element mozda postoji u filteru
// Vraca false = sigurno ne postoji, true = mozda postoji
func (bf *BloomFilter) Contains(key []byte) bool {
	for _, pos := range bf.getPositions(key) {
		if !bf.bitset[pos] {
			return false
		}
	}
	return true
}

// Racuna hash pozicije za dati kljuc koristeci double hashing tehniku
func (bf *BloomFilter) getPositions(key []byte) []uint {
	positions := make([]uint, bf.hashCount)

	h1 := hashWith(key, 0)
	h2 := hashWith(key, h1)

	for i := uint(0); i < bf.hashCount; i++ {
		positions[i] = (uint(h1) + i*uint(h2)) % bf.size
	}
	return positions
}

// Hash funkcija sa seed vrednoscu
func hashWith(data []byte, seed uint32) uint32 {
	var h hash.Hash32 = fnv.New32a()
	seedBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(seedBytes, seed)
	h.Write(seedBytes)
	h.Write(data)
	return h.Sum32()
}

// Serijalizuje Bloom Filter u niz bajtova za cuvanje na disk
// Format: [Size(4)] [HashCount(4)] [Bitset...]
func (bf *BloomFilter) Serialize() []byte {
	// 4 bajta za size + 4 bajta za hashCount + po 1 bajt za svaki bit
	buf := make([]byte, 8+len(bf.bitset))

	binary.LittleEndian.PutUint32(buf[0:4], uint32(bf.size))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(bf.hashCount))

	for i, bit := range bf.bitset {
		if bit {
			buf[8+i] = 1
		}
	}

	return buf
}

// Deserijalizuje Bloom Filter iz niza bajtova
func DeserializeBloomFilter(data []byte) *BloomFilter {
	if len(data) < 8 {
		return nil
	}

	size := binary.LittleEndian.Uint32(data[0:4])
	hashCount := binary.LittleEndian.Uint32(data[4:8])

	bitset := make([]bool, size)
	for i := uint32(0); i < size && int(8+i) < len(data); i++ {
		bitset[i] = data[8+i] == 1
	}

	return &BloomFilter{
		bitset:    bitset,
		size:      uint(size),
		hashCount: uint(hashCount),
	}
}
