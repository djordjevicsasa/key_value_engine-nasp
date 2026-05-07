package sstable

import (
	"key_value_engine-nasp/block"
	"key_value_engine-nasp/model"
	"os"
	"testing"
)

func TestSSTableWriteAndRead(t *testing.T) {
	dir := "test_sstable_dir"
	defer os.RemoveAll(dir)
	os.MkdirAll(dir, 0755)

	mgr := block.NewCachedManager(block.NewManager(4096), 10)

	records := []*model.Record{
		model.NewRecord("key1", []byte("val1"), false),
		model.NewRecord("key2", []byte("val2"), false),
		model.NewRecord("key3", []byte("val3"), true),
	}

	sst, err := WriteSSTable(dir, "sst_test_1", records, 2, 100, 0.01, mgr)
	if err != nil {
		t.Fatalf("WriteSSTable failed: %v", err)
	}

	// Verify Data Integrity using Merkle Tree
	valid, err := VerifySSTable(sst, mgr)
	if err != nil || !valid {
		t.Errorf("VerifySSTable failed or returned invalid")
	}

	// Search for key2
	rec, err := SearchSSTable(sst, "key2", mgr)
	if err != nil {
		t.Fatalf("SearchSSTable failed: %v", err)
	}
	if rec == nil || string(rec.Value) != "val2" {
		t.Errorf("Failed to retrieve key2")
	}

	// Read all data
	all, err := ReadAllData(sst.DataPath, mgr)
	if err != nil {
		t.Fatalf("ReadAllData failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("Expected 3 records, got %d", len(all))
	}
}
