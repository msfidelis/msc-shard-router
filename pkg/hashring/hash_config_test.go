package hashring

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// TestHashAlgorithmConfiguration testa a configuração dos algoritmos via variável de ambiente
func TestHashAlgorithmConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expectedLog string
	}{
		{
			name:        "SHA512",
			envValue:    "SHA512",
			expectedLog: "Hash algorithm configured: SHA512",
		},
		{
			name:        "SHA256",
			envValue:    "SHA256",
			expectedLog: "Hash algorithm configured: SHA256",
		},
		{
			name:        "SHA1",
			envValue:    "SHA1",
			expectedLog: "Hash algorithm configured: SHA1",
		},
		{
			name:        "MD5",
			envValue:    "MD5",
			expectedLog: "Hash algorithm configured: MD5",
		},
		{
			name:        "MURMUR",
			envValue:    "MURMUR",
			expectedLog: "Hash algorithm configured: MURMUR",
		},
		{
			name:        "Default (empty)",
			envValue:    "",
			expectedLog: "No HASHING_ALGORITHM specified, defaulting to SHA512",
		},
		{
			name:        "Invalid algorithm",
			envValue:    "INVALID",
			expectedLog: "Unknown hash algorithm 'INVALID', defaulting to SHA512",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Limpar e definir variável de ambiente
			os.Unsetenv("HASHING_ALGORITHM")
			if tt.envValue != "" {
				os.Setenv("HASHING_ALGORITHM", tt.envValue)
			}

			// Criar hash ring para testar configuração
			ring := NewConsistentHashRing(3)

			// Verificar se o ring foi criado com sucesso
			if ring == nil {
				t.Fatalf("Failed to create hash ring")
			}

			// Adicionar um nó para testar funcionalidade
			ring.AddNode("test-shard")

			// Testar se consegue buscar um nó
			node := ring.GetNode("test-key")
			if node != "test-shard" {
				t.Errorf("Expected node 'test-shard', got '%s'", node)
			}
		})
	}

	// Limpar variável de ambiente após os testes
	os.Unsetenv("HASHING_ALGORITHM")
}

// TestHashAlgorithmDistribution testa se diferentes algoritmos produzem distribuições diferentes
func TestHashAlgorithmDistribution(t *testing.T) {
	algorithms := []string{"SHA512", "SHA256", "MD5", "SHA1", "MURMUR"}
	testKey := "test-key-123"

	results := make(map[string]string)

	for _, algo := range algorithms {
		os.Setenv("HASHING_ALGORITHM", algo)

		ring := NewConsistentHashRing(3)
		ring.AddNode("shard01")
		ring.AddNode("shard02")
		ring.AddNode("shard03")

		node := ring.GetNode(testKey)
		results[algo] = node

		t.Logf("Algorithm %s: key '%s' mapped to '%s'", algo, testKey, node)
	}

	// Verificar se pelo menos alguns algoritmos produzem resultados diferentes
	// (não é garantido, mas estatisticamente provável)
	uniqueResults := make(map[string]bool)
	for _, result := range results {
		uniqueResults[result] = true
	}

	if len(uniqueResults) == 1 {
		t.Logf("Warning: All algorithms mapped to the same shard for this test key")
	}

	// Limpar
	os.Unsetenv("HASHING_ALGORITHM")
}

// TestHashAlgorithmPerformance faz um benchmark básico dos algoritmos
func TestHashAlgorithmPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	algorithms := []string{"SHA512", "SHA256", "SHA1", "MD5", "MURMUR"}
	testKeys := make([]string, 1000)

	// Gerar chaves de teste
	for i := 0; i < 1000; i++ {
		testKeys[i] = fmt.Sprintf("test-key-%d", i)
	}

	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			os.Setenv("HASHING_ALGORITHM", algo)

			ring := NewConsistentHashRing(10)
			ring.AddNode("shard01")
			ring.AddNode("shard02")
			ring.AddNode("shard03")

			// Medir tempo de distribuição
			start := time.Now()
			distribution := make(map[string]int)

			for _, key := range testKeys {
				node := ring.GetNode(key)
				distribution[node]++
			}

			elapsed := time.Since(start)

			t.Logf("Algorithm %s: %d lookups in %v (%.2f μs/lookup)",
				algo, len(testKeys), elapsed, float64(elapsed.Nanoseconds())/float64(len(testKeys))/1000.0)

			// Verificar distribuição básica
			for shard, count := range distribution {
				percentage := float64(count) / float64(len(testKeys)) * 100.0
				t.Logf("  %s: %d keys (%.1f%%)", shard, count, percentage)
			}
		})
	}

	os.Unsetenv("HASHING_ALGORITHM")
}

// Benchmark para comparar performance dos algoritmos
func BenchmarkHashAlgorithms(b *testing.B) {
	algorithms := []string{"SHA512", "SHA256", "SHA1", "MD5", "MURMUR"}

	for _, algo := range algorithms {
		b.Run(algo, func(b *testing.B) {
			os.Setenv("HASHING_ALGORITHM", algo)

			ring := NewConsistentHashRing(10)
			ring.AddNode("shard01")
			ring.AddNode("shard02")
			ring.AddNode("shard03")

			testKey := "benchmark-key-test"

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ring.GetNode(testKey)
			}
		})
	}

	os.Unsetenv("HASHING_ALGORITHM")
}
