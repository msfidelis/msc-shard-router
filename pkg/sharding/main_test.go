package sharding

import (
	"net/http"
	"testing"
)

// MockHashRing é um mock da interface HashRing para testes
type MockHashRing struct {
	nodes       map[string]bool
	getNodeFunc func(key string) string
}

func (m *MockHashRing) AddNode(nodeID string) {
	if m.nodes == nil {
		m.nodes = make(map[string]bool)
	}
	m.nodes[nodeID] = true
}

func (m *MockHashRing) GetNode(key string) string {
	if m.getNodeFunc != nil {
		return m.getNodeFunc(key)
	}
	// Comportamento padrão simples para testes
	if len(m.nodes) == 0 {
		return ""
	}
	// Retorna o primeiro nó (simplificado para testes)
	for node := range m.nodes {
		return node
	}
	return ""
}

func TestNewShardRouter(t *testing.T) {
	shardingKey := "user_id"
	router := NewShardRouter(shardingKey)

	if router == nil {
		t.Fatal("Expected non-nil shard router")
	}

	// Type assertion para verificar implementação
	concreteRouter := router.(*ShardRouterImpl)
	if concreteRouter.shardingKey != shardingKey {
		t.Errorf("Expected sharding key to be '%s', got '%s'", shardingKey, concreteRouter.shardingKey)
	}
}

func TestShardRouterImpl_InitHashRing(t *testing.T) {
	router := NewShardRouter("user_id").(*ShardRouterImpl)

	if router.hashRing != nil {
		t.Error("Expected hash ring to be nil initially")
	}

	router.InitHashRing(3)

	if router.hashRing == nil {
		t.Error("Expected hash ring to be initialized")
	}
}

func TestShardRouterImpl_AddShard(t *testing.T) {
	router := NewShardRouter("user_id").(*ShardRouterImpl)
	mockHashRing := &MockHashRing{}
	router.hashRing = mockHashRing

	shardURL := "http://shard01:80"
	router.AddShard(shardURL)

	if !mockHashRing.nodes[shardURL] {
		t.Error("Expected shard to be added to hash ring")
	}
}

func TestShardRouterImpl_AddShard_PanicWithoutInit(t *testing.T) {
	router := NewShardRouter("user_id").(*ShardRouterImpl)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when adding shard without initializing hash ring")
		}
	}()

	router.AddShard("http://shard01:80")
}

func TestShardRouterImpl_GetShardingKey(t *testing.T) {
	tests := []struct {
		name        string
		shardingKey string
		headerValue string
		expected    string
	}{
		{
			name:        "Valid header",
			shardingKey: "user_id",
			headerValue: "123",
			expected:    "123",
		},
		{
			name:        "Empty header",
			shardingKey: "user_id",
			headerValue: "",
			expected:    "",
		},
		{
			name:        "Different header key",
			shardingKey: "client_id",
			headerValue: "abc",
			expected:    "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewShardRouter(tt.shardingKey)

			req, err := http.NewRequest("GET", "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set(tt.shardingKey, tt.headerValue)

			result := router.GetShardingKey(req)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestShardRouterImpl_GetShardHost(t *testing.T) {
	router := NewShardRouter("user_id").(*ShardRouterImpl)

	expectedShard := "http://shard01:80"
	mockHashRing := &MockHashRing{
		getNodeFunc: func(key string) string {
			if key == "test-key" {
				return expectedShard
			}
			return ""
		},
	}
	router.hashRing = mockHashRing

	result := router.GetShardHost("test-key")
	if result != expectedShard {
		t.Errorf("Expected '%s', got '%s'", expectedShard, result)
	}
}

func TestShardRouterImpl_GetShardHost_PanicWithoutInit(t *testing.T) {
	router := NewShardRouter("user_id").(*ShardRouterImpl)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when getting shard host without initializing hash ring")
		}
	}()

	router.GetShardHost("test-key")
}

func TestShardRouterImpl_Integration(t *testing.T) {
	// Teste de integração completo
	router := NewShardRouter("user_id")

	// Inicializar hash ring
	router.InitHashRing(3)

	// Adicionar alguns shards
	shards := []string{
		"http://shard01:80",
		"http://shard02:80",
		"http://shard03:80",
	}

	for _, shard := range shards {
		router.AddShard(shard)
	}

	// Testar roteamento
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	testKey := "user123"
	req.Header.Set("user_id", testKey)

	// Obter chave de sharding
	shardingKey := router.GetShardingKey(req)
	if shardingKey != testKey {
		t.Errorf("Expected sharding key '%s', got '%s'", testKey, shardingKey)
	}

	// Obter shard host
	shardHost := router.GetShardHost(shardingKey)
	if shardHost == "" {
		t.Error("Expected non-empty shard host")
	}

	// Verificar consistência
	for i := 0; i < 5; i++ {
		host := router.GetShardHost(shardingKey)
		if host != shardHost {
			t.Error("Shard routing should be consistent")
		}
	}
}
