package memtable

import (
	"key_value_engine-nasp/model"
	"sort"
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
func (bt *BTreeMemtable) splitChild(parent *BTreeNode, index int) {
	child := parent.children[index]
	mid := btreeMinDegree - 1
	newNode := &BTreeNode{
		leaf: child.leaf,
		keys: make([]*model.Record, len(child.keys[mid+1:])),
	}
	copy(newNode.keys, child.keys[mid+1:])

	if !child.leaf {
		newNode.children = make([]*BTreeNode, len(child.children[mid+1:]))
		copy(newNode.children, child.children[mid+1:])
	}

	medianKey := child.keys[mid]
	child.keys = child.keys[:mid]
	if !child.leaf {
		child.children = child.children[:mid+1]
	}
	parent.keys = append(parent.keys, nil)
	for i := len(parent.keys) - 1; i > index; i-- {
		parent.keys[i] = parent.keys[i-1]
	}
	parent.keys[index] = medianKey

	parent.children = append(parent.children, nil)
	for i := len(parent.children) - 1; i > index+1; i-- {
		parent.children[i] = parent.children[i-1]
	}
	parent.children[index+1] = newNode
}
func (bt *BTreeMemtable) insertNonFull(node *BTreeNode, record *model.Record) {
	i := len(node.keys) - 1
	if node.leaf {
		node.keys = append(node.keys, nil)
		for i >= 0 && record.Key < node.keys[i].Key {
			node.keys[i+1] = node.keys[i]
			i--
		}
		node.keys[i+1] = record
	} else {
		for i >= 0 && record.Key < node.keys[i].Key {
			i--
		}
		i++
		if i < len(node.children) && len(node.children[i].keys) == 2*btreeMinDegree-1 {
			bt.splitChild(node, i)
			if record.Key > node.keys[i].Key {
				i++
			}
		}
		if i < len(node.children) {
			bt.insertNonFull(node.children[i], record)
		}
	}
}
func (bt *BTreeMemtable) Put(record *model.Record) {
	if existing := bt.search(bt.root, record.Key); existing != nil {
		bt.size -= existing.Size()
		existing.Value = record.Value
		existing.Tombstone = record.Tombstone
		existing.Timestamp = record.Timestamp
		bt.size += existing.Size()
		return
	}
	if len(bt.root.keys) == 2*btreeMinDegree-1 {
		newRoot := &BTreeNode{leaf: false}
		newRoot.children = append(newRoot.children, bt.root)
		bt.splitChild(newRoot, 0)
		bt.root = newRoot
	}
	bt.insertNonFull(bt.root, record)
	bt.size += record.Size()
	bt.count++
}
func (bt *BTreeMemtable) inorderTraversal(node *BTreeNode, records *[]*model.Record) {
	if node == nil {
		return
	}
	for i := 0; i < len(node.keys); i++ {
		if !node.leaf && i < len(node.children) {
			bt.inorderTraversal(node.children[i], records)
		}
		*records = append(*records, node.keys[i])
	}
	if !node.leaf && len(node.children) > len(node.keys) {
		bt.inorderTraversal(node.children[len(node.keys)], records)
	}
}
func (bt *BTreeMemtable) GetAllSorted() []*model.Record {
	var records []*model.Record
	bt.inorderTraversal(bt.root, &records)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Key < records[j].Key
	})
	return records
}
