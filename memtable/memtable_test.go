package memtable

import (
	"bytes"
	"key_value_engine-nasp/model"
	"testing"
)

func testMemtableImplementation(t *testing.T, tableType string) {
	mt, err := NewMemtable(tableType)
	if err != nil {
		t.Fatalf("Failed to create %s: %v", tableType, err)
	}

	mt.Put(model.NewRecord("key1", []byte("val1"), false))
	mt.Put(model.NewRecord("key3", []byte("val3"), false))
	mt.Put(model.NewRecord("key2", []byte("val2"), false))

	if mt.Count() != 3 {
		t.Errorf("[%s] Expected count 3, got %d", tableType, mt.Count())
	}

	r := mt.Get("key2")
	if r == nil || !bytes.Equal(r.Value, []byte("val2")) {
		t.Errorf("[%s] Get failed for key2", tableType)
	}

	r = mt.Get("missing")
	if r != nil {
		t.Errorf("[%s] Get should return nil for missing key", tableType)
	}

	sorted := mt.GetAllSorted()
	if len(sorted) != 3 {
		t.Fatalf("[%s] Sorted length %d", tableType, len(sorted))
	}
	if sorted[0].Key != "key1" || sorted[1].Key != "key2" || sorted[2].Key != "key3" {
		t.Errorf("[%s] Sorting failed", tableType)
	}
}

func TestMemtables(t *testing.T) {
	testMemtableImplementation(t, "hashmap")
	testMemtableImplementation(t, "skiplist")
	testMemtableImplementation(t, "btree")
}
