package block

import (
	"fmt"
	"os"
	"path/filepath"
)

type Manager struct {
	blockSize int
}

func NewManager(blockSizeBytes int) *Manager {
	return &Manager{
		blockSize: blockSizeBytes,
	}
}

func (m *Manager) BlockSize() int {
	return m.blockSize
}

func (m *Manager) Read(filePath string, blockIndex int) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ne moze se otvoriti fajl %s: %w", filePath, err)
	}
	defer f.Close()

	offset := int64(blockIndex) * int64(m.blockSize)
	buf := make([]byte, m.blockSize)

	n, err := f.ReadAt(buf, offset)
	if err != nil && n == 0 {
		return nil, fmt.Errorf("greska pri citanju bloka %d iz %s: %w", blockIndex, filePath, err)
	}

	return buf[:n], nil
}

func (m *Manager) Write(filePath string, blockIndex int, data []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("ne moze se kreirati direktorijum %s: %w", dir, err)
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("ne moze se otvoriti fajl %s za pisanje: %w", filePath, err)
	}
	defer f.Close()

	block := make([]byte, m.blockSize)
	copy(block, data)

	offset := int64(blockIndex) * int64(m.blockSize)
	_, err = f.WriteAt(block, offset)
	if err != nil {
		return fmt.Errorf("greska pri pisanju bloka %d u %s: %w", blockIndex, filePath, err)
	}

	return nil
}
