package engine

import (
	"fmt"
	"time"

	"key_value_engine-nasp/block"
	"key_value_engine-nasp/cache"
	"key_value_engine-nasp/config"
	"key_value_engine-nasp/memtable"
	"key_value_engine-nasp/model"
	"key_value_engine-nasp/ratelimit"
	"key_value_engine-nasp/sstable"
	"key_value_engine-nasp/wal"
)

type Engine struct {
	cfg         *config.Config
	wal         *wal.WAL
	pool        *memtable.MemtablePool
	getCache    *cache.LRUCache
	rateLimiter *ratelimit.TokenBucket
	blockMgr    *block.CachedManager
	sstables    []*sstable.SSTable
	sstCounter  int
}

func NewEngine(cfg *config.Config) (*Engine, error) {

	mgr := block.NewManager(cfg.BlockSize())
	cachedMgr := block.NewCachedManager(mgr, cfg.BlockCacheCapacity)

	w, err := wal.NewWAL(cfg.WalDir, cfg.WalBlockSize, cfg.WalSegmentSize, mgr)
	if err != nil {
		return nil, fmt.Errorf("greska pri inicijalizaciji WAL-a: %w", err)
	}

	pool, err := memtable.NewMemtablePool(cfg.MemtableType, cfg.MemtableMaxSize(), cfg.MemtablePoolSize)
	if err != nil {
		return nil, fmt.Errorf("greska pri inicijalizaciji Memtable pool-a: %w", err)
	}

	getCache := cache.NewLRUCache(cfg.GetCacheCapacity)

	rl := ratelimit.NewTokenBucket(cfg.TokenBucketMaxTokens, cfg.TokenBucketRefillIntervalMs)

	e := &Engine{
		cfg:         cfg,
		wal:         w,
		pool:        pool,
		getCache:    getCache,
		rateLimiter: rl,
		blockMgr:    cachedMgr,
	}

	if err := e.loadExistingSSTables(); err != nil {
		return nil, fmt.Errorf("greska pri ucitavanju SSTable-a: %w", err)
	}

	records, err := w.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("greska pri WAL recovery-ju: %w", err)
	}
	if len(records) > 0 {
		if err := pool.LoadFromRecords(records); err != nil {
			return nil, fmt.Errorf("greska pri ucitavanju WAL zapisa: %w", err)
		}
		for pool.NeedsFlush() {
			if err := e.flush(); err != nil {
				return nil, fmt.Errorf("greska pri flush-u posle recovery-ja: %w", err)
			}
		}

	}

	return e, nil
}

func (e *Engine) Put(key string, value []byte) error {
	if !e.rateLimiter.Allow() {
		return model.ErrRateLimited
	}

	record := model.NewRecord(key, value, false)

	if err := e.wal.Append(record); err != nil {
		return fmt.Errorf("greska pri upisu u WAL: %w", err)
	}

	if err := e.pool.Put(record); err != nil {
		return fmt.Errorf("greska pri upisu u Memtable: %w", err)
	}

	e.getCache.Delete(key)

	if e.pool.NeedsFlush() {
		if err := e.flush(); err != nil {
			return fmt.Errorf("greska pri flush-u: %w", err)
		}
	}

	return nil
}

func (e *Engine) Get(key string) ([]byte, error) {
	if !e.rateLimiter.Allow() {
		return nil, model.ErrRateLimited
	}

	rec, err := e.getRecord(key)
	if err != nil {
		return nil, err
	}
	if rec.Tombstone {
		return nil, model.ErrDeleted
	}
	return rec.Value, nil
}

func (e *Engine) getRecord(key string) (*model.Record, error) {
	if val, found := e.getCache.Get(key); found {
		rec := val.(*model.Record)
		return rec, nil
	}

	if rec := e.pool.Get(key); rec != nil {
		e.getCache.Put(key, rec)
		return rec, nil
	}

	for _, sst := range e.sstables {
		rec, err := sstable.SearchSSTable(sst, key, e.blockMgr)
		if err != nil {
			continue
		}
		if rec != nil {
			e.getCache.Put(key, rec)
			return rec, nil
		}
	}

	return nil, model.ErrKeyNotFound
}

func (e *Engine) Delete(key string) error {
	if !e.rateLimiter.Allow() {
		return model.ErrRateLimited
	}

	rec, err := e.getRecord(key)
	if err == model.ErrKeyNotFound {
		return model.ErrKeyNotFound
	}
	if err != nil {
		return err
	}
	if rec.Tombstone {
		return model.ErrKeyNotFound
	}

	record := model.NewRecord(key, nil, true)

	if err := e.wal.Append(record); err != nil {
		return fmt.Errorf("greska pri upisu tombstone u WAL: %w", err)
	}

	if err := e.pool.Put(record); err != nil {
		return fmt.Errorf("greska pri upisu tombstone u Memtable: %w", err)
	}

	e.getCache.Delete(key)

	if e.pool.NeedsFlush() {
		if err := e.flush(); err != nil {
			return fmt.Errorf("greska pri flush-u: %w", err)
		}
	}

	return nil
}

func (e *Engine) flush() error {
	oldest, err := e.pool.GetOldestForFlush()
	if err != nil {
		return err
	}
	if oldest == nil {
		return nil
	}

	records := oldest.GetAllSorted()
	if len(records) == 0 {
		return nil
	}

	e.sstCounter++
	id := fmt.Sprintf("sst_%d_%d", time.Now().UnixMilli(), e.sstCounter)

	sst, err := sstable.WriteSSTable(
		e.cfg.SSTableDir,
		id,
		records,
		e.cfg.SummaryStep,
		e.cfg.BloomExpectedItems,
		e.cfg.BloomFalsePositiveRate,
		e.blockMgr,
	)
	if err != nil {
		return err
	}

	e.sstables = append([]*sstable.SSTable{sst}, e.sstables...)

	if err := e.wal.Clear(); err != nil {
		return fmt.Errorf("greska pri ciscenju WAL-a: %w", err)
	}

	for _, rec := range e.pool.GetUnflushedRecords() {
		if err := e.wal.Append(rec); err != nil {
			return fmt.Errorf("greska pri ponovnom upisu u WAL: %w", err)
		}
	}

	return nil
}
