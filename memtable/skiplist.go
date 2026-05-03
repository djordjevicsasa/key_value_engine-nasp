package memtable

import (
	"key_value_engine-nasp/model"
	"math/rand"
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
func (sl *SkipListMemtable) Size() int {
	return sl.size
}

func (sl *SkipListMemtable) IsFull(maxSize int) bool {
	return sl.size >= maxSize
}

func (sl *SkipListMemtable) Count() int {
	return sl.count
}

func (sl *SkipListMemtable) randomLevel() int {
	level := 1
	for level < skipListMaxLevel && rand.Float64() < skipListP {
		level++
	}
	return level
}

func (sl *SkipListMemtable) Put(record *model.Record) {
	update := make([]*SkipListNode, skipListMaxLevel)
	current := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].record.Key < record.Key {
			current = current.forward[i]
		}
		update[i] = current
	}
	if current.forward[0] != nil && current.forward[0].record.Key == record.Key {
		oldSize := current.forward[0].record.Size()
		current.forward[0].record = record
		sl.size += record.Size() - oldSize
		return
	}
	newLevel := sl.randomLevel()
	if newLevel > sl.level {
		for i := sl.level; i < newLevel; i++ {
			update[i] = sl.header
		}
		sl.level = newLevel
	}
	newNode := &SkipListNode{
		record:  record,
		forward: make([]*SkipListNode, newLevel),
	}

	for i := 0; i < newLevel; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	sl.size += record.Size()
	sl.count++
}
func (sl *SkipListMemtable) Get(key string) *model.Record {
	current := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].record.Key < key {
			current = current.forward[i]
		}
	}
	current = current.forward[0]
	if current != nil && current.record.Key == key {
		return current.record
	}
	return nil
}
func (sl *SkipListMemtable) GetAllSorted() []*model.Record {
	var records []*model.Record
	current := sl.header.forward[0]

	for current != nil {
		records = append(records, current.record)
		current = current.forward[0]
	}
	return records
}
