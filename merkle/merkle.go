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
	
	// Gradimo stablo odozdo na gore
	currentLevel := leaves
	for len(currentLevel) > 1 {
		var nextLevel []*MerkleNode

		for i := 0; i < len(currentLevel); i += 2 {
			// Eksplicitna alokacija da ne mutiramo postojece hash vrednosti
			combined := make([]byte, 0, len(currentLevel[i].Hash)+len(currentLevel[i+1].Hash))
			combined = append(combined, currentLevel[i].Hash...)
			combined = append(combined, currentLevel[i+1].Hash...)
			h := sha256.Sum256(combined)

			parent := &MerkleNode{
				Hash:  h[:],
				Left:  currentLevel[i],
				Right: currentLevel[i+1],
			}
			nextLevel = append(nextLevel, parent)
		}

		// Ako je neparan, dupliramo poslednji
		if len(nextLevel) > 1 && len(nextLevel)%2 != 0 {
			nextLevel = append(nextLevel, nextLevel[len(nextLevel)-1])
		}

		levels = append(levels, nextLevel)
		currentLevel = nextLevel
	}

	return &MerkleTree{
		Root:  currentLevel[0],
		Nodes: levels,
	}
}

// Vraca hash korena stabla kao hex string
func (t *MerkleTree) RootHash() string {
	if t.Root == nil {
		return ""
	}
	return hex.EncodeToString(t.Root.Hash)
}
