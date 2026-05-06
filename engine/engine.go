package engine

import (
	"fmt"

	"key_value_engine-nasp/block"
	"key_value_engine-nasp/cache"
	"key_value_engine-nasp/config"
	"key_value_engine-nasp/memtable"
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
