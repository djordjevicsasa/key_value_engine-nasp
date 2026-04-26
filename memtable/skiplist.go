package memtable

import (
	"key_value_engine-nasp/model"
)

const (
	skipListMaxLevel = 16
	skipListP        = 0.25
)

type SkipListNode struct {
	record  *model.Record
	forward []*SkipListNode
}
type SkipListMemtable struct {
	header *SkipListNode
	level  int
	size   int
	count  int
}

func NewSkipListMemtable() *SkipListMemtable {
	header := &SkipListNode{
		forward: make([]*SkipListNode, skipListMaxLevel),
	}
	return &SkipListMemtable{
		header: header,
		level:  0,
	}
}
