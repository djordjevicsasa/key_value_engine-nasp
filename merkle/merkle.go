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
// Serijalizuje stablo u niz bajtova za cuvanje na disk
// Format: [BrojNivoa(4)] za svaki nivo:[BrojCvorova(4)] [Hash(32)]...
func (t *MerkleTree) Serialize() []byte{
	if  t.Root == nil{

		return make([]byte,4)//samo broj nivova = 0

	}


	var buf []byte

	// Broj nivoa
	levelCount := make([]byte, 4)
	binary.LittleEndian.PutUint32(levelCount, uint32(len(t.Nodes)))
	buf = append(buf, levelCount...)


	// Svaki nivo
	for _, level := range t.Nodes {
		nodeCount := make([]byte, 4)
		binary.LittleEndian.PutUint32(nodeCount, uint32(len(level)))
		buf = append(buf, nodeCount...)

		for _, node := range level {
			buf = append(buf, node.Hash...)
		}
	}

	return buf
}
	// Deserijalizuje Merkle stablo iz bajtova
func DeserializeTree(data []byte) *MerkleTree {
	if len(data) < 4 {
		return &MerkleTree{}
	}

	levelCount := int(binary.LittleEndian.Uint32(data[0:4]))
	if levelCount == 0 {
		return &MerkleTree{}
	}

	offset := 4
	levels := make([][]*MerkleNode, levelCount)
	for l := 0; l < levelCount; l++ {
		if offset+4 > len(data) {
			break
		}
		nodeCount := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4

		nodes := make([]*MerkleNode, nodeCount)
		for n := 0; n < nodeCount; n++ {
			if offset+32 > len(data) {
				break
			}
			hash := make([]byte, 32)
			copy(hash, data[offset:offset+32])
			nodes[n] = &MerkleNode{Hash: hash}
			offset += 32
		}
		levels[l] = nodes
	}

	// Poslednji nivo sadrzi koren
	tree := &MerkleTree{Nodes: levels}
	if len(levels) > 0 && len(levels[len(levels)-1]) > 0 {
		tree.Root = levels[len(levels)-1][0]
	}

	return tree
}

// Verifikuje integritet podataka poredjenjem sa sacuvanim stablom
func Verify(data [][]byte, savedTree *MerkleTree) bool {
	if savedTree == nil || savedTree.Root == nil {
		return len(data) == 0
	}

	newTree := BuildTree(data)
	return newTree.RootHash() == savedTree.RootHash()
}
