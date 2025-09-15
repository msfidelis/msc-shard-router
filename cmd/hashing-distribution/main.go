package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spaolacci/murmur3"
)

// HashFunction representa uma função de hash
type HashFunction struct {
	Name string
	Func func(string) uint64
}

// Node representa um nó no hash ring
type Node struct {
	ID   string
	Hash uint64
}

// DistributionResult armazena os resultados da distribuição
type DistributionResult struct {
	Algorithm    string
	Distribution map[string]int
	TotalKeys    int
	StdDev       float64
	MaxDev       float64
	MinDev       float64
	Variance     float64
	Quality      string
	Duration     time.Duration
}

// Implementações das funções de hash
func hashKeyMD5(s string) uint64 {
	s = strings.ToLower(s)
	hasher := md5.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	return binary.BigEndian.Uint64(hashBytes[:8])
}

func hashKeySHA1(s string) uint64 {
	s = strings.ToLower(s)
	hasher := sha1.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	return binary.BigEndian.Uint64(hashBytes[:8])
}

func hashKeySHA256(s string) uint64 {
	s = strings.ToLower(s)
	hasher := sha256.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	return binary.BigEndian.Uint64(hashBytes[:8])
}

func hashKeySHA512(s string) uint64 {
	s = strings.ToLower(s)
	hasher := sha512.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	return binary.BigEndian.Uint64(hashBytes[:8])
}

func hashKeyMurmur(s string) uint64 {
	s = strings.ToLower(s)
	return murmur3.Sum64([]byte(s))
}

// createHashRing cria um hash ring com 3 shards usando a função de hash especificada
func createHashRing(hashFunc func(string) uint64, numReplicas int) []Node {
	shards := []string{"shard01", "shard02", "shard03"}
	var nodes []Node

	for _, shard := range shards {
		for i := 0; i < numReplicas; i++ {
			replicaID := fmt.Sprintf("%s-%d", shard, i)
			hash := hashFunc(replicaID)
			nodes = append(nodes, Node{ID: shard, Hash: hash})
		}
	}

	// Ordenar nós por hash
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Hash < nodes[j].Hash
	})

	return nodes
}

// getShardForKey encontra o shard para uma chave usando busca binária
func getShardForKey(nodes []Node, key string, hashFunc func(string) uint64) string {
	if len(nodes) == 0 {
		return ""
	}

	hash := hashFunc(key)
	idx := sort.Search(len(nodes), func(i int) bool {
		return nodes[i].Hash >= hash
	})

	if idx == len(nodes) {
		idx = 0
	}

	return nodes[idx].ID
}

// analyzeDistribution analisa a distribuição das chaves usando um algoritmo específico
func analyzeDistribution(keys []string, hashFunc HashFunction, numReplicas int) DistributionResult {
	start := time.Now()

	// Criar hash ring
	nodes := createHashRing(hashFunc.Func, numReplicas)

	// Distribuir chaves
	distribution := make(map[string]int)
	for _, key := range keys {
		shard := getShardForKey(nodes, key, hashFunc.Func)
		distribution[shard]++
	}

	duration := time.Since(start)

	// Calcular estatísticas
	totalKeys := len(keys)
	expected := float64(totalKeys) / 3.0 // 3 shards

	var deviations []float64
	var maxDev, minDev float64 = 0, math.MaxFloat64
	var sumSquaredDev float64

	for _, count := range distribution {
		deviation := math.Abs(float64(count) - expected)
		deviations = append(deviations, deviation)
		sumSquaredDev += deviation * deviation

		if deviation > maxDev {
			maxDev = deviation
		}
		if deviation < minDev {
			minDev = deviation
		}
	}

	// Calcular desvio padrão e variância
	variance := sumSquaredDev / 3.0
	stdDev := math.Sqrt(variance)

	// Determinar qualidade
	avgDev := (maxDev + minDev + deviations[1]) / 3.0
	var quality string
	deviationPercent := avgDev / expected * 100.0

	if deviationPercent <= 5.0 {
		quality = "EXCELENTE"
	} else if deviationPercent <= 10.0 {
		quality = "MUITO BOA"
	} else if deviationPercent <= 15.0 {
		quality = "BOA"
	} else if deviationPercent <= 25.0 {
		quality = "REGULAR"
	} else {
		quality = "RUIM"
	}

	return DistributionResult{
		Algorithm:    hashFunc.Name,
		Distribution: distribution,
		TotalKeys:    totalKeys,
		StdDev:       stdDev,
		MaxDev:       maxDev,
		MinDev:       minDev,
		Variance:     variance,
		Quality:      quality,
		Duration:     duration,
	}
}

// readKeysFromFile lê todas as chaves do arquivo
func readKeysFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %v", err)
	}
	defer file.Close()

	var keys []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			keys = append(keys, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %v", err)
	}

	return keys, nil
}

// printResults exibe os resultados em formato simples
func printResults(results []DistributionResult) {
	if len(results) == 0 {
		return
	}

	totalKeys := results[0].TotalKeys
	expected := float64(totalKeys) / 3.0

	for _, result := range results {
		fmt.Printf("\n%s\n", result.Algorithm)

		// Distribuição por shard
		shards := []string{"shard01", "shard02", "shard03"}
		for _, shard := range shards {
			count := result.Distribution[shard]
			percentage := float64(count) / float64(totalKeys) * 100.0
			deviation := math.Abs(float64(count) - expected)

			fmt.Printf("  %-8s: %d chaves (%5.1f%%) - desvio: %.1f\n",
				shard, count, percentage, deviation)
		}

		fmt.Printf("  Estatísticas:\n")
		fmt.Printf("    Desvio padrão: %.1f\n", result.StdDev)
		fmt.Printf("    Variância: %.1f\n", result.Variance)
		fmt.Printf("    Performance: %v\n", result.Duration)
		fmt.Printf("    Qualidade: %s\n", result.Quality)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Uso: %s <caminho-para-arquivo-de-chaves>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]

	// Verificar se arquivo existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatalf("Arquivo não encontrado: %s", filename)
	}

	// Ler chaves do arquivo
	keys, err := readKeysFromFile(filename)
	if err != nil {
		log.Fatalf("Erro ao ler arquivo: %v", err)
	}

	if len(keys) == 0 {
		log.Fatalf("Nenhuma chave encontrada no arquivo")
	}

	// Definir funções de hash disponíveis
	hashFunctions := []HashFunction{
		{"SHA-512", hashKeySHA512},
		{"SHA-256", hashKeySHA256},
		{"SHA-1", hashKeySHA1},
		{"MD5", hashKeyMD5},
		{"MURMUR", hashKeyMurmur},
	}

	// Analisar cada algoritmo
	var results []DistributionResult
	numReplicas := 10

	for _, hashFunc := range hashFunctions {
		result := analyzeDistribution(keys, hashFunc, numReplicas)
		results = append(results, result)
	}

	// Exibir resultados
	printResults(results)
}
