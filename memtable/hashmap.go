package memtable

import (
	"key_value_engine-nasp/model"
	"sort"
)

type HashMapMemtable struct {
	data map[string]*model.Record
	size int
}

func NewHashMapMemtable() *HashMapMemtable {
	return &HashMapMemtable{
		data: make(map[string]*model.Record),
	}
}

func (m *HashMapMemtable) Put(record *model.Record) {
	if old, exists := m.data[record.Key]; exists {
		m.size -= old.Size()
	}
	m.data[record.Key] = record
	m.size += record.Size()
}

func (m *HashMapMemtable) Get(key string) *model.Record {
	rec, exists := m.data[key]
	if !exists {
		return nil
	}
	return rec
}

func (m *HashMapMemtable) GetAllSorted() []*model.Record {
	records := make([]*model.Record, 0, len(m.data))
	for _, rec := range m.data {
		records = append(records, rec)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].Key < records[j].Key
	})
	return records
}

func (m *HashMapMemtable) Size() int {
	return m.size
}

func (m *HashMapMemtable) IsFull(maxSize int) bool {
	return m.size >= maxSize
}

func (m *HashMapMemtable) Count() int {
	return len(m.data)
}
