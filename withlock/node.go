package withlock

import "unsafe"

const (
	sizeNode = uint32(unsafe.Sizeof(node{}))
	maxHeight = 20
	//p = 0.25
	p = 0.44
)

type node struct {
	keySize uint32
	valueSize uint32
	keyOffset uint32
	levels [maxHeight]uint32
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
