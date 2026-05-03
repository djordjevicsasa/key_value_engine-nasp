package block

import (
	"fmt"
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

func (cm *CachedManager) ReadFile(filePath string) ([]byte, error) {
	fileSize, err := cm.manager.FileSize(filePath)
	if err != nil {
		return nil, err
	}
	if fileSize == 0 {
		return []byte{}, nil
	}

	bs := cm.manager.BlockSize()
	blockCount := (int(fileSize) + bs - 1) / bs
	result := make([]byte, 0, fileSize)

	for i := 0; i < blockCount; i++ {
		blockData, err := cm.Read(filePath, i)
		if err != nil {
			return nil, err
		}
		result = append(result, blockData...)
	}

	return result[:fileSize], nil
}

func (cm *CachedManager) ReadRaw(filePath string, offset int64, size int) ([]byte, error) {
	return cm.manager.ReadRaw(filePath, offset, size)
}

func (cm *CachedManager) WriteFile(filePath string, data []byte) error {
	return cm.manager.WriteFile(filePath, data)
}

func (cm *CachedManager) FileSize(filePath string) (int64, error) {
	return cm.manager.FileSize(filePath)
}

func (cm *CachedManager) BlockSize() int {
	return cm.manager.BlockSize()
}

func cacheKey(filePath string, blockIndex int) string {
	return fmt.Sprintf("%s:%d", filePath, blockIndex)
}
