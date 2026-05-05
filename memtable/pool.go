package memtable

import (
	"fmt"
	"key_value_engine-nasp/model"
)

type MemtablePool struct {
	active    Memtable
	readOnly  []Memtable
	maxSize   int
	poolSize  int
	tableType string
}

func NewMemtablePool(tableType string, maxSizeBytes int, poolSize int) (*MemtablePool, error) {
	if poolSize < 1 {
		return nil, fmt.Errorf("memtable pool mora imati bar 1 instancu")
	}

	active, err := NewMemtable(tableType)
	if err != nil {
		return nil, err
	}
	return &MemtablePool{
		active:    active,
		readOnly:  make([]Memtable, 0),
		maxSize:   maxSizeBytes,
		poolSize:  poolSize,
		tableType: tableType,
	}, nil
}

func sortRecords(records []*model.Record) {
	for i := 1; i < len(records); i++ {
		for j := i; j > 0 && records[j].Key < records[j-1].Key; j-- {
			records[j], records[j-1] = records[j-1], records[j]
		}
	}
}

func (p *MemtablePool) LoadFromRecords(records []*model.Record) error {
	for _, rec := range records {
		if err := p.Put(rec); err != nil {
			return err
		}
	}
	return nil
}

func (p *MemtablePool) ClearAll() error {
	if err := p.createActive(); err != nil {
		return err
	}
	p.readOnly = make([]Memtable, 0)
	return nil
}

func (p *MemtablePool) createActive() error {
	newActive, err := NewMemtable(p.tableType)
	if err != nil {
		return err
	}
	p.active = newActive
	return nil
}

func (p *MemtablePool) handleFullActive() error {
	p.readOnly = append(p.readOnly, p.active)

	if len(p.readOnly) < p.poolSize {
		return p.createActive()
	}

	p.active = nil
	return nil
}

func (p *MemtablePool) Put(record *model.Record) error {
	if p.active == nil {
		if err := p.createActive(); err != nil {
			return err
		}
	}
	p.active.Put(record)
	if p.active.IsFull(p.maxSize) {
		if err := p.handleFullActive(); err != nil {
			return err
		}
	}
	return nil
}

func (p *MemtablePool) Get(key string) *model.Record {
	if p.active != nil {
		if rec := p.active.Get(key); rec != nil {
			return rec
		}
	}
	for i := len(p.readOnly) - 1; i >= 0; i-- {
		if rec := p.readOnly[i].Get(key); rec != nil {
			return rec
		}
	}
	return nil
}
