package main

import (
	"bufio"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/csv"
	"flag"
	"fmt"
	"hash"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spaolacci/murmur3"
)

// HashRing representa um anel de hash consistente
type HashRing struct {
	nodes    map[uint64]string
	sortKeys []uint64
	replicas int
	hashFunc func([]byte) uint64
}

// AnalysisResult armazena resultado da análise por algoritmo e shards
type AnalysisResult struct {
	Algorithm     string
	NumShards     int
	CoefVariation float64
	StdDev        float64
	Mean          float64
	MinPercent    float64
	MaxPercent    float64
	Delta         int
	DeltaPercent  float64
	TotalKeys     int
}

func main() {
	// Parse de flags
	replicasFlag := flag.Int("replicas", 200, "Número de réplicas virtuais por shard")
	minShards := flag.Int("min", 3, "Número mínimo de shards")
	maxShards := flag.Int("max", 200, "Número máximo de shards")
	outputFile := flag.String("output", "cv_analysis.csv", "Arquivo de saída CSV")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Uso: go run main.go <arquivo_uuids.txt> [flags]")
	}

	uuidFile := flag.Arg(0)

	// Carregar UUIDs
	fmt.Printf("🔬 Análise de Coeficiente de Variação vs Número de Shards\n")
	fmt.Printf("📁 Arquivo: %s\n", uuidFile)
	fmt.Printf("📊 Range: %d a %d shards\n", *minShards, *maxShards)
	fmt.Printf("🔄 Réplicas: %d por shard\n\n", *replicasFlag)

	uuids, err := loadUUIDs(uuidFile)
	if err != nil {
		log.Fatalf("Erro ao carregar UUIDs: %v", err)
	}
	fmt.Printf("✓ %d UUIDs carregados\n\n", len(uuids))

	// Algoritmos a testar
	algorithms := []struct {
		name     string
		hashFunc func([]byte) uint64
	}{
		{"MURMUR3", hashKeyMurmur},
		{"SHA-1", hashKeySHA1},
		{"SHA-256", hashKeySHA256},
		{"SHA-512", hashKeySHA512},
	}

	// Coletar resultados
	var results []AnalysisResult
	totalTests := len(algorithms) * (*maxShards - *minShards + 1)
	currentTest := 0

	fmt.Println("🚀 Executando análise...")
	startTime := time.Now()

	for _, algo := range algorithms {
		for numShards := *minShards; numShards <= *maxShards; numShards++ {
			currentTest++
			fmt.Printf("\r⏳ Progresso: %d/%d [%s - %d shards]          ",
				currentTest, totalTests, algo.name, numShards)

			result := analyzeDistribution(uuids, numShards, *replicasFlag, algo.name, algo.hashFunc)
			results = append(results, result)
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\n✅ Análise concluída em %v\n\n", duration)

	// Gerar tabela formatada
	printFormattedTable(results, *minShards, *maxShards)

	// Salvar CSV
	err = saveCSV(results, *outputFile)
	if err != nil {
		log.Fatalf("Erro ao salvar CSV: %v", err)
	}

	fmt.Printf("\n💾 Dados salvos em: %s\n", *outputFile)
	fmt.Println("✨ Análise finalizada!")
}

// loadUUIDs carrega UUIDs do arquivo
func loadUUIDs(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var uuids []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		uuid := strings.TrimSpace(scanner.Text())
		if uuid != "" {
			uuids = append(uuids, uuid)
		}
	}

	return uuids, scanner.Err()
}

// analyzeDistribution analisa distribuição para configuração específica
func analyzeDistribution(uuids []string, numShards, replicas int, algorithm string, hashFunc func([]byte) uint64) AnalysisResult {
	// Criar hash ring
	ring := NewHashRing(replicas, hashFunc)
	for i := 1; i <= numShards; i++ {
		shardName := fmt.Sprintf("shard%02d", i)
		ring.AddNode(shardName)
	}

	// Distribuir UUIDs
	distribution := make(map[string]int)
	for _, uuid := range uuids {
		shard := ring.GetNode(uuid)
		distribution[shard]++
	}

	// Calcular estatísticas
	totalKeys := len(uuids)

	var sum, sumSquares float64
	var maxCount, minCount int
	var maxPercent, minPercent float64

	for _, count := range distribution {
		sum += float64(count)
		sumSquares += float64(count) * float64(count)

		if count > maxCount {
			maxCount = count
		}
		if minCount == 0 || count < minCount {
			minCount = count
		}
	}

	mean := sum / float64(numShards)
	variance := (sumSquares / float64(numShards)) - (mean * mean)
	stdDev := math.Sqrt(variance)
	coefVariation := (stdDev / mean) * 100.0

	maxPercent = (float64(maxCount) / float64(totalKeys)) * 100.0
	minPercent = (float64(minCount) / float64(totalKeys)) * 100.0
	delta := maxCount - minCount
	deltaPercent := maxPercent - minPercent

	return AnalysisResult{
		Algorithm:     algorithm,
		NumShards:     numShards,
		CoefVariation: coefVariation,
		StdDev:        stdDev,
		Mean:          mean,
		MinPercent:    minPercent,
		MaxPercent:    maxPercent,
		Delta:         delta,
		DeltaPercent:  deltaPercent,
		TotalKeys:     totalKeys,
	}
}

