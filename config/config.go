package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	WalDir         string `json:"wal_dir"`
	WalSegmentSize int    `json:"wal_segment_size"`
	WalBlockSize   int    `json:"wal_block_size"`

	MemtableType      string `json:"memtable_type"`
	MemtableMaxSizeKB int    `json:"memtable_max_size_kb"`
	MemtablePoolSize  int    `json:"memtable_pool_size"`

	SSTableDir  string `json:"sstable_dir"`
	SummaryStep int    `json:"summary_step"`

	BlockSizeKB        int `json:"block_size_kb"`
	BlockCacheCapacity int `json:"block_cache_capacity"`

	GetCacheCapacity int `json:"get_cache_capacity"`

	BloomExpectedItems     int     `json:"bloom_expected_items"`
	BloomFalsePositiveRate float64 `json:"bloom_false_positive_rate"`

	TokenBucketMaxTokens        int `json:"token_bucket_max_tokens"`
	TokenBucketRefillIntervalMs int `json:"token_bucket_refill_interval_ms"`
}

func DefaultConfig() *Config {
	return &Config{
		WalDir:                      "data/wal",
		WalSegmentSize:              5,
		WalBlockSize:                4096,
		MemtableType:                "skiplist",
		MemtableMaxSizeKB:           64,
		MemtablePoolSize:            3,
		SSTableDir:                  "data/sstable",
		SummaryStep:                 5,
		BlockSizeKB:                 4,
		BlockCacheCapacity:          50,
		GetCacheCapacity:            100,
		BloomExpectedItems:          1000,
		BloomFalsePositiveRate:      0.01,
		TokenBucketMaxTokens:        100,
		TokenBucketRefillIntervalMs: 1000,
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {

		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) BlockSize() int {
	return c.BlockSizeKB * 1024
}

func (c *Config) MemtableMaxSize() int {
	return c.MemtableMaxSizeKB * 1024
}
