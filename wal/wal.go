package wal

import (
	"encoding/binary"
	"fmt"
	"os"

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
