package setup

import (
	"app/pkg/interfaces"
	"net/http"
	"os"
	"testing"
)

// MockShardRouter é um mock da interface ShardRouter para testes
type MockShardRouter struct {
	hashRingSize int
	shards       []string
	initCalled   bool
	getNodeFunc  func(key string) string
}

func (m *MockShardRouter) InitHashRing(size int) {
	m.hashRingSize = size
	m.initCalled = true
}

func (m *MockShardRouter) AddShard(shardHost string) {
	m.shards = append(m.shards, shardHost)
}

func (m *MockShardRouter) GetShardingKey(r *http.Request) string {
	// Mock implementation - not needed for setup tests
	return ""
}

func (m *MockShardRouter) GetShardHost(key string) string {
	if m.getNodeFunc != nil {
		return m.getNodeFunc(key)
	}
	return ""
}

// Garantir que MockShardRouter implementa a interface
var _ interfaces.ShardRouter = (*MockShardRouter)(nil)

func TestNewConfigManager(t *testing.T) {
	cm := NewConfigManager()
	if cm == nil {
		t.Fatal("Expected non-nil config manager")
	}
}

func TestConfigManagerImpl_GetShardingKey(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "Valid sharding key",
			envValue: "user_id",
			expected: "user_id",
		},
		{
			name:     "Empty sharding key",
			envValue: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			oldValue := os.Getenv("SHARDING_KEY")
			defer os.Setenv("SHARDING_KEY", oldValue)

			os.Setenv("SHARDING_KEY", tt.envValue)

			cm := NewConfigManager()
			result := cm.GetShardingKey()

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConfigManagerImpl_LoadShards(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		expectCount int
	}{
		{
			name: "Valid shards",
			envVars: map[string]string{
				"SHARD_01_URL": "http://shard01:80",
				"SHARD_02_URL": "http://shard02:80",
				"SHARD_03_URL": "http://shard03:80",
			},
			expectError: false,
			expectCount: 3,
		},
		{
			name: "Single shard",
			envVars: map[string]string{
				"SHARD_01_URL": "http://shard01:80",
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name:        "No shards",
			envVars:     map[string]string{},
			expectError: true,
			expectCount: 0,
		},
		{
			name: "Mixed environment variables",
			envVars: map[string]string{
				"SHARD_01_URL": "http://shard01:80",
				"OTHER_VAR":    "value",
				"SHARD_02_URL": "http://shard02:80",
				"NOT_SHARD":    "http://not-a-shard:80",
			},
			expectError: false,
			expectCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing environment variables
			clearShardEnvVars()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			cm := NewConfigManager()
			shards, err := cm.LoadShards()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(shards) != tt.expectCount {
				t.Errorf("Expected %d shards, got %d", tt.expectCount, len(shards))
			}

			// Verify shard data
			if !tt.expectError && len(shards) > 0 {
				for _, shard := range shards {
					if shard.ID <= 0 {
						t.Errorf("Expected positive shard ID, got %d", shard.ID)
					}
					if shard.Name == "" {
						t.Error("Expected non-empty shard name")
					}
					if shard.URL == "" {
						t.Error("Expected non-empty shard URL")
					}
				}
			}
		})
	}
}

func TestConfigManagerImpl_LoadShards_OnlyOnce(t *testing.T) {
	// Set up environment
	os.Setenv("SHARD_01_URL", "http://shard01:80")
	defer os.Unsetenv("SHARD_01_URL")

	cm := NewConfigManager().(*ConfigManagerImpl)

	// First call
	shards1, err1 := cm.LoadShards()
	if err1 != nil {
		t.Fatalf("Unexpected error: %v", err1)
	}

	// Change environment (should not affect second call due to sync.Once)
	os.Setenv("SHARD_02_URL", "http://shard02:80")
	defer os.Unsetenv("SHARD_02_URL")

	// Second call
	shards2, err2 := cm.LoadShards()
	if err2 != nil {
		t.Fatalf("Unexpected error: %v", err2)
	}

	// Should return the same result
	if len(shards1) != len(shards2) {
		t.Error("LoadShards should return the same result on multiple calls")
	}

	if len(shards2) != 1 {
		t.Errorf("Expected 1 shard (from first call), got %d", len(shards2))
	}
}

func TestSplitEnv(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Valid key=value",
			input:    "KEY=value",
			expected: []string{"KEY", "value"},
		},
		{
			name:     "Empty value",
			input:    "KEY=",
			expected: []string{"KEY", ""},
		},
		{
			name:     "Value with equals",
			input:    "KEY=value=with=equals",
			expected: []string{"KEY", "value=with=equals"},
		},
		{
			name:     "No equals sign",
			input:    "KEYVALUE",
			expected: []string{"KEYVALUE", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitEnv(tt.input)

			if len(result) != 2 {
				t.Errorf("Expected 2 elements, got %d", len(result))
				return
			}

			if result[0] != tt.expected[0] || result[1] != tt.expected[1] {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestInit(t *testing.T) {
	// Set up environment
	os.Setenv("SHARDING_KEY", "user_id")
	os.Setenv("SHARD_01_URL", "http://shard01:80")
	defer func() {
		os.Unsetenv("SHARDING_KEY")
		os.Unsetenv("SHARD_01_URL")
	}()

	err := Init()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestInitWithRouter(t *testing.T) {
	// Set up environment
	os.Setenv("SHARDING_KEY", "user_id")
	os.Setenv("SHARD_01_URL", "http://shard01:80")
	os.Setenv("SHARD_02_URL", "http://shard02:80")
	defer func() {
		os.Unsetenv("SHARDING_KEY")
		os.Unsetenv("SHARD_01_URL")
		os.Unsetenv("SHARD_02_URL")
	}()

	mockRouter := &MockShardRouter{}
	err := InitWithRouter(mockRouter)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !mockRouter.initCalled {
		t.Error("Expected InitHashRing to be called")
	}

	if mockRouter.hashRingSize != 2 {
		t.Errorf("Expected hash ring size 2, got %d", mockRouter.hashRingSize)
	}

	if len(mockRouter.shards) != 2 {
		t.Errorf("Expected 2 shards added, got %d", len(mockRouter.shards))
	}

	expectedShards := []string{"http://shard01:80", "http://shard02:80"}
	for i, expectedShard := range expectedShards {
		found := false
		for _, actualShard := range mockRouter.shards {
			if actualShard == expectedShard {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected shard %s not found in %v", expectedShard, mockRouter.shards)
		}
		_ = i // avoid unused variable error
	}
}

func TestInitWithRouter_NoShardingKey(t *testing.T) {
	// Clear SHARDING_KEY
	oldValue := os.Getenv("SHARDING_KEY")
	os.Unsetenv("SHARDING_KEY")
	defer os.Setenv("SHARDING_KEY", oldValue)

	mockRouter := &MockShardRouter{}
	err := InitWithRouter(mockRouter)

	if err == nil {
		t.Error("Expected error when SHARDING_KEY is not set")
	}
}

// Helper function to clear shard environment variables
func clearShardEnvVars() {
	for _, env := range os.Environ() {
		pair := splitEnv(env)
		key := pair[0]
		if len(key) > 5 && key[:5] == "SHARD" && key[len(key)-4:] == "_URL" {
			os.Unsetenv(key)
		}
	}
}
