package skiplist

import (
	"encoding/binary"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func randomKey(rng *rand.Rand) []byte {
	b := make([]byte, 8)
	key := rng.Uint32()
	key2 := rng.Uint32()
	binary.LittleEndian.PutUint32(b, key)
	binary.LittleEndian.PutUint32(b[4:], key2)
	return b
}

// Standard test. Some fraction is read. Some fraction is write. Writes have
// to go through mutex lock.
func BenchmarkReadWriteMap(b *testing.B) {
	m := make(map[string][]byte)
	value := []byte("a")
	var mutex sync.RWMutex
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			mutex.Lock()
			m[string(randomKey(rng))] = value
			mutex.Unlock()
		}
	})
}

func TestAdd10000(t *testing.T) {
	r := sizeNode + 2
	var arena = NewArena(uint32(10000000 + 1) * r)
	var list, _ = NewList(arena)
	var w sync.WaitGroup
	for i := 0; i < 4; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			for {
				list.Add([]byte("a"), []byte("b"))
			}
		}()
	}

	w.Wait()
}


func BenchmarkParallelAdd2(b *testing.B) {
	sizeHead := sizeNode + align
	sizeNodeWithAlign := sizeNode + 11 + align

	b.Run("run parallel", func(b *testing.B) {
		var arena = NewArena(sizeHead + (uint32(b.N) * sizeNodeWithAlign))
		var list, _ = NewList(arena)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				token := make([]byte, 10)
				rand.Read(token)
				err := list.Add(token, []byte("b"))
				if err != nil {
					b.Error(err)
					b.FailNow()
				}
			}
		})
	})
}
