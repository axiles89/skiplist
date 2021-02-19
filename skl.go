package skiplist

import (
	"bytes"
	"fmt"
	"github.com/axiles89/skiplist/fastrand"
	"math"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	pValue = 1 / math.E
)

type needToUpdate [maxHeight]struct{
	node *node
	nextOffset uint32
}

func (nd *needToUpdate) add(level uint32, nodeForUpdate *node, nextOffset uint32) {
	nd[level] = struct {
		node       *node
		nextOffset uint32
	}{node: nodeForUpdate, nextOffset: nextOffset}
}

type List struct {
	arena *Arena
	head *node
	level uint32
	mu sync.Mutex
	mm map[uint32]bool
}

func NewList(arena *Arena) (*List, error) {
	l := List{
		arena: arena,
		level: 1,
		mm: make(map[uint32]bool),
	}
	err := l.addHeadNode()
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (l *List) getLevel() uint32 {
	return atomic.LoadUint32(&l.level)
}

func (l *List) updateLevel(old, new uint32) bool {
	for {
		if old >= new {
			return true
		}
		if atomic.CompareAndSwapUint32(&l.level, old, new) {
			return true
		}
		old = l.getLevel()
	}
}


func (l *List) addHeadNode() error {
	offset, err := l.arena.alloc(sizeNode, 0)
	if err != nil {
		return err
	}
	l.head = (*node)(l.arena.getPointerByOffset(offset))
	return nil
}

func (l *List) newNode(key, value []byte, level uint32) (*node, error) {
	keySize, valueSize := uint32(len(key)), uint32(len(value))

	unusedSize := (maxHeight - level) * 4
	//s := sizeNode
	//fmt.Println(s)
	offset, err := l.arena.alloc(sizeNode - unusedSize + keySize + valueSize, unusedSize)
	if err != nil {
		return nil, err
	}
	newNode := (*node)(l.arena.getPointerByOffset(offset))
	newNode.keySize = keySize
	newNode.valueSize = valueSize
	newNode.keyOffset = offset + sizeNode - unusedSize

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

func (l *List) findUpdateNodeForLevel(currentNode *node, level uint32, key []byte, j bool) (updateNode *node, nextOffset uint32) {
	var (
		nextNode *node
		testNode *node
		testOffset uint32
		nextKey []byte
	)
	for {
		nextOffset = currentNode.getOffsetForLevel(level)

		//if we dont have offset(either currentNode = head or last node)
		if nextOffset == 0 {
			updateNode = currentNode
			break
		}

		if nextOffset > uint32(len(l.arena.buf)) {
			fmt.Println(nextNode, testOffset, testNode)
		}

		nextNode = (*node)(l.arena.getPointerByOffset(nextOffset))

		if nextNode.keyOffset + nextNode.keySize  > uint32(len(l.arena.buf)) {
			nextOffset2 := currentNode.getOffsetForLevel(level)
			fmt.Println(nextNode, testOffset, testNode, nextOffset2)
		}
		nextKey = nextNode.getKey(l.arena)
		// if key < nextKey
		if bytes.Compare(nextKey, key) > 0{
			updateNode = currentNode
			break
		} else {
			testNode = currentNode
			testOffset = nextOffset
			currentNode = nextNode
		}
	}
	return
}

func (l *List) prepareNeedToUpdate(key []byte, needToUpdate *needToUpdate) {
	currentNode := l.head

	var (
		level uint32
	)
	for currentLevel := l.getLevel(); currentLevel > 0; currentLevel-- {
		level = currentLevel - 1
		updateNode, nextOffset := l.findUpdateNodeForLevel(currentNode, level, key, true)
		currentNode = updateNode
		needToUpdate.add(level, currentNode, nextOffset)
	}
}

func (l *List) Add(key, value []byte) error {
	var needToUpdate needToUpdate

	l.prepareNeedToUpdate(key, &needToUpdate)

	newNodeLevel := l.randomLevel()
	newNode, err := l.newNode(key, value, newNodeLevel)
	if err != nil {
		return err
	}

	offsetNewNode := l.arena.getPointerOffset(unsafe.Pointer(newNode))

	var (
		updateNode *node
		lastOffset uint32 = 0
		currentListLevel = l.getLevel()
		level uint32 = 0
	)
	for ; level < newNodeLevel; level++ {
		if level > currentListLevel - 1 {
			updateNode = l.head
			lastOffset = 0
		} else {
			// Если с момента формирование needToUpdate не появилось новых занятых уровней
			if needToUpdate[level].node != nil {
				updateNode = needToUpdate[level].node
				lastOffset = needToUpdate[level].nextOffset
			} else {
				// Если появился новый уровень, то проходим по нему и ищем node для обновляния offset
				updateNode, lastOffset = l.findUpdateNodeForLevel(l.head, level, key, false)
			}
		}

		// ttt         1     n    2
		for {
			newNode.levels[level] = updateNode.getOffsetForLevel(level)
			//newNode.levels[level] = atomic.LoadUint32(&((*updateNode).levels[level]))
			if updateNode.casUpdateOffsetForLevel(level, lastOffset, offsetNewNode) {
				break
			}
			updateNode, lastOffset = l.findUpdateNodeForLevel(updateNode, level, key, false)
		}
	}

	l.updateLevel(currentListLevel, newNodeLevel)
	return nil
}
