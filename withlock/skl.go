package withlock

import (
	"bytes"
	"github.com/axiles89/skiplist/fastrand"
	"math"
	"sync"
)

const (
	pValue = 1 / math.E
)

type needToUpdate [maxHeight]*node

func (nd *needToUpdate) add(level uint32, node *node) {
	nd[level] = node
}

type List struct {
	arena *Arena
	head *node
	level uint32
	sync.Mutex
}

func NewList(arena *Arena) (*List, error) {
	l := List{
		arena: arena,
		level: 2,
	}
	err := l.addHeadNode()
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (l *List) addHeadNode() error {
	offset, err := l.arena.alloc(sizeNode)
	if err != nil {
		return err
	}
	l.head = (*node)(l.arena.getPointerByOffset(offset))
	return nil
}

func (l *List) newNode(key, value []byte) (*node, error) {
	keySize, valueSize := uint32(len(key)), uint32(len(value))
	offset, err := l.arena.alloc(sizeNode + keySize + valueSize)
	if err != nil {
		return nil, err
	}
	newNode := (*node)(l.arena.getPointerByOffset(offset))
	newNode.keySize = keySize
	newNode.valueSize = valueSize
	newNode.keyOffset = offset + sizeNode

	copy(newNode.getKey(l.arena), key)
	copy(newNode.getValue(l.arena), value)
	return newNode, nil
}

var (
	probabilities [maxHeight]uint32
)

func init() {
	// Precompute the skiplist probabilities so that only a single random number
	// needs to be generated and so that the optimal pvalue can be used (inverse
	// of Euler's number).
	p := float64(1.0)
	for i := 0; i < maxHeight; i++ {
		probabilities[i] = uint32(float64(math.MaxUint32) * p)
		p *= pValue
	}
}

func (l *List) randomLevel() uint32 {
	rnd := fastrand.Uint32()

	h := uint32(1)

	for h < maxHeight && rnd <= probabilities[h] {
		h++
	}

	return h
}

func (l *List) needToUpdate(key []byte, needToUpdate *needToUpdate) {
	currentNode := l.head

	var (
		nextOffset, level uint32
		nextKey []byte
	)
	for currentLevel := l.level; currentLevel > 0; currentLevel-- {
		level = currentLevel - 1
		for {
			nextOffset = currentNode.levels[level]
			//if we dont have offset(either currentNode = head or last node)
			if nextOffset == 0 {
				needToUpdate.add(level, currentNode)
				break
			}
			nextNode := (*node)(l.arena.getPointerByOffset(nextOffset))
			nextKey = nextNode.getKey(l.arena)
			// if key <= nextKey
			if bytes.Compare(nextKey, key) >= 0{
				needToUpdate.add(level, currentNode)
				break
			} else {
				currentNode = nextNode
			}
		}
	}
}

func (l*List) AddWithoutLock(key, value []byte) error {
	return nil
}

func (l *List) AddWithLock(key, value []byte) error {
	l.Lock()
	defer l.Unlock()
	var needToUpdate needToUpdate

	l.needToUpdate(key, &needToUpdate)

	newNode, err := l.newNode(key, value)
	if err != nil {
		return err
	}
	newNodeLevel := l.randomLevel()
	offsetNewNode := newNode.getOffset()

	var updateNode *node
	for currentLevel := newNodeLevel; currentLevel > 0; currentLevel-- {
		level := currentLevel - 1
		if level > l.level - 1 {
			updateNode = l.head
		} else {
			updateNode = needToUpdate[level]
		}

		// ttt         1     n    2

		newNode.levels[level] = updateNode.levels[level]
		updateNode.levels[level] = offsetNewNode
	}
	return nil
}
