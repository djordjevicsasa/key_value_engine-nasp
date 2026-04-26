package memtable

import (
	"key_value_engine-nasp/model"
)

const btreeMinDegree = 4

type BTreeNode struct {
	keys     []*model.Record
	children []*BTreeNode
	leaf     bool
}

type BTreeMemtable struct {
	root  *BTreeNode
	size  int
	count int
}

func NewBTreeMemtable() *BTreeMemtable {
	return &BTreeMemtable{
		root: &BTreeNode{leaf: true},
	}
}
