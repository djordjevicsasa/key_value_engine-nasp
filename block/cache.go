package block

import (
	"key_value_engine-nasp/cache"
)

type CachedManager struct {
	manager *Manager
	cache   *cache.LRUCache
}

func NewCachedManager(manager *Manager, cacheCapacity int) *CachedManager {
	return &CachedManager{
		manager: manager,
		cache:   cache.NewLRUCache(cacheCapacity),
	}
}

func (cm *CachedManager) Manager() *Manager {
	return cm.manager
}

func (cm *CachedManager) Read(filePath string, blockIndex int) ([]byte, error) {
	key := cacheKey(filePath, blockIndex)

	if val, found := cm.cache.Get(key); found {
		return val.([]byte), nil
	}

	data, err := cm.manager.Read(filePath, blockIndex)
	if err != nil {
		return nil, err
	}

	cm.cache.Put(key, data)
	return data, nil
}

func (cm *CachedManager) Write(filePath string, blockIndex int, data []byte) error {
	err := cm.manager.Write(filePath, blockIndex, data)
	if err != nil {
		return err
	}

	key := cacheKey(filePath, blockIndex)
	block := make([]byte, len(data))
	copy(block, data)
	cm.cache.Put(key, block)
	return nil
}

func (cm *CachedManager) ClearCache() {
	cm.cache.Clear()
}
