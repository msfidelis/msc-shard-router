package sharding

import (
	"app/pkg/hashring"
	"fmt"
	"net/http"
	"os"
)

var hashRing *hashring.ConsistentHashRing

func InitHashRing(size int) {
	hashRing = hashring.NewConsistentHashRing(size)
}

func AddShard(shardHost string) {
	fmt.Println("Adding shard to hash ring: ", shardHost)
	hashRing.AddNode(shardHost)
}

func GetShardingKey(r *http.Request) string {
	shardingKey := os.Getenv("SHARDING_KEY")
	key := r.Header.Get(shardingKey)
	return key
}

func GetShardHost(key string) string {
	node := hashRing.GetNode(key)
	fmt.Printf("Mapping sharding key %s to host: %s\n", key, node)
	return node
}
