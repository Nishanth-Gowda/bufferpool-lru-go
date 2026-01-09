# LRU Cache Implementation

## Overview
This is an optimized implementation of an LRU (Least Recently Used) Cache in Go using a **HashMap + Doubly Linked List** data structure.

## Time Complexity
- **Get(key)**: O(1)
- **Put(key, value)**: O(1)

## Space Complexity
- O(capacity) - stores at most `capacity` items

## Data Structure Choice

### Why HashMap + Doubly Linked List?
- **HashMap**: Provides O(1) lookup for cache entries
- **Doubly Linked List**: Provides O(1) insertion/deletion for maintaining LRU order
  - Most recently used items are at the **front** (head)
  - Least recently used items are at the **back** (tail)

## Key Optimizations Implemented

### 1. Pointer Cleanup
**Why?** Prevents memory leaks and helps garbage collection.

```go
// In RemoveNode - clear pointers after removal
node.Prev = nil
node.Next = nil
```

### 2. Prevent Dangling References
**Why?** When a node is moved from middle to front, old `Prev` pointer must be cleared.

```go
// In AddFront - clear prev before re-linking
node.Prev = nil  // Critical for reused nodes
node.Next = list.Head
```

### 3. Direct Map Access
**Why?** Using `map[key]*Node` allows direct O(1) access without iteration.

```go
node, ok := lru.cache[key]  // O(1) lookup
```

## Potential Further Optimizations

### Option 1: Add Fast Path for Head Node
Skip unnecessary operations if the node is already at the front:

```go
func (lru *LRUCache) Get(key int) int {
    node, ok := lru.cache[key]
    if !ok {
        return -1
    }
    
    // Fast path: already at front
    if node == lru.list.Head {
        return node.Value
    }
    
    lru.list.RemoveNode(node)
    lru.list.AddFront(node)
    return node.Value
}
```

### Option 2: Use Sentinel Nodes (Dummy Head/Tail)
Eliminates nil checks in list operations:

```go
// Instead of nil head/tail, use dummy nodes
type DoublyLinkedList struct {
    Head *Node  // dummy sentinel
    Tail *Node  // dummy sentinel
}
```

### Option 3: Sync Pool for Node Allocation
Reuse node objects to reduce GC pressure:

```go
var nodePool = sync.Pool{
    New: func() interface{} {
        return &Node{}
    },
}
```

## Comparison with Other Approaches

| Approach | Get | Put | Space | Pros | Cons |
|----------|-----|-----|-------|------|------|
| **HashMap + DLL** (Current) | O(1) | O(1) | O(n) | Fast, Simple | Extra space for pointers |
| Array + Timestamps | O(n) | O(n) | O(n) | Simple | Slow eviction |
| Heap + HashMap | O(log n) | O(log n) | O(n) | Priority-based | Slower than DLL |
| Ring Buffer | O(1) | O(1) | O(n) | Cache-friendly | Complex implementation |

## Current vs Optimized Comparison

### Before Optimization
```go
func (list *DoublyLinkedList) AddFront(node *Node) {
    node.Next = list.Head  // ❌ node.Prev may point to old location
    list.Head.Prev = node
    list.Head = node
}

func (list *DoublyLinkedList) RemoveNode(node *Node) {
    // ... removal logic ...
    // ❌ Doesn't clear node.Prev and node.Next
}
```

**Issues:**
- Dangling references when reusing nodes
- Potential memory leaks
- GC has to trace unnecessary pointers

### After Optimization
```go
func (list *DoublyLinkedList) AddFront(node *Node) {
    node.Prev = nil  // ✅ Clear before adding
    node.Next = list.Head
    list.Head.Prev = node
    list.Head = node
}

func (list *DoublyLinkedList) RemoveNode(node *Node) {
    // ... removal logic ...
    node.Prev = nil  // ✅ Help GC
    node.Next = nil
}
```

**Benefits:**
- ✅ No dangling references
- ✅ Better garbage collection
- ✅ Cleaner memory state
- ✅ Easier debugging

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/nishanthgowda/btree/lru/lru"
)

func main() {
    cache := lru.NewLRUCache(2)
    
    cache.Put(1, 1)  // cache: {1=1}
    cache.Put(2, 2)  // cache: {1=1, 2=2}
    
    fmt.Println(cache.Get(1))  // returns 1, cache: {2=2, 1=1}
    
    cache.Put(3, 3)  // evicts key 2, cache: {1=1, 3=3}
    
    fmt.Println(cache.Get(2))  // returns -1 (not found)
    
    cache.Put(4, 4)  // evicts key 1, cache: {3=3, 4=4}
    
    fmt.Println(cache.Get(1))  // returns -1 (not found)
    fmt.Println(cache.Get(3))  // returns 3
    fmt.Println(cache.Get(4))  // returns 4
}
```

## Output
```
1
-1
-1
3
4
```

## Verdict
✅ **Yes, this is now an optimized implementation!**

The core algorithm was already optimal (O(1) operations), but the improvements made:
1. Fix memory leaks through proper pointer cleanup
2. Prevent dangling references
3. Improve garbage collection efficiency
4. Maintain clean memory state

For production use, consider adding:
- Thread safety (sync.RWMutex)
- Generic types for key/value (Go 1.18+)
- TTL (Time To Live) support
- Eviction callbacks
- Metrics/monitoring
