package lru_test

import (
	"testing"

	bufferpool "github.com/nishanthgowda/btree/lru/bufferpool-lru"
	"github.com/nishanthgowda/btree/lru/lru"
)

// Benchmark scenarios to test:
// 1. Sequential writes with no eviction
// 2. Sequential writes with eviction
// 3. Mixed read/write operations (cache hits)
// 4. Mixed read/write operations (cache misses)
// 5. High locality access pattern (80/20 rule)

// BenchmarkNormalLRU_SequentialWrites tests sequential writes with no eviction
func BenchmarkNormalLRU_SequentialWrites(b *testing.B) {
	cache := lru.NewLRUCache(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Put(i%1000, i)
	}
}

// BenchmarkBufferPoolLRU_SequentialWrites tests sequential writes with no eviction
func BenchmarkBufferPoolLRU_SequentialWrites(b *testing.B) {
	cache := bufferpool.NewBufferPool(1000, 0.7) // 70% old list
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Put(i%1000, i)
	}
}

// BenchmarkNormalLRU_WithEviction tests sequential writes that trigger evictions
func BenchmarkNormalLRU_WithEviction(b *testing.B) {
	cache := lru.NewLRUCache(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Put(i, i)
	}
}

// BenchmarkBufferPoolLRU_WithEviction tests sequential writes that trigger evictions
func BenchmarkBufferPoolLRU_WithEviction(b *testing.B) {
	cache := bufferpool.NewBufferPool(100, 0.7)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Put(i, i)
	}
}

// BenchmarkNormalLRU_CacheHits tests read performance with 100% cache hits
func BenchmarkNormalLRU_CacheHits(b *testing.B) {
	cache := lru.NewLRUCache(1000)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(i % 1000)
	}
}

// BenchmarkBufferPoolLRU_CacheHits tests read performance with 100% cache hits
func BenchmarkBufferPoolLRU_CacheHits(b *testing.B) {
	cache := bufferpool.NewBufferPool(1000, 0.7)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(i % 1000)
	}
}

// BenchmarkNormalLRU_CacheMisses tests read performance with 100% cache misses
func BenchmarkNormalLRU_CacheMisses(b *testing.B) {
	cache := lru.NewLRUCache(100)

	// Pre-populate cache with different keys
	for i := 0; i < 100; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(i + 1000) // These keys don't exist
	}
}

// BenchmarkBufferPoolLRU_CacheMisses tests read performance with 100% cache misses
func BenchmarkBufferPoolLRU_CacheMisses(b *testing.B) {
	cache := bufferpool.NewBufferPool(100, 0.7)

	// Pre-populate cache with different keys
	for i := 0; i < 100; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(i + 1000) // These keys don't exist
	}
}

// BenchmarkNormalLRU_MixedOps tests mixed read/write operations
func BenchmarkNormalLRU_MixedOps(b *testing.B) {
	cache := lru.NewLRUCache(1000)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			cache.Get(i % 1000)
		} else {
			cache.Put(i%1000, i)
		}
	}
}

// BenchmarkBufferPoolLRU_MixedOps tests mixed read/write operations
func BenchmarkBufferPoolLRU_MixedOps(b *testing.B) {
	cache := bufferpool.NewBufferPool(1000, 0.7)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			cache.Get(i % 1000)
		} else {
			cache.Put(i%1000, i)
		}
	}
}

// BenchmarkNormalLRU_HighLocality simulates high locality access pattern (80/20 rule)
// 80% of accesses go to 20% of the data
func BenchmarkNormalLRU_HighLocality(b *testing.B) {
	cache := lru.NewLRUCache(1000)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%10 < 8 {
			// 80% of accesses to first 200 items (20% of cache)
			cache.Get(i % 200)
		} else {
			// 20% of accesses to remaining 800 items
			cache.Get(200 + (i % 800))
		}
	}
}

// BenchmarkBufferPoolLRU_HighLocality simulates high locality access pattern (80/20 rule)
func BenchmarkBufferPoolLRU_HighLocality(b *testing.B) {
	cache := bufferpool.NewBufferPool(1000, 0.7)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%10 < 8 {
			// 80% of accesses to first 200 items (20% of cache)
			cache.Get(i % 200)
		} else {
			// 20% of accesses to remaining 800 items
			cache.Get(200 + (i % 800))
		}
	}
}

// BenchmarkNormalLRU_UpdateExisting tests performance when updating existing keys
func BenchmarkNormalLRU_UpdateExisting(b *testing.B) {
	cache := lru.NewLRUCache(1000)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Put(i%1000, i*2) // Update existing values
	}
}

// BenchmarkBufferPoolLRU_UpdateExisting tests performance when updating existing keys
func BenchmarkBufferPoolLRU_UpdateExisting(b *testing.B) {
	cache := bufferpool.NewBufferPool(1000, 0.7)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Put(i%1000, i*2) // Update existing values
	}
}

// BenchmarkNormalLRU_SmallCache tests performance with a very small cache (high eviction rate)
func BenchmarkNormalLRU_SmallCache(b *testing.B) {
	cache := lru.NewLRUCache(10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Put(i, i)
	}
}

// BenchmarkBufferPoolLRU_SmallCache tests performance with a very small cache (high eviction rate)
func BenchmarkBufferPoolLRU_SmallCache(b *testing.B) {
	cache := bufferpool.NewBufferPool(10, 0.7)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Put(i, i)
	}
}

// BenchmarkNormalLRU_LargeCache tests performance with a large cache
func BenchmarkNormalLRU_LargeCache(b *testing.B) {
	cache := lru.NewLRUCache(10000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Put(i%10000, i)
	}
}

// BenchmarkBufferPoolLRU_LargeCache tests performance with a large cache
func BenchmarkBufferPoolLRU_LargeCache(b *testing.B) {
	cache := bufferpool.NewBufferPool(10000, 0.7)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Put(i%10000, i)
	}
}
