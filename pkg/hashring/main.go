package hashring

import (
	"app/pkg/interfaces"
	"crypto/sha512"
	"encoding/binary"
	"sort"
	"strconv"
	"strings"
)

type Node struct {
	ID   string
	Hash uint64
}

// ConsistentHashRing representa o hash ring que contém vários nós.
// Implementa a interface interfaces.HashRing
type ConsistentHashRing struct {
	Nodes       []Node
	NumReplicas int
}

// Garantir que ConsistentHashRing implementa a interface HashRing
var _ interfaces.HashRing = (*ConsistentHashRing)(nil)

// NewConsistentHashRing cria um novo anel de hash ring.
func NewConsistentHashRing(numReplicas int) interfaces.HashRing {
	return &ConsistentHashRing{
		Nodes:       []Node{},
		NumReplicas: numReplicas,
	}
}

// AddNode adiciona um nó ao hash ring com múltiplas réplicas virtuais
func (ring *ConsistentHashRing) AddNode(nodeID string) {
	for i := 0; i < ring.NumReplicas; i++ {
		replicaID := nodeID + strconv.Itoa(i)
		hash := hashKey(replicaID)
		ring.Nodes = append(ring.Nodes, Node{ID: nodeID, Hash: hash})
	}
	sort.Slice(ring.Nodes, func(i, j int) bool {
		return ring.Nodes[i].Hash < ring.Nodes[j].Hash
	})
}

// hashKey calcula o hash do tenant e a converte para uint64.
func hashKey(s string) uint64 {
	s = strings.ToLower(s)
	hash := sha512.New()
	hash.Write([]byte(s))
	hashBytes := hash.Sum(nil)
	return binary.BigEndian.Uint64(hashBytes[:8])
}

// GetNode retorna o node onde o Tenant deverá estar alocado
func (ring *ConsistentHashRing) GetNode(key string) string {
	if len(ring.Nodes) == 0 {
		return ""
	}

	hash := hashKey(key)
	idx := sort.Search(len(ring.Nodes), func(i int) bool {
		return ring.Nodes[i].Hash >= hash
	})

	// Se o índice estiver fora dos limites, retorna ao primeiro nó
	if idx == len(ring.Nodes) {
		idx = 0
	}

	return ring.Nodes[idx].ID
}

// Exemplos do artigo: https://fidelissauro.dev/sharding/
