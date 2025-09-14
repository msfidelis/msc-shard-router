package setup

import (
	"app/pkg/sharding"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
)

type Shard struct {
	Id   int
	Name string
	Url  string
}

type Config struct {
	Shards []Shard
}

var config Config
var once sync.Once

func Init() {

	shardingKey := os.Getenv("SHARDING_KEY")
	if shardingKey == "" {
		panic("SHARDING_KEY not set")
	}

	// Verifica se existem variáveis com o padrão SHARD_*_URL
	pattern := regexp.MustCompile(`^SHARD_(\d+)_URL$`)

	for _, env := range os.Environ() {
		// Split the environment variable into key and value
		pair := splitEnv(env)
		key := pair[0]

		// Check if the key matches the pattern
		if matches := pattern.FindStringSubmatch(key); matches != nil {
			shardID, _ := strconv.Atoi(matches[1])
			shard := Shard{
				Id:   shardID,
				Name: "SHARD_" + matches[1],
				Url:  pair[1],
			}
			config.Shards = append(config.Shards, shard)
			fmt.Printf("Mapping shard %v on host: %s\n", shard.Id, shard.Url)
		}
	}

	// Setup Hash Ring
	fmt.Printf("Setting up Hash Ring with %v nodes\n", len(config.Shards))
	sharding.InitHashRing(len(config.Shards))

	for _, shard := range config.Shards {
		sharding.AddShard(shard.Url)
	}

}

func splitEnv(env string) []string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env, ""}
}
