package wal

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"key_value_engine-nasp/block"
	"key_value_engine-nasp/model"
)

const (
	fragmentFull   byte = 0
	fragmentFirst  byte = 1
	fragmentMiddle byte = 2
	fragmentLast   byte = 3
)

const fragmentHeaderSize = 5

type WAL struct {
	dir            string
	blockSize      int
	maxBlocks      int
	currentSegment int
	currentBlock   int
	blockOffset    int
	manager        *block.Manager
}

func NewWAL(dir string, blockSize int, maxBlocks int, mgr *block.Manager) (*WAL, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("ne moze se kreirati WAL direktorijum: %w", err)
	}

	w := &WAL{
		dir:       dir,
		blockSize: blockSize,
		maxBlocks: maxBlocks,
		manager:   mgr,
	}

	segments, err := w.listSegments()
	if err != nil {
		return nil, err
	}

	if len(segments) > 0 {
		w.currentSegment = segments[len(segments)-1]

		w.findWritePosition()
	}

	return w, nil
}

func (w *WAL) Append(record *model.Record) error {
	data := record.Serialize()

	if w.currentBlock >= w.maxBlocks {
		w.currentSegment++
		w.currentBlock = 0
		w.blockOffset = 0
	}

	remaining := data
	first := true

	for len(remaining) > 0 {
		available := w.blockSize - w.blockOffset - fragmentHeaderSize
		if available <= 0 {

			w.currentBlock++
			w.blockOffset = 0
			available = w.blockSize - fragmentHeaderSize
		}

		writeSize := available
		if writeSize > len(remaining) {
			writeSize = len(remaining)
		}

		chunk := remaining[:writeSize]
		remaining = remaining[writeSize:]

		var fragType byte
		if first && len(remaining) == 0 {
			fragType = fragmentFull
		} else if first {
			fragType = fragmentFirst
		} else if len(remaining) == 0 {
			fragType = fragmentLast
		} else {
			fragType = fragmentMiddle
		}

		header := make([]byte, fragmentHeaderSize)
		header[0] = fragType
		binary.LittleEndian.PutUint32(header[1:], uint32(writeSize))

		segPath := w.segmentPath(w.currentSegment)
		offset := int64(w.currentBlock)*int64(w.blockSize) + int64(w.blockOffset)

		writeData := append(header, chunk...)
		if err := w.manager.WriteRaw(segPath, offset, writeData); err != nil {
			return fmt.Errorf("greska pri upisu fragmenta u WAL: %w", err)
		}

		w.blockOffset += fragmentHeaderSize + writeSize
		first = false

		if w.blockOffset >= w.blockSize {
			w.currentBlock++
			w.blockOffset = 0
		}
	}

	return nil
}

func (w *WAL) ReadAll() ([]*model.Record, error) {
	segments, err := w.listSegments()
	if err != nil {
		return nil, err
	}

	var records []*model.Record

	for _, segIndex := range segments {
		segRecords, err := w.readSegment(segIndex)
		if err != nil {
			continue
		}
		records = append(records, segRecords...)
	}

	return records, nil
}

func (w *WAL) readSegment(segIndex int) ([]*model.Record, error) {
	segPath := w.segmentPath(segIndex)

	fileData, err := w.manager.ReadFile(segPath)
	if err != nil {
		return nil, err
	}

	var records []*model.Record
	var assemblyBuf []byte

	totalBlocks := (len(fileData) + w.blockSize - 1) / w.blockSize

	for blockIdx := 0; blockIdx < totalBlocks; blockIdx++ {
		blockStart := blockIdx * w.blockSize
		blockEnd := blockStart + w.blockSize
		if blockEnd > len(fileData) {
			blockEnd = len(fileData)
		}
		blockData := fileData[blockStart:blockEnd]

		offset := 0
		for offset+fragmentHeaderSize <= len(blockData) {
			fragType := blockData[offset]
			dataSize := int(binary.LittleEndian.Uint32(blockData[offset+1 : offset+5]))

			if dataSize == 0 {
				break
			}

			if offset+fragmentHeaderSize+dataSize > len(blockData) {
				break
			}

			chunk := blockData[offset+fragmentHeaderSize : offset+fragmentHeaderSize+dataSize]
			offset += fragmentHeaderSize + dataSize

			switch fragType {
			case fragmentFull:
				rec, err := model.DeserializeRecord(chunk)
				if err == nil {
					records = append(records, rec)
				}
				assemblyBuf = nil

			case fragmentFirst:
				assemblyBuf = make([]byte, 0, dataSize*2)
				assemblyBuf = append(assemblyBuf, chunk...)

			case fragmentMiddle:
				if assemblyBuf != nil {
					assemblyBuf = append(assemblyBuf, chunk...)
				}

			case fragmentLast:
				if assemblyBuf != nil {
					assemblyBuf = append(assemblyBuf, chunk...)
					rec, err := model.DeserializeRecord(assemblyBuf)
					if err == nil {
						records = append(records, rec)
					}
					assemblyBuf = nil
				}
			}
		}
	}

	return records, nil
}

func (w *WAL) findWritePosition() {
	segPath := w.segmentPath(w.currentSegment)
	fileSize, err := w.manager.FileSize(segPath)
	if err != nil {
		return
	}

	if fileSize == 0 {
		return
	}

	fSize := int(fileSize)
	w.currentBlock = (fSize - 1) / w.blockSize
	w.blockOffset = 0

	data, err := w.manager.ReadFile(segPath)
	if err != nil {
		return
	}

	blockStart := w.currentBlock * w.blockSize
	if blockStart >= len(data) {
		return
	}

	blockEnd := blockStart + w.blockSize
	if blockEnd > len(data) {
		blockEnd = len(data)
	}
	blockData := data[blockStart:blockEnd]

	offset := 0
	for offset+fragmentHeaderSize <= len(blockData) {
		dataSize := int(binary.LittleEndian.Uint32(blockData[offset+1 : offset+5]))
		if dataSize == 0 {
			break
		}
		offset += fragmentHeaderSize + dataSize
	}
	w.blockOffset = offset
}

func (w *WAL) Clear() error {
	segments, err := w.listSegments()
	if err != nil {
		return err
	}

	for _, seg := range segments {
		path := w.segmentPath(seg)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	w.currentSegment = 0
	w.currentBlock = 0
	w.blockOffset = 0

	return nil
}

func (w *WAL) listSegments() ([]int, error) {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var indices []int
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "wal_") {
			continue
		}
		var idx int
		if _, err := fmt.Sscanf(entry.Name(), "wal_%d.log", &idx); err == nil {
			indices = append(indices, idx)
		}
	}

	sort.Ints(indices)
	return indices, nil
}

func (w *WAL) segmentPath(index int) string {
	return filepath.Join(w.dir, fmt.Sprintf("wal_%05d.log", index))
}
