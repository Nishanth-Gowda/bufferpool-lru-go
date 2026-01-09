# LRU Cache Implementations in Go

A comprehensive comparison of **two LRU cache implementations**: a traditional LRU cache and a MySQL InnoDB-inspired BufferPool LRU cache, complete with extensive benchmarks.

## 📋 Table of Contents
- [Implementations](#implementations)
- [The Inspiration: MySQL InnoDB Buffer Pool](#the-inspiration-mysql-innodb-buffer-pool)
- [How BufferPool LRU Works](#how-bufferpool-lru-works)
- [Benchmark Results](#benchmark-results)
- [When to Use Which](#when-to-use-which)
- [Usage Examples](#usage-examples)
- [Project Structure](#project-structure)

---

## Implementations

### 1. **Normal LRU Cache** (`lru/lru/`)
A classic LRU cache implementation using HashMap + Doubly Linked List.

**Features:**
- O(1) `Get` and `Put` operations
- Most recently used items at the front (head)
- Least recently used items at the back (tail)
- Automatic eviction when capacity is reached

### 2. **BufferPool LRU Cache** (`lru/bufferpool-lru/`)
An advanced LRU implementation inspired by MySQL InnoDB's buffer pool management strategy.

**Features:**
- O(1) operations (same as normal LRU)
- Splits cache into **"New"** and **"Old"** sublists
- Configurable split ratio (e.g., 70% old, 30% new)
- Protects frequently accessed items from eviction
- Better performance for high-locality workloads

---

## The Inspiration: MySQL InnoDB Buffer Pool

### Background
MySQL's InnoDB storage engine uses a sophisticated buffer pool management algorithm to cache frequently accessed database pages in memory. The challenge it solves:

> **Problem**: Traditional LRU evicts items that were accessed once recently, even if other items are accessed repeatedly. This is suboptimal for database workloads with full table scans or bulk operations.

### InnoDB's Solution: The Midpoint Insertion Strategy

InnoDB maintains a **single list** but logically splits it into two regions:
```
[New List (30%)] → [MidPoint] → [Old List (70%)]
     ↑ Hot data         ↑           ↑ Warm/cold data
```

**Key Behaviors:**
1. **New pages** are inserted at the **midpoint** (not the head!), going into the "Old" sublist
2. **Only when accessed again** do they get promoted to the "New" sublist (moved to head)
3. This prevents a single table scan from evicting all your hot cache data

### Why This Matters
Consider a database scenario:
- You have 1000 frequently accessed customer records (hot data)
- Someone runs `SELECT * FROM orders` scanning 10,000 order records (cold data)
- **Traditional LRU**: The scan would evict all 1000 customer records ❌
- **BufferPool LRU**: Scan pages enter the "Old" list and evict each other, preserving hot customer data ✅

---

## How BufferPool LRU Works

### Architecture
```
BufferPool Structure:
┌─────────────────────────────────────────────────┐
│ HashMap (cache)                                 │
│ {key1: *Node1, key2: *Node2, ...}              │
└─────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────┐
│ Doubly Linked List                              │
│                                                 │
│  [Head] ← New List (IsOld=false) → [MidPoint]  │
│                     30%                  ↓      │
│                                    Old List     │
│                                   (IsOld=true)  │
│                                        70%      │
│                                         ↓       │
│                                      [Tail]     │
└─────────────────────────────────────────────────┘
```

### Key Components

#### 1. **Node Structure**
```go
type Node struct {
    Key   int
    Value int
    Prev  *Node
    Next  *Node
    IsOld bool  // Tracks which sublist the node belongs to
}
```

#### 2. **BufferPool Fields**
```go
type BufferPool struct {
    capacity       int              // Total cache size
    cache          map[int]*Node    // O(1) lookup
    list           *DoublyLinkedList
    MidPoint       *Node            // Boundary between new and old
    OldRatio       float64          // E.g., 0.7 for 70% old
    MaxOldSize     int              // Target old list size
    currentOldSize int              // Actual old list size
}
```

### Operations

#### **Put (Insert)**
```
New item → Insert at MidPoint (becomes head of "Old" list)
         → Mark as IsOld = true
         → If Old list exceeds MaxOldSize, move MidPoint forward
```

#### **Get (Access)**
```
If node.IsOld == true:
    → Promote to "New" list (move to Head)
    → Set IsOld = false
    → Decrement currentOldSize
    → Update MidPoint if necessary
Else:
    → Already in "New" list, just move to Head
```

#### **Eviction**
```
Always evict from Tail (the oldest item in "Old" list)
→ This protects "New" list items from eviction
```

### Example Flow

```
Initial state: capacity=5, OldRatio=0.6 (MaxOldSize=3)

Step 1: Put(1,A), Put(2,B), Put(3,C)
List: [1]→[2]→[3]
       ↑MidPoint
All marked IsOld=true

Step 2: Get(1) - Promotes to New
List: [1]→[2]→[3]
      New  ↑   Old
           MidPoint

Step 3: Put(4,D), Put(5,E)
List: [1]→[4]→[5]→[2]→[3]
      New        ↑  Old
                 MidPoint

Step 4: Get(3) - Promotes from Old to New
List: [3]→[1]→[4]→[5]→[2]
      New            ↑  Old
                     MidPoint

Step 5: Put(6,F) - Cache full, evicts [2]
List: [3]→[1]→[6]→[4]→[5]
      New            ↑  Old
                     MidPoint
```

---

## Benchmark Results

Benchmarked on **Apple M4 Pro** (12 cores, ARM64) with **3-second runs per benchmark**.

### Complete Results

| Benchmark Scenario | Normal LRU | BufferPool LRU | Winner | % Difference |
|-------------------|------------|----------------|---------|--------------|
| **Sequential Writes** | 10.05 ns/op | **9.63 ns/op** | BufferPool | **+4.2%** ✅ |
| **With Eviction** | **30.84 ns/op** | 39.91 ns/op | Normal | **-29.4%** ⚠️ |
| **Cache Hits** | 7.92 ns/op | **7.86 ns/op** | BufferPool | +0.8% ✅ |
| **Cache Misses** | 11.28 ns/op | **10.39 ns/op** | BufferPool | **+7.9%** ✅ |
| **Mixed Ops (50/50)** | 9.24 ns/op | **8.64 ns/op** | BufferPool | **+6.6%** ✅ |
| **High Locality (80/20)** | 9.80 ns/op | **9.40 ns/op** | BufferPool | **+4.1%** ✅ |
| **Update Existing** | 10.88 ns/op | **9.60 ns/op** | BufferPool | **+11.7%** ✅ ⭐ |
| **Small Cache (10)** | **32.68 ns/op** | 44.08 ns/op | Normal | **-34.9%** ⚠️ |
| **Large Cache (10K)** | 14.51 ns/op | **14.02 ns/op** | BufferPool | **+3.4%** ✅ |

**Memory Usage**: Both implementations use **48 B/op** and **1 alloc/op** for operations requiring new node allocation.

### Key Takeaways

#### ✅ **BufferPool LRU Wins in 7/9 Scenarios**

1. **Best Performance**: Update operations (+11.7%)
   - The midpoint strategy reduces unnecessary list reorganization for repeated updates

2. **Strong in Mixed Workloads**: +6.6% improvement
   - Real-world caches have mixed read/write patterns

3. **High Locality Advantage**: +4.1% improvement
   - Validates the core design goal: protect frequently accessed data

4. **Scales Better**: Large cache performance (+3.4%)
   - Overhead becomes negligible with larger caches

#### ⚠️ **Normal LRU Wins in High-Eviction Scenarios**

1. **Eviction Overhead**: -29.4% slower
   - BufferPool must update midpoint, adjust counters, check IsOld flags

2. **Small Cache Penalty**: -34.9% slower
   - Overhead is proportionally higher with tiny caches

### Performance Distribution

```
Sequential Writes:    [████████████████████] BufferPool +4.2%
Cache Hits:           [████████████████████] BufferPool +0.8%
Cache Misses:         [████████████████████] BufferPool +7.9%
Mixed Operations:     [████████████████████] BufferPool +6.6%
High Locality:        [████████████████████] BufferPool +4.1%
Update Existing:      [████████████████████] BufferPool +11.7% ⭐
Large Cache:          [████████████████████] BufferPool +3.4%

With Eviction:        [█████████████       ] Normal +29.4% faster
Small Cache:          [█████████████       ] Normal +34.9% faster
```

---

## When to Use Which

### 🎯 Use **BufferPool LRU** when:

✅ **Frequent updates** to existing cache entries  
✅ **High locality** workloads (80/20 rule: 80% of accesses to 20% of data)  
✅ **Large cache sizes** (>1000 entries)  
✅ **Mixed read/write** patterns (typical in databases, web applications)  
✅ **High cache hit rates** (>70%)  
✅ **Protection needed** against scan/bulk operation pollution

**Ideal Use Cases:**
- Database buffer pools
- Web application caches (user sessions, popular content)
- CDN edge caches
- Application-level query result caches
- Key-value stores with hot key patterns

### 🎯 Use **Normal LRU** when:

✅ **Very small cache** (<100 entries)  
✅ **High eviction rate** (cache thrashing scenarios)  
✅ **Uniform access patterns** (no clear hot/cold data)  
✅ **Simplicity matters** more than performance  
✅ **Memory overhead** must be minimized (no midpoint tracking)

**Ideal Use Cases:**
- Simple browser caches
- Small in-memory lookups
- Educational implementations
- Embedded systems with limited memory

---

## Usage Examples

### Normal LRU Cache

```go
package main

import (
    "fmt"
    "github.com/nishanthgowda/btree/lru/lru"
)

func main() {
    cache := lru.NewLRUCache(3)
    
    cache.Put(1, 100)  // cache: {1=100}
    cache.Put(2, 200)  // cache: {1=100, 2=200}
    cache.Put(3, 300)  // cache: {1=100, 2=200, 3=300}
    
    fmt.Println(cache.Get(1))  // 100, moves 1 to front
    
    cache.Put(4, 400)  // evicts 2, cache: {1=100, 3=300, 4=400}
    
    fmt.Println(cache.Get(2))  // -1 (not found)
    fmt.Println(cache.Get(3))  // 300
}
```

### BufferPool LRU Cache

```go
package main

import (
    "fmt"
    bufferpool "github.com/nishanthgowda/btree/lru/bufferpool-lru"
)

func main() {
    // 70% old list, 30% new list
    cache := bufferpool.NewBufferPool(10, 0.7)
    
    // Insert items - they go to midpoint (head of Old list)
    for i := 1; i <= 10; i++ {
        cache.Put(i, i*100)
    }
    
    // Accessing an item promotes it to New list
    val := cache.Get(5)  // Moves 5 to "New" list
    fmt.Println(val)     // 500
    
    // Insert new item - evicts from tail of Old list
    cache.Put(11, 1100)  // Evicts oldest item in Old list
    
    // Item 5 is protected in New list and won't be evicted next
    fmt.Println(cache.Get(5))  // 500 (still present)
}
```

### Simulating Database Workload

```go
func simulateDatabaseWorkload() {
    cache := bufferpool.NewBufferPool(1000, 0.7)
    
    // Load hot data (frequently accessed customer records)
    hotKeys := []int{1, 2, 3, 4, 5}
    for _, key := range hotKeys {
        cache.Put(key, key*100)
        cache.Get(key)  // Promotes to New list
    }
    
    // Simulate a table scan (lots of one-time accesses)
    for i := 1000; i < 2000; i++ {
        cache.Put(i, i)  // These go to Old list
    }
    
    // Hot data should still be accessible
    for _, key := range hotKeys {
        val := cache.Get(key)
        if val != -1 {
            fmt.Printf("Hot key %d still in cache: %d\n", key, val)
        }
    }
}
```

---

## Project Structure

```
lru/
├── README.md                     # This file
├── lru/                          # Normal LRU implementation
│   └── lru.go
├── bufferpool-lru/               # BufferPool LRU implementation
│   └── lru.go
├── doubly-ll/                    # Shared doubly linked list
│   └── doubly_linked_list.go
├── benchmarks/                   # Comprehensive benchmark suite
│   └── lru_benchmark_test.go
└── main.go                       # Demo/test program
```

---

## Running Benchmarks

### Quick Run
```bash
cd /path/to/B-Trees/lru
go test -bench=. -benchmem ./benchmarks/
```

### Detailed Run (3 seconds per benchmark)
```bash
go test -bench=. -benchmem -benchtime=3s ./benchmarks/
```

### Run Specific Benchmarks
```bash
# Only Normal LRU
go test -bench=BenchmarkNormalLRU -benchmem ./benchmarks/

# Only BufferPool LRU
go test -bench=BenchmarkBufferPoolLRU -benchmem ./benchmarks/

# Only high locality tests
go test -bench=HighLocality -benchmem ./benchmarks/
```

### Compare with benchstat
```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Run and save results
go test -bench=BenchmarkNormalLRU -benchmem ./benchmarks/ > normal.txt
go test -bench=BenchmarkBufferPoolLRU -benchmem ./benchmarks/ > bufferpool.txt

# Compare
benchstat normal.txt bufferpool.txt
```

---

## Technical Details

### Time Complexity
Both implementations maintain **O(1)** time complexity:
- **Get(key)**: O(1) - HashMap lookup + list manipulation
- **Put(key, value)**: O(1) - HashMap insert + list manipulation

### Space Complexity
- **Normal LRU**: O(capacity)
- **BufferPool LRU**: O(capacity) + O(1) for midpoint tracking
  - Additional space: 1 pointer (MidPoint), 2 integers (MaxOldSize, currentOldSize), 1 float (OldRatio)
  - Per-node overhead: 1 boolean (IsOld)

### Concurrency
⚠️ **Neither implementation is thread-safe**. For concurrent use, wrap with `sync.RWMutex`:

```go
type ThreadSafeLRU struct {
    mu    sync.RWMutex
    cache *lru.LRUCache
}

func (t *ThreadSafeLRU) Get(key int) int {
    t.mu.RLock()
    defer t.mu.RUnlock()
    return t.cache.Get(key)
}

func (t *ThreadSafeLRU) Put(key, value int) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.cache.Put(key, value)
}
```

---

## Future Enhancements

### Possible Improvements
1. **Generic Types** (Go 1.18+)
   ```go
   type LRUCache[K comparable, V any] struct { ... }
   ```

2. **TTL Support** (time-based eviction)
   ```go
   type Node struct {
       ...
       ExpiresAt time.Time
   }
   ```

3. **Eviction Callbacks**
   ```go
   cache.OnEvict(func(key, value int) {
       log.Printf("Evicted: %d=%d", key, value)
   })
   ```

4. **Adaptive Ratio Tuning**
   - Dynamically adjust OldRatio based on access patterns
   - Monitor hit rates and eviction patterns

5. **Metrics and Monitoring**
   - Hit/miss ratios
   - Eviction counts
   - Promotion rates (Old → New)

---

## References

### MySQL InnoDB Buffer Pool
- [MySQL 8.0 Reference Manual - The InnoDB Buffer Pool](https://dev.mysql.com/doc/refman/8.0/en/innodb-buffer-pool.html)
- [Making the Buffer Pool Scan Resistant](https://dev.mysql.com/doc/refman/8.0/en/innodb-performance-midpoint_insertion.html)

### LRU Cache Algorithms
- [LeetCode Problem 146: LRU Cache](https://leetcode.com/problems/lru-cache/)
- [Cache Replacement Policies](https://en.wikipedia.org/wiki/Cache_replacement_policies)

---

## License

MIT License - feel free to use this in your projects!

---

## Contributing

Found a bug or have an optimization idea? Feel free to open an issue or submit a pull request!

---

**Built with ❤️ exploring database internals and cache algorithms**
