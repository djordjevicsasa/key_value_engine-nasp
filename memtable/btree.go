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
func (bt *BTreeMemtable) Size() int {
	return bt.size
}

func (bt *BTreeMemtable) IsFull(maxSize int) bool {
	return bt.size >= maxSize
}

func (bt *BTreeMemtable) Count() int {
	return bt.count
}
func (bt *BTreeMemtable) search(node *BTreeNode, key string) *model.Record {
	if node == nil {
		return nil
	}

	i := 0
	for i < len(node.keys) && key > node.keys[i].Key {
		i++
	}

	if i < len(node.keys) && key == node.keys[i].Key {
		return node.keys[i]
	}

	if node.leaf {
		return nil
	}

	if i < len(node.children) {
		return bt.search(node.children[i], key)
	}
	return nil
}
func (bt *BTreeMemtable) Get(key string) *model.Record {
	return bt.search(bt.root, key)
}
