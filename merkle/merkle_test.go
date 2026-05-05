package merkle

import (
	"testing"
)

func TestMerkleTree(t *testing.T) {
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
	}

	tree := BuildTree(data)

	if tree.Root == nil {
		t.Fatal("Root should not be nil")
	}

	hash := tree.RootHash()
	if hash == "" {
		t.Errorf("Root hash should not be empty")
	}

	serialized := tree.Serialize()
	tree2 := DeserializeTree(serialized)

	if tree2.RootHash() != hash {
		t.Errorf("Deserialized tree hash mismatch. Expected %s, got %s", hash, tree2.RootHash())
	}

	if !Verify(data, tree) {
		t.Errorf("Tree verification failed with original data")
	}

	corruptedData := [][]byte{
		[]byte("data1"),
		[]byte("corrupted"),
		[]byte("data3"),
	}

	if Verify(corruptedData, tree) {
		t.Errorf("Tree verification should fail with corrupted data")
	}
}
