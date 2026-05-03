package memtable

import (
	"fmt"
	"key_value_engine-nasp/model"
)

// Memtable interfejs — sve implementacije (hashmap, skiplist, btree) moraju ga zadovoljiti
type Memtable interface {
	// Dodaje ili azurira zapis
	Put(record *model.Record)

	// Trazi zapis po kljucu. Vraca nil ako ne postoji.
	Get(key string) *model.Record

	// Vraca sve zapise sortirane po kljucu
	GetAllSorted() []*model.Record

	// Vraca trenutnu velicinu u bajtovima (aproksimativno)
	Size() int

	// Proverava da li je memtable dostigla maksimalnu velicinu
	IsFull(maxSize int) bool

	// Vraca broj zapisa
	Count() int
}

// Kreira novu Memtable instancu na osnovu tipa iz konfiguracije
func NewMemtable(memtableType string) (Memtable, error) {
	switch memtableType {
	case "hashmap":
		return NewHashMapMemtable(), nil
	case "skiplist":
		return NewSkipListMemtable(), nil
	case "btree":
		return NewBTreeMemtable(), nil
	default:
		return nil, fmt.Errorf("nepoznat tip memtable: %s (dozvoljeni: hashmap, skiplist, btree)", memtableType)
	}
}
