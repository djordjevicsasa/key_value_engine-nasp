package memtable

import (
	"key_value_engine-nasp/model"
	"testing"
)

func TestMemtablePool(t *testing.T) {
	pool, err := NewMemtablePool("skiplist", 25, 2)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	r1 := model.NewRecord("k1", []byte("v1"), false)
	r2 := model.NewRecord("k2", []byte("v2"), false)

	pool.Put(r1)

	if pool.NeedsFlush() {
		t.Errorf("Should not need flush yet")
	}

	pool.Put(r2)

	if !pool.NeedsFlush() {
		t.Errorf("Should need flush after all configured memtables are full")
	}

	r := pool.Get("k1")
	if r == nil {
		t.Errorf("Get failed")
	}

	sorted := pool.GetAllSorted()
	if len(sorted) != 2 {
		t.Errorf("Expected 2 records in pool, got %d", len(sorted))
	}
}
