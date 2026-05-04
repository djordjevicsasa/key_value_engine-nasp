package wal

import (
	"key_value_engine-nasp/block"
	"key_value_engine-nasp/model"
	"os"
	"testing"
)

func TestWALFragmentation(t *testing.T) {
	dir := "test_wal_dir"
	defer os.RemoveAll(dir)

	mgr := block.NewManager(100)
	w, err := NewWAL(dir, 100, 5, mgr)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}

	largeVal := make([]byte, 200)
	for i := range largeVal {
		largeVal[i] = 'A'
	}

	rec1 := model.NewRecord("large_key", largeVal, false)

	if err := w.Append(rec1); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	rec2 := model.NewRecord("small", []byte("val"), false)
	if err := w.Append(rec2); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	records, err := w.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}

	if records[0].Key != "large_key" {
		t.Errorf("Key mismatch, got %s", records[0].Key)
	}

	if records[1].Key != "small" {
		t.Errorf("Key mismatch, got %s", records[1].Key)
	}
}
