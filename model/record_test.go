package model

import (
	"bytes"
	"testing"
)

func TestRecordSerialization(t *testing.T) {
	key := "test_key"
	value := []byte("test_value")

	rec := NewRecord(key, value, false)
	serialized := rec.Serialize()

	if rec.Size() != len(serialized) {
		t.Fatalf("Size mismatch: expected %d, got %d", rec.Size(), len(serialized))
	}

	size, err := GetRecordSize(serialized)
	if err != nil {
		t.Fatalf("Failed to get size: %v", err)
	}

	if size != len(serialized) {
		t.Errorf("Expected consumed size %d, got %d", len(serialized), size)
	}

	deserialized, err := DeserializeRecord(serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if deserialized.Key != key {
		t.Errorf("Expected key %s, got %s", key, deserialized.Key)
	}

	if !bytes.Equal(deserialized.Value, value) {
		t.Errorf("Expected value %s, got %s", string(value), string(deserialized.Value))
	}

	if deserialized.Tombstone != false {
		t.Errorf("Expected Tombstone to be false")
	}

	if deserialized.Timestamp != rec.Timestamp {
		t.Errorf("Expected timestamp %d, got %d", rec.Timestamp, deserialized.Timestamp)
	}
}
