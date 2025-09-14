package sharding

import (
	"app/pkg/hashring"
	"app/pkg/interfaces"
	"fmt"
	"net/http"
	"os"
)

// ShardRouterImpl implementa a interface ShardRouter
type ShardRouterImpl struct {
	hashRing    interfaces.HashRing
	shardingKey string
}

// Garantir que ShardRouterImpl implementa a interface ShardRouter
var _ interfaces.ShardRouter = (*ShardRouterImpl)(nil)

// NewShardRouter cria uma nova instância de ShardRouter
func NewShardRouter(shardingKey string) interfaces.ShardRouter {
	return &ShardRouterImpl{
		shardingKey: shardingKey,
	}
}

func (sr *ShardRouterImpl) InitHashRing(size int) {
	if sr.hashRing == nil {
		// Importar a função de criação do hashring
		sr.hashRing = createHashRing(size)
	}
}

func (sr *ShardRouterImpl) AddShard(shardHost string) {
	if sr.hashRing == nil {
		panic("Hash ring not initialized. Call InitHashRing first.")
	}
	fmt.Println("Adding shard to hash ring: ", shardHost)
	sr.hashRing.AddNode(shardHost)
}

func (sr *ShardRouterImpl) GetShardingKey(r *http.Request) string {
	if sr.shardingKey == "" {
		// Fallback para variável de ambiente se não foi configurado
		sr.shardingKey = os.Getenv("SHARDING_KEY")
	}
	return r.Header.Get(sr.shardingKey)
}

func (sr *ShardRouterImpl) GetShardHost(key string) string {
	if sr.hashRing == nil {
		panic("Hash ring not initialized. Call InitHashRing first.")
	}
	node := sr.hashRing.GetNode(key)
	fmt.Printf("Mapping sharding key %s to host: %s\n", key, node)
	return node
}

// createHashRing é uma função auxiliar para criar o hash ring
// Isso permite injeção de dependência em testes
func createHashRing(size int) interfaces.HashRing {
	return hashring.NewConsistentHashRing(size)
}
