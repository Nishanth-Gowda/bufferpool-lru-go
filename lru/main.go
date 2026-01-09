package main

import (
	"fmt"

	"github.com/nishanthgowda/btree/lru/lru"
)

func main() {
	lruCache := lru.NewLRUCache(2)
	lruCache.Put(1, 1)
	lruCache.Put(2, 2)
	fmt.Println(lruCache.Get(1))
	lruCache.Put(3, 3)
	fmt.Println(lruCache.Get(2))
	lruCache.Put(4, 4)
	fmt.Println(lruCache.Get(1))
	fmt.Println(lruCache.Get(3))
	fmt.Println(lruCache.Get(4))
}
