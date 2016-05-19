// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mmf

import (
	"bytes"
	"io"
)

// MemoryRegionReader is a reader for safe operations over a shared memory region.
// It holds a reference to the region, so the former can't be gc'ed.
type MemoryRegionReader struct {
	region *MemoryRegion
	*bytes.Reader
}

// NewMemoryRegionReader creates a new reader for the given region.
func NewMemoryRegionReader(region *MemoryRegion) *MemoryRegionReader {
	return &MemoryRegionReader{
		region: region,
		Reader: bytes.NewReader(region.Data()),
	}
}

// MemoryRegionWriter is a writer for safe operations over a shared memory region.
// It holds a reference to the region, so the former can't be gc'ed.
type MemoryRegionWriter struct {
	region *MemoryRegion
	pos    int64
}

// NewMemoryRegionWriter creates a new writer for the given region.
func NewMemoryRegionWriter(region *MemoryRegion) *MemoryRegionWriter {
	return &MemoryRegionWriter{region: region}
}

// WriteAt is to implement io.WriterAt.
func (w *MemoryRegionWriter) WriteAt(p []byte, off int64) (n int, err error) {
	data := w.region.Data()
	n = len(data) - int(off)
	if n > 0 {
		if n > len(p) {
			n = len(p)
		}
		copy(data[off:], p[:n])
	}
	if n < len(p) {
		err = io.EOF
	}
	return
}

// Write is to implement io.Writer.
func (w *MemoryRegionWriter) Write(p []byte) (n int, err error) {
	n, err = w.WriteAt(p, w.pos)
	w.pos += int64(n)
	return n, err
}
