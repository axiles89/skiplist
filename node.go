package skiplist

import (
	"sync/atomic"
	"unsafe"
)

const (
	sizeNode = uint32(unsafe.Sizeof(node{}))
	maxHeight = 20
)

type node struct {
	keySize uint32
	valueSize uint32
	keyOffset uint32
	levels [maxHeight]uint32
}


func (n *node) getOffsetForLevel(level uint32) uint32  {
	l := atomic.LoadUint32(&n.levels[level])
	return l
}

func (n *node) casUpdateOffsetForLevel(level, oldOffset, newOffset uint32) bool {
	return atomic.CompareAndSwapUint32(&n.levels[level], oldOffset, newOffset)
}

func (n *node) getOffset() uint32 {
	return n.keyOffset - sizeNode
}

func (n *node) getKey(arena *Arena) []byte {
	return arena.getBytes(n.keyOffset, n.keySize)
}

func (n *node) getValue(arena *Arena) []byte {
	return arena.getBytes(n.keyOffset + n.keySize, n.valueSize)
}
