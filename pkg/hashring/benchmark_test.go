package hashring

import (
	"fmt"
	"testing"
)

func BenchmarkHashKey(b *testing.B) {
	keys := []string{
		"user123",
		"client456",
		"tenant789",
		"very-long-key-with-many-characters-to-test-performance",
		"short",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%len(keys)]
		_ = hashKey(key)
	}
}

func BenchmarkConsistentHashRing_GetNode(b *testing.B) {
	ring := NewConsistentHashRing(3).(*ConsistentHashRing)

	// Setup shards
	shards := []string{
		"http://shard01:80",
		"http://shard02:80",
		"http://shard03:80",
		"http://shard04:80",
		"http://shard05:80",
	}

	for _, shard := range shards {
		ring.AddNode(shard)
	}

	// Test keys
	keys := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		keys[i] = fmt.Sprintf("user%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%len(keys)]
		_ = ring.GetNode(key)
	}
}

func BenchmarkConsistentHashRing_AddNode(b *testing.B) {
	ring := NewConsistentHashRing(10).(*ConsistentHashRing)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ring.AddNode(fmt.Sprintf("shard%d", i))
	}
}

func BenchmarkConsistentHashRing_GetNode_DifferentRingSizes(b *testing.B) {
	ringsSizes := []int{1, 3, 5, 10, 50, 100}

	for _, size := range ringsSizes {
		b.Run(fmt.Sprintf("RingSize_%d", size), func(b *testing.B) {
			ring := NewConsistentHashRing(size).(*ConsistentHashRing)

			// Add shards
			for i := 0; i < 5; i++ {
				ring.AddNode(fmt.Sprintf("shard%d", i))
			}

			// Test keys
			keys := make([]string, 100)
			for i := 0; i < 100; i++ {
				keys[i] = fmt.Sprintf("user%d", i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := keys[i%len(keys)]
				_ = ring.GetNode(key)
			}
		})
	}
}

func BenchmarkConsistentHashRing_Distribution(b *testing.B) {
	ring := NewConsistentHashRing(5).(*ConsistentHashRing)

	// Setup 3 shards
	shards := []string{"shard01", "shard02", "shard03"}
	for _, shard := range shards {
		ring.AddNode(shard)
	}

	b.ResetTimer()

	// Test distribution performance
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("user%d", i)
		_ = ring.GetNode(key)
	}
}
