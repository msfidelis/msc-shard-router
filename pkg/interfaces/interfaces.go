package interfaces

import "net/http"

// HashRing define a interface para operações de hash consistente
type HashRing interface {
	AddNode(nodeID string)
	GetNode(key string) string
}

// ShardRouter define a interface para roteamento de shards
type ShardRouter interface {
	GetShardingKey(r *http.Request) string
	GetShardHost(key string) string
	InitHashRing(size int)
	AddShard(shardHost string)
}

// ConfigManager define a interface para gerenciamento de configuração
type ConfigManager interface {
	LoadShards() ([]Shard, error)
	GetShardingKey() string
}

// Shard representa um shard no sistema
type Shard struct {
	ID   int
	Name string
	URL  string
}

// ProxyHandler define a interface para o handler de proxy
type ProxyHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// MetricsRecorder define a interface para registro de métricas
type MetricsRecorder interface {
	RecordRequest(shard string)
	RecordResponse(shard string, statusCode int)
}
