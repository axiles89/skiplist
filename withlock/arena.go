package withlock

import (
	"errors"
	"unsafe"
)

type Arena struct {
	buf []byte
	newSize uint32
}

var (
	ErrArenaFull = errors.New("allocation failed because arena is full")
)

func NewArena(size uint32) *Arena {
	a := Arena{
		buf: make([]byte, size),
	}
	return &a
}

func (a *Arena) alloc(size uint32) (uint32, error) {
	if a.newSize + size > uint32(len(a.buf)) {
		return 0, ErrArenaFull
	}
	a.newSize = a.newSize + size
	return a.newSize - size, nil
}

func (a *Arena) getPointerByOffset(offset uint32) unsafe.Pointer {
	return unsafe.Pointer(&a.buf[offset])
}

func (a *Arena) getBytes(offset, size uint32) []byte {
	return a.buf[offset:offset + size]
}
