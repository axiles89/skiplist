package withlock

import (
	"strconv"
	"sync"
	"testing"
)

func BenchmarkParallelAdd2(b *testing.B) {
	r := sizeNode + 2
	var arena = NewArena(uint32(b.N + 1) * r)
	var list, _ = NewList(arena)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := list.AddWithLock([]byte("a"), []byte("b"))
			if err != nil {
				b.Error(err)
				b.FailNow()
			}
		}
	})
	return
}

func BenchmarkAdd(b *testing.B) {
	var arena = NewArena(2 << 20)
	var list, _ = NewList(arena)

	b.ResetTimer()
	for i := 0; i < 5000; i++ {
		err := list.AddWithLock([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}

func BenchmarkAddWithLock500(b *testing.B) {
	arena := NewArena(2 << 20)
	list, _ := NewList(arena)
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < 5000; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := list.AddWithLock([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
			if err != nil {
				b.Error(err)
				b.FailNow()
			}
		}()
	}

	wg.Wait()
}
