package engine

import (
	"key_value_engine-nasp/config"
	"key_value_engine-nasp/model"
	"os"
	"testing"
)

func TestEngineLifecycle(t *testing.T) {
	// Setup test config
	cfg := config.DefaultConfig()
	cfg.WalDir = "test_data/wal"
	cfg.SSTableDir = "test_data/sstable"

	defer os.RemoveAll("test_data")

	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}

	err = eng.Put("key1", []byte("val1"))
	if err != nil {
		t.Errorf("Put failed: %v", err)
	}

	val, err := eng.Get("key1")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if string(val) != "val1" {
		t.Errorf("Expected val1, got %s", string(val))
	}

	err = eng.Delete("key1")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	_, err = eng.Get("key1")
	if err != model.ErrDeleted && err != model.ErrKeyNotFound {
		t.Errorf("Expected ErrDeleted or ErrKeyNotFound, got %v", err)
	}

	err = eng.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	eng2, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("NewEngine recovery failed: %v", err)
	}

	_, err = eng2.Get("key1")
	if err != model.ErrDeleted && err != model.ErrKeyNotFound {
		t.Errorf("Expected ErrDeleted or ErrKeyNotFound, got %v", err)
	}

	err = eng2.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
