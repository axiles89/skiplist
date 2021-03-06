package main

import (
	"fmt"
	"github.com/axiles89/skiplist"
	"math/rand"
	"time"
)

// 0 1 2 3 4 5 6 7 8 9 10 11 12 13

func main() {

	rand.Seed(time.Now().UnixNano())

	t := rand.Float32()

	arena := skiplist.NewArena(10000)
	list, _ := skiplist.NewList(arena)

	err := list.Add([]byte("a"), []byte("b"))
	err = list.Add([]byte("a"), []byte("d"))
	fmt.Println(err, t)
}
