package skiplist

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAlloc(t *testing.T) {
	t.Run("Alloc with empty arena", func(t *testing.T) {
		arena := NewArena(40)
		offset, err := arena.alloc(3)
		require.Nil(t, err)
		require.Equal(t, uint32(0), offset, "offset is wrong")
	})
	t.Run("Alloc with allign", func(t *testing.T) {
		arena := NewArena(40)
		arena.newSize = 2
		offset, err := arena.alloc(3)
		require.Nil(t, err)
		require.Equal(t, uint32(4), offset, "offset is wrong")
	})
	t.Run("Arena is full before alloc", func(t *testing.T) {
		arena := NewArena(2)
		offset, err := arena.alloc(3)
		require.Empty(t, offset)
		require.ErrorIs(t, err, ErrArenaFull, "arena did not return error")
	})
	t.Run("Arena is full after alloc", func(t *testing.T) {
		arena := NewArena(4)
		arena.newSize = 3
		offset, err := arena.alloc(3)
		require.Empty(t, offset)
		require.ErrorIs(t, err, ErrArenaFull, "arena did not return error")
	})
}