// printFormattedTable imprime tabela formatada por algoritmo
func printFormattedTable(results []AnalysisResult, minShards, maxShards int) {
	algorithms := []string{"MURMUR3", "SHA-1", "SHA-256", "SHA-512"}

	fmt.Printf("%s\n", strings.Repeat("=", 120))
	fmt.Println("ANÁLISE: COEFICIENTE DE VARIAÇÃO vs NÚMERO DE SHARDS")
	fmt.Printf("%s\n\n", strings.Repeat("=", 120))

	for _, algo := range algorithms {
		fmt.Printf("┌─ %s %s\n", algo, strings.Repeat("─", 110))
		fmt.Printf("│\n")
		fmt.Printf("│  Shards | CV%%      | Desvio Padrão | Média       | Delta (Amp.)     | Min%%     | Max%%     \n")
		fmt.Printf("│  %s\n", strings.Repeat("─", 105))

		for numShards := minShards; numShards <= maxShards; numShards++ {
			for _, r := range results {
				if r.Algorithm == algo && r.NumShards == numShards {
					fmt.Printf("│  %-6d | %7.2f | %13.2f | %11.2f | %7d (%5.2f%%) | %7.2f | %7.2f\n",
						r.NumShards, r.CoefVariation, r.StdDev, r.Mean,
						r.Delta, r.DeltaPercent, r.MinPercent, r.MaxPercent)
					break
				}
			}
		}
		fmt.Printf("│\n")
		fmt.Printf("└%s\n\n", strings.Repeat("─", 115))
	}

	// Tabela comparativa
	fmt.Printf("%s\n", strings.Repeat("=", 120))
	fmt.Println("TABELA COMPARATIVA: CV% POR ALGORITMO E NÚMERO DE SHARDS")
	fmt.Printf("%s\n\n", strings.Repeat("=", 120))

	// Cabeçalho
	fmt.Printf("%-8s |", "Shards")
	for _, algo := range algorithms {
		fmt.Printf(" %-12s |", algo)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 70))

	// Dados
	for numShards := minShards; numShards <= maxShards; numShards++ {
		fmt.Printf("%-8d |", numShards)
		for _, algo := range algorithms {
			for _, r := range results {
				if r.Algorithm == algo && r.NumShards == numShards {
					fmt.Printf(" %10.2f%% |", r.CoefVariation)
					break
				}
			}
		}
		fmt.Println()
	}

	fmt.Println()
}

// saveCSV salva resultados em arquivo CSV
func saveCSV(results []AnalysisResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Cabeçalho
	headers := []string{
		"Algorithm", "NumShards", "CV%", "StdDev", "Mean",
		"MinPercent", "MaxPercent", "Delta", "DeltaPercent", "TotalKeys",
	}
	writer.Write(headers)

	// Dados
	for _, r := range results {
		row := []string{
			r.Algorithm,
			strconv.Itoa(r.NumShards),
			fmt.Sprintf("%.4f", r.CoefVariation),
			fmt.Sprintf("%.2f", r.StdDev),
			fmt.Sprintf("%.2f", r.Mean),
			fmt.Sprintf("%.4f", r.MinPercent),
			fmt.Sprintf("%.4f", r.MaxPercent),
			strconv.Itoa(r.Delta),
			fmt.Sprintf("%.4f", r.DeltaPercent),
			strconv.Itoa(r.TotalKeys),
		}
		writer.Write(row)
	}

	return nil
}

// NewHashRing cria um novo hash ring
func NewHashRing(replicas int, hashFunc func([]byte) uint64) *HashRing {
	return &HashRing{
		nodes:    make(map[uint64]string),
		replicas: replicas,
		hashFunc: hashFunc,
	}
}

// AddNode adiciona um nó ao hash ring
func (h *HashRing) AddNode(node string) {
	for i := 0; i < h.replicas; i++ {
		virtualKey := fmt.Sprintf("%s:%d", node, i)
		hash := h.hashFunc([]byte(virtualKey))
		h.nodes[hash] = node
		h.sortKeys = append(h.sortKeys, hash)
	}
	sort.Slice(h.sortKeys, func(i, j int) bool {
		return h.sortKeys[i] < h.sortKeys[j]
	})
}

// GetNode retorna o nó responsável por uma chave
func (h *HashRing) GetNode(key string) string {
	if len(h.sortKeys) == 0 {
		return ""
	}

	hash := h.hashFunc([]byte(key))
	idx := sort.Search(len(h.sortKeys), func(i int) bool {
		return h.sortKeys[i] >= hash
	})

	if idx == len(h.sortKeys) {
		idx = 0
	}

	return h.nodes[h.sortKeys[idx]]
}

// Funções de hash
func hashKeyMurmur(key []byte) uint64 {
	return murmur3.Sum64(key)
}

func hashKeySHA1(key []byte) uint64 {
	h := sha1.New()
	h.Write(key)
	return hashToUint64(h)
}

func hashKeySHA256(key []byte) uint64 {
	h := sha256.New()
	h.Write(key)
	return hashToUint64(h)
}

func hashKeySHA512(key []byte) uint64 {
	h := sha512.New()
	h.Write(key)
	return hashToUint64(h)
}

func hashToUint64(h hash.Hash) uint64 {
	sum := h.Sum(nil)
	return uint64(sum[0])<<56 | uint64(sum[1])<<48 | uint64(sum[2])<<40 | uint64(sum[3])<<32 |
		uint64(sum[4])<<24 | uint64(sum[5])<<16 | uint64(sum[6])<<8 | uint64(sum[7])
}
