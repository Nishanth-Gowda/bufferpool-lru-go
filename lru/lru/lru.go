package lru

import "github.com/nishanthgowda/btree/lru/doubly-ll"

type LRUCache struct {
	capacity int
	cache    map[int]*doublyll.Node
	list     *doublyll.DoublyLinkedList
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[int]*doublyll.Node),
		list:     doublyll.NewDoublyLinkedList(),
	}
}

func (lru *LRUCache) Get(key int) int {

	node, ok := lru.cache[key]
	if !ok {
		return -1
	}

	lru.list.RemoveNode(node)
	lru.list.AddFront(node)
	return node.Value
}

func (lru *LRUCache) Put(key int, value int) {

	node, ok := lru.cache[key]

	// If key is already present, update the value and move to front
	if ok {
		node.Value = value
		lru.list.RemoveNode(node)
		lru.list.AddFront(node)
		return
	}

	// If cache is full, remove the least recently used item
	if len(lru.cache) == lru.capacity {
		delete(lru.cache, lru.list.Tail.Key)
		lru.list.RemoveNode(lru.list.Tail)
	}

	newNode := &doublyll.Node{
		Key:   key,
		Value: value,
	}

	// Add the new item to the front of the list
	lru.cache[key] = newNode
	lru.list.AddFront(newNode)
}
