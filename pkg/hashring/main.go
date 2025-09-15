package hashring

import (
	"app/pkg/interfaces"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spaolacci/murmur3"
)

// HashAlgorithm define os algoritmos de hash disponíveis
type HashAlgorithm string

const (
	MD5     HashAlgorithm = "MD5"
	SHA1    HashAlgorithm = "SHA1"
	SHA256  HashAlgorithm = "SHA256"
	SHA512  HashAlgorithm = "SHA512"
	MURMUR3 HashAlgorithm = "MURMUR3"
)

type Node struct {
	ID   string
	Hash uint64
}

// ConsistentHashRing representa o hash ring que contém vários nós.
// Implementa a interface interfaces.HashRing
type ConsistentHashRing struct {
	Nodes         []Node
	NumReplicas   int
	HashAlgorithm string
	hashFunc      func(string) uint64
}

// Garantir que ConsistentHashRing implementa a interface HashRing
var _ interfaces.HashRing = (*ConsistentHashRing)(nil)

// NewConsistentHashRing cria um novo anel de hash ring.
func NewConsistentHashRing(numReplicas int) interfaces.HashRing {
	ring := &ConsistentHashRing{
		Nodes:       []Node{},
		NumReplicas: numReplicas,
	}

	// Configurar algoritmo de hash baseado na variável de ambiente
	ring.configureHashAlgorithm()

	return ring
}

func (ring *ConsistentHashRing) GetHashAlgorithm() string {
	return ring.HashAlgorithm
}

// configureHashAlgorithm configura o algoritmo de hash baseado na variável HASHING_ALGORITHM
func (ring *ConsistentHashRing) configureHashAlgorithm() {
	algorithm := HashAlgorithm(strings.ToUpper(os.Getenv("HASHING_ALGORITHM")))

	switch algorithm {
	case MD5:
		ring.hashFunc = hashKeyMD5
		log.Printf("Hash algorithm configured: MD5")
	case SHA1:
		ring.hashFunc = hashKeySHA1
		log.Printf("Hash algorithm configured: SHA1")
	case SHA256:
		ring.hashFunc = hashKeySHA256
		log.Printf("Hash algorithm configured: SHA256")
	case SHA512:
		ring.hashFunc = hashKeySHA512
		log.Printf("Hash algorithm configured: SHA512")
	case MURMUR3:
		ring.hashFunc = hashKeyMurmur3
		log.Printf("Hash algorithm configured: MURMUR")
	default:
		// Default para SHA512 se não especificado ou inválido
		ring.hashFunc = hashKeySHA512
		if algorithm != "" {
			log.Printf("Unknown hash algorithm '%s', defaulting to SHA512", algorithm)
		} else {
			log.Printf("No HASHING_ALGORITHM specified, defaulting to SHA512")
		}
		algorithm = "SHA512"
	}
	ring.HashAlgorithm = string(algorithm)
}

// AddNode adiciona um nó ao hash ring com múltiplas réplicas virtuais
func (ring *ConsistentHashRing) AddNode(nodeID string) {
	for i := 0; i < ring.NumReplicas; i++ {
		replicaID := nodeID + strconv.Itoa(i)
		hash := ring.hashFunc(replicaID)
		ring.Nodes = append(ring.Nodes, Node{ID: nodeID, Hash: hash})
	}
	sort.Slice(ring.Nodes, func(i, j int) bool {
		return ring.Nodes[i].Hash < ring.Nodes[j].Hash
	})
}

// Implementações dos algoritmos de hash

// hashKeyMD5 calcula hash MD5
func hashKeyMD5(s string) uint64 {
	s = strings.ToLower(s)
	hasher := md5.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	return binary.BigEndian.Uint64(hashBytes[:8])
}

// hashKeySHA1 calcula hash SHA1
func hashKeySHA1(s string) uint64 {
	s = strings.ToLower(s)
	hasher := sha1.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	return binary.BigEndian.Uint64(hashBytes[:8])
}

// hashKeySHA256 calcula hash SHA256
func hashKeySHA256(s string) uint64 {
	s = strings.ToLower(s)
	hasher := sha256.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	return binary.BigEndian.Uint64(hashBytes[:8])
}

// hashKeySHA512 calcula hash SHA512 (função original)
func hashKeySHA512(s string) uint64 {
	s = strings.ToLower(s)
	hasher := sha512.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	return binary.BigEndian.Uint64(hashBytes[:8])
}

// hashKeyMurmur calcula hash Murmur3
func hashKeyMurmur3(s string) uint64 {
	s = strings.ToLower(s)
	return murmur3.Sum64([]byte(s))
}

// GetNode retorna o node onde o Tenant deverá estar alocado
func (ring *ConsistentHashRing) GetNode(key string) string {
	if len(ring.Nodes) == 0 {
		return ""
	}

	hash := ring.hashFunc(key)
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
