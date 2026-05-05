package merkle

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// MerkleTree — struktura za verifikaciju integriteta podataka u SSTable-u
type MerkleTree struct {
	Root  *MerkleNode
	Nodes [][]*MerkleNode // nivoi stabla, od listova ka korenu
}

type MerkleNode struct {
	Hash  []byte
	Left  *MerkleNode
	Right *MerkleNode
}

// Gradi Merkle stablo od niza podataka (vrednosti zapisa iz SSTable Data)
func BuildTree(data [][]byte) *MerkleTree {
	if len(data) == 0 {
		return &MerkleTree{}
	}

	// Pravimo listove
	leaves := make([]*MerkleNode, len(data))
	for i, d := range data {
		h := sha256.Sum256(d)
		leaves[i] = &MerkleNode{Hash: h[:]}
	}

	// Ako je neparan broj, dupliramo poslednji list
	if len(leaves)%2 != 0 {
		leaves = append(leaves, leaves[len(leaves)-1])
	}

	levels := [][]*MerkleNode{leaves}
