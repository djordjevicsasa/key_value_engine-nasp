package bloom

import (
	"testing"
)

func TestBloomFilter(t *testing.T) {
	bf := NewBloomFilter(100, 0.01)

	keys := []string{"key1", "key2", "key3"}
	for _, k := range keys {
		bf.Add([]byte(k))
	}

	for _, k := range keys {
		if !bf.Contains([]byte(k)) {
			t.Errorf("Bloom filter should contain key: %s", k)
		}
	}

	if bf.Contains([]byte("not_added")) {
		t.Errorf("Bloom filter reported true for non-existent key")
	}
}

func TestBloomFilterSerialization(t *testing.T) {
	bf := NewBloomFilter(10, 0.01)
	bf.Add([]byte("test_key"))

	data := bf.Serialize()
	bf2 := DeserializeBloomFilter(data)

	if !bf2.Contains([]byte("test_key")) {
		t.Errorf("Deserialized Bloom filter lost data")
	}
}
