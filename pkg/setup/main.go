package setup

import (
	"app/pkg/interfaces"
	"app/pkg/sharding"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
)

// ConfigManagerImpl implementa a interface ConfigManager
type ConfigManagerImpl struct {
	shards      []interfaces.Shard
	shardingKey string
	once        sync.Once
}

// Garantir que ConfigManagerImpl implementa a interface ConfigManager
var _ interfaces.ConfigManager = (*ConfigManagerImpl)(nil)

// NewConfigManager cria uma nova instância de ConfigManager
func NewConfigManager() interfaces.ConfigManager {
	return &ConfigManagerImpl{}
}

func (cm *ConfigManagerImpl) LoadShards() ([]interfaces.Shard, error) {
	var err error
	cm.once.Do(func() {
		cm.shards, err = cm.discoverShards()
	})
	return cm.shards, err
}

func (cm *ConfigManagerImpl) GetShardingKey() string {
	if cm.shardingKey == "" {
		cm.shardingKey = os.Getenv("SHARDING_KEY")
	}
	return cm.shardingKey
}

func (cm *ConfigManagerImpl) discoverShards() ([]interfaces.Shard, error) {
	var shards []interfaces.Shard

	// Verifica se existem variáveis com o padrão SHARD_*_URL
	pattern := regexp.MustCompile(`^SHARD_(\d+)_URL$`)

	for _, env := range os.Environ() {
		// Split the environment variable into key and value
		pair := splitEnv(env)
		key := pair[0]

		// Check if the key matches the pattern
		if matches := pattern.FindStringSubmatch(key); matches != nil {
			shardID, err := strconv.Atoi(matches[1])
			if err != nil {
				return nil, fmt.Errorf("invalid shard ID %s: %v", matches[1], err)
			}

			shard := interfaces.Shard{
				ID:   shardID,
				Name: "SHARD_" + matches[1],
				URL:  pair[1],
			}
			shards = append(shards, shard)
			fmt.Printf("Mapping shard %v on host: %s\n", shard.ID, shard.URL)
		}
	}

	if len(shards) == 0 {
		return nil, fmt.Errorf("no shards found. Please set SHARD_*_URL environment variables")
	}

	return shards, nil
}

func splitEnv(env string) []string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env, ""}
}

// Init inicializa o sistema com as configurações descobertas
// Esta função mantém compatibilidade com o código existente
func Init() error {
	return InitWithRouter(nil)
}

// InitWithRouter permite injeção de dependência do ShardRouter
func InitWithRouter(router interfaces.ShardRouter) error {
	configManager := NewConfigManager()

	shardingKey := configManager.GetShardingKey()
	if shardingKey == "" {
		return fmt.Errorf("SHARDING_KEY not set")
	}

	shards, err := configManager.LoadShards()
	if err != nil {
		return err
	}

	// Se não foi fornecido um router, criar um novo
	if router == nil {
		router = sharding.NewShardRouter(shardingKey)
	}

	// Setup Hash Ring
	fmt.Printf("Setting up Hash Ring with %v nodes\n", len(shards))
	router.InitHashRing(len(shards))

	for _, shard := range shards {
		router.AddShard(shard.URL)
	}

	return nil
}
