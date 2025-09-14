package hashring

import (
	"testing"
)

func TestNewConsistentHashRing(t *testing.T) {
	ring := NewConsistentHashRing(3)

	if ring == nil {
		t.Fatal("Expected non-nil hash ring")
	}

	// Type assertion para acessar campos internos para teste
	concreteRing := ring.(*ConsistentHashRing)
	if concreteRing.NumReplicas != 3 {
		t.Errorf("Expected NumReplicas to be 3, got %d", concreteRing.NumReplicas)
	}

	if len(concreteRing.Nodes) != 0 {
		t.Errorf("Expected empty nodes slice, got %d nodes", len(concreteRing.Nodes))
	}
}

func TestAddNode(t *testing.T) {
	ring := NewConsistentHashRing(3)

	ring.AddNode("shard01")

	concreteRing := ring.(*ConsistentHashRing)
	if len(concreteRing.Nodes) != 3 {
		t.Errorf("Expected 3 nodes after adding one shard, got %d", len(concreteRing.Nodes))
	}

	// Verificar se todos os nós têm o mesmo ID mas hashes diferentes
	for i, node := range concreteRing.Nodes {
		if node.ID != "shard01" {
			t.Errorf("Expected node %d ID to be 'shard01', got '%s'", i, node.ID)
		}
	}

	// Verificar se os nós estão ordenados por hash
	for i := 1; i < len(concreteRing.Nodes); i++ {
		if concreteRing.Nodes[i-1].Hash >= concreteRing.Nodes[i].Hash {
			t.Error("Nodes should be sorted by hash")
		}
	}
}

func TestGetNode(t *testing.T) {
	ring := NewConsistentHashRing(3)

	// Teste com ring vazio
	node := ring.GetNode("test-key")
	if node != "" {
		t.Errorf("Expected empty string for empty ring, got '%s'", node)
	}

	// Adicionar alguns shards
	ring.AddNode("shard01")
	ring.AddNode("shard02")
	ring.AddNode("shard03")

	// Teste consistência - a mesma chave deve sempre retornar o mesmo shard
	key := "user123"
	firstResult := ring.GetNode(key)

	for i := 0; i < 10; i++ {
		result := ring.GetNode(key)
		if result != firstResult {
			t.Errorf("GetNode should be deterministic. First result: %s, iteration %d result: %s", firstResult, i, result)
		}
	}

	// Teste com diferentes chaves
	testCases := []string{"user123", "user456", "user789", "admin", "guest"}
	for _, testKey := range testCases {
		result := ring.GetNode(testKey)
		if result == "" {
			t.Errorf("GetNode should return a valid shard for key '%s'", testKey)
		}
	}
}

func TestHashKey(t *testing.T) {
	// Teste se a função hash é determinística
	key := "test-key"
	hash1 := hashKey(key)
	hash2 := hashKey(key)

	if hash1 != hash2 {
		t.Error("hashKey should be deterministic")
	}

	// Teste se diferentes chaves produzem hashes diferentes (na maioria dos casos)
	hash3 := hashKey("different-key")
	if hash1 == hash3 {
		t.Error("Different keys should generally produce different hashes")
	}

	// Teste case-insensitive
	hash4 := hashKey("TEST-KEY")
	hash5 := hashKey("test-key")
	if hash4 != hash5 {
		t.Error("hashKey should be case-insensitive")
	}
}

func TestDistribution(t *testing.T) {
	ring := NewConsistentHashRing(3)

	// Adicionar 3 shards
	shards := []string{"shard01", "shard02", "shard03"}
	for _, shard := range shards {
		ring.AddNode(shard)
	}

	// Testar distribuição com muitas chaves
	distribution := make(map[string]int)
	numKeys := 1000

	for i := 0; i < numKeys; i++ {
		key := "user" + string(rune(i))
		shard := ring.GetNode(key)
		distribution[shard]++
	}

	// Verificar se todos os shards receberam pelo menos algumas chaves
	for _, shard := range shards {
		count := distribution[shard]
		if count == 0 {
			t.Errorf("Shard %s received no keys, poor distribution", shard)
		}

		// Verificar se a distribuição não está muito desbalanceada
		// (cada shard deveria receber aproximadamente 1/3 das chaves)
		expectedRatio := float64(numKeys) / float64(len(shards))
		actualRatio := float64(count)

		// Aceitar uma variação de 50% para ser realista
		minExpected := expectedRatio * 0.5
		maxExpected := expectedRatio * 1.5

		if actualRatio < minExpected || actualRatio > maxExpected {
			t.Logf("Warning: Shard %s has %d keys (%.2f%%), expected around %.0f keys",
				shard, count, (actualRatio/float64(numKeys))*100, expectedRatio)
		}
	}

	t.Logf("Distribution: %v", distribution)
}
