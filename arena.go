package skiplist

import (
	"errors"
	"sync/atomic"
	"unsafe"
)

type Arena struct {
	buf []byte
	newSize uint32
}

var (
	ErrArenaFull = errors.New("allocation failed because arena is full")
	align uint32 = 3
)

func NewArena(size uint32) *Arena {
	a := Arena{
		buf: make([]byte, size),
	}
	return &a
}


func (a *Arena) alloc(size, unusedSize uint32) (uint32, error) {
	if atomic.LoadUint32(&a.newSize) > uint32(len(a.buf)) {
		return 0, ErrArenaFull
	}

	sizeWithAllign := size + align

	newSize := atomic.AddUint32(&a.newSize, sizeWithAllign)
	if newSize + unusedSize > uint32(len(a.buf)) {
		return 0, ErrArenaFull
	}

	offset := (newSize - sizeWithAllign + align) & ^align
	return offset, nil
}

func (a *Arena) getPointerByOffset(offset uint32) unsafe.Pointer {
	return unsafe.Pointer(&a.buf[offset])
}

func (a *Arena) getPointerOffset(ptr unsafe.Pointer) uint32 {
	if ptr == nil {
		return 0
	}
	return uint32(uintptr(ptr) - uintptr(unsafe.Pointer(&a.buf[0])))
}

func (a *Arena) getBytes(offset, size uint32) []byte {
	return a.buf[offset:offset + size]
}
