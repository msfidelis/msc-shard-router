package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
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
	Algorithm       string
	NumShards       int
	Distribution    map[string]int
	TotalKeys       int
	StdDev          float64
	MaxDev          float64
	MinDev          float64
	Variance        float64
	CoefVariation   float64 // Coeficiente de Variação
	MaxPercent      float64
	MinPercent      float64
	Quality         string
	Duration        time.Duration
	UniformityScore float64 // Score de 0-100
}

// BenchmarkConfig configurações do benchmark
type BenchmarkConfig struct {
	ShardCounts   []int
	NumReplicas   int
	GenerateCSV   bool
	GenerateTable bool
	Verbose       bool
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

// createHashRing cria um hash ring com N shards usando a função de hash especificada
func createHashRing(hashFunc func(string) uint64, numShards, numReplicas int) []Node {
	var nodes []Node

	for i := 1; i <= numShards; i++ {
		shard := fmt.Sprintf("shard%02d", i)
		for j := 0; j < numReplicas; j++ {
			replicaID := fmt.Sprintf("%s-%d", shard, j)
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
func analyzeDistribution(keys []string, hashFunc HashFunction, numShards, numReplicas int) DistributionResult {
	start := time.Now()

	// Criar hash ring
	nodes := createHashRing(hashFunc.Func, numShards, numReplicas)

	// Distribuir chaves
	distribution := make(map[string]int)
	for _, key := range keys {
		shard := getShardForKey(nodes, key, hashFunc.Func)
		distribution[shard]++
	}

	duration := time.Since(start)

	// Calcular estatísticas
	totalKeys := len(keys)
	expected := float64(totalKeys) / float64(numShards)

	var deviations []float64
	var maxDev, minDev float64 = 0, math.MaxFloat64
	var maxCount, minCount int
	var sumSquaredDev float64

	for _, count := range distribution {
		deviation := math.Abs(float64(count) - expected)
		deviations = append(deviations, deviation)
		sumSquaredDev += deviation * deviation

		if count > maxCount {
			maxCount = count
		}
		if count < minCount || minCount == 0 {
			minCount = count
		}

		if deviation > maxDev {
			maxDev = deviation
		}
		if deviation < minDev {
			minDev = deviation
		}
	}

	// Calcular desvio padrão e variância
	variance := sumSquaredDev / float64(numShards)
	stdDev := math.Sqrt(variance)

	// Coeficiente de Variação (CV)
	coefVariation := (stdDev / expected) * 100.0

	// Percentuais
	maxPercent := (float64(maxCount) / float64(totalKeys)) * 100.0
	minPercent := (float64(minCount) / float64(totalKeys)) * 100.0

	// Score de uniformidade (0-100, onde 100 é perfeito)
	uniformityScore := 100.0 - math.Min(coefVariation, 100.0)

	// Determinar qualidade baseada no CV
	var quality string
	if coefVariation <= 2.0 {
		quality = "EXCELENTE"
	} else if coefVariation <= 5.0 {
		quality = "MUITO BOA"
	} else if coefVariation <= 10.0 {
		quality = "BOA"
	} else if coefVariation <= 20.0 {
		quality = "REGULAR"
	} else {
		quality = "RUIM"
	}

	return DistributionResult{
		Algorithm:       hashFunc.Name,
		NumShards:       numShards,
		Distribution:    distribution,
		TotalKeys:       totalKeys,
		StdDev:          stdDev,
		MaxDev:          maxDev,
		MinDev:          minDev,
		Variance:        variance,
		CoefVariation:   coefVariation,
		MaxPercent:      maxPercent,
		MinPercent:      minPercent,
		Quality:         quality,
		Duration:        duration,
		UniformityScore: uniformityScore,
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

// printResults exibe os resultados em formato acadêmico detalhado
func printResults(results []DistributionResult, config BenchmarkConfig) {
	if len(results) == 0 {
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("BENCHMARK DE DISTRIBUIÇÃO DE HASH PARA ARQUITETURA CELULAR")
	fmt.Println("Análise de Algoritmos de Hashing vs. Número de Shards")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("\nTotal de UUID v4: %d\n", results[0].TotalKeys)
	fmt.Printf("Réplicas virtuais por shard: %d\n", config.NumReplicas)
	fmt.Println()

	// Agrupar resultados por número de shards
	resultsByShards := make(map[int][]DistributionResult)
	for _, result := range results {
		resultsByShards[result.NumShards] = append(resultsByShards[result.NumShards], result)
	}

	// Ordenar números de shards
	var shardCounts []int
	for count := range resultsByShards {
		shardCounts = append(shardCounts, count)
	}
	sort.Ints(shardCounts)

	// Imprimir resultados por número de shards
	for _, numShards := range shardCounts {
		results := resultsByShards[numShards]
		expected := float64(results[0].TotalKeys) / float64(numShards)

		fmt.Printf("\n%s\n", strings.Repeat("-", 100))
		fmt.Printf("CONFIGURAÇÃO: %d SHARDS (Esperado: %.0f chaves por shard)\n", numShards, expected)
		fmt.Printf("%s\n", strings.Repeat("-", 100))

		// Ordenar por score de uniformidade
		sort.Slice(results, func(i, j int) bool {
			return results[i].UniformityScore > results[j].UniformityScore
		})

		for rank, result := range results {
			// Calcular valores absolutos para melhor visualização
			var maxCount, minCount int
			for _, count := range result.Distribution {
				if count > maxCount {
					maxCount = count
				}
				if minCount == 0 || count < minCount {
					minCount = count
				}
			}
			deltaAbsoluto := maxCount - minCount
			deltaPercentual := result.MaxPercent - result.MinPercent

			fmt.Printf("\n[%d] %s\n", rank+1, result.Algorithm)
			fmt.Printf("    ├─ Qualidade: %s (Score: %.2f/100)\n", result.Quality, result.UniformityScore)
			fmt.Printf("    ├─ Coeficiente de Variação: %.2f%%\n", result.CoefVariation)
			fmt.Printf("    ├─ Desvio Padrão: %.2f chaves\n", result.StdDev)
			fmt.Printf("    ├─ Variância: %.2f\n", result.Variance)
			fmt.Printf("    ├─ Delta (Amplitude): %d chaves (%.2f%% diferença)\n", deltaAbsoluto, deltaPercentual)
			fmt.Printf("    ├─ Distribuição: %.2f%% - %.2f%% (Min: %d | Max: %d)\n",
				result.MinPercent, result.MaxPercent, minCount, maxCount)
			fmt.Printf("    ├─ Performance: %v\n", result.Duration)

			if config.Verbose {
				fmt.Printf("    └─ Detalhes por shard:\n")
				// Ordenar shards para exibição
				var shardNames []string
				for shard := range result.Distribution {
					shardNames = append(shardNames, shard)
				}
				sort.Strings(shardNames)

				for i, shard := range shardNames {
					count := result.Distribution[shard]
					percent := (float64(count) / float64(result.TotalKeys)) * 100.0
					deviation := math.Abs(float64(count) - expected)
					deviationPercent := (deviation / expected) * 100.0

					prefix := "       ├─"
					if i == len(shardNames)-1 {
						prefix = "       └─"
					}
					fmt.Printf("%s %s: %d (%.2f%%) [Desvio: %.0f / %.1f%%]\n",
						prefix, shard, count, percent, deviation, deviationPercent)
				}
			}
		}
	}

	// Resumo comparativo
	printComparativeSummary(resultsByShards, shardCounts)
}

// printComparativeSummary imprime resumo comparativo entre configurações
func printComparativeSummary(resultsByShards map[int][]DistributionResult, shardCounts []int) {
	fmt.Printf("\n\n%s\n", strings.Repeat("=", 100))
	fmt.Println("RESUMO COMPARATIVO: MELHOR ALGORITMO POR CONFIGURAÇÃO")
	fmt.Printf("%s\n", strings.Repeat("=", 100))

	fmt.Printf("\n%-12s | %-10s | %-12s | %-8s | %-15s | %-15s | %-18s | %-10s\n",
		"Num Shards", "Algoritmo", "Qualidade", "CV%", "Desvio Padrão", "Delta (Amp.)", "Score", "Blast Radius")
	fmt.Println(strings.Repeat("-", 130))

	for _, numShards := range shardCounts {
		results := resultsByShards[numShards]

		// Encontrar melhor algoritmo (menor CV)
		sort.Slice(results, func(i, j int) bool {
			return results[i].CoefVariation < results[j].CoefVariation
		})

		best := results[0]
		blastRadius := (1.0 / float64(numShards)) * 100.0
		
		// Calcular Delta (Amplitude)
		var maxCount, minCount int
		for _, count := range best.Distribution {
			if count > maxCount {
				maxCount = count
			}
			if minCount == 0 || count < minCount {
				minCount = count
			}
		}
		deltaAbsoluto := maxCount - minCount
		deltaPercentual := best.MaxPercent - best.MinPercent

		fmt.Printf("%-12d | %-10s | %-12s | %7.2f%% | %10.2f chaves | %6d (%5.2f%%) | %14.2f | %9.2f%%\n",
			numShards, best.Algorithm, best.Quality, best.CoefVariation,
			best.StdDev, deltaAbsoluto, deltaPercentual,
			best.UniformityScore, blastRadius)
	}

	fmt.Println()
}

// generateCSVReport gera relatório em CSV para análise
func generateCSVReport(results []DistributionResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Cabeçalho
	headers := []string{
		"Algoritmo", "NumShards", "TotalChaves", "DesvPadrao", "Variancia",
		"CoefVariacao", "MinPercent", "MaxPercent", "Diferenca", "Quality",
		"UniformityScore", "BlastRadius", "Duration",
	}
	writer.Write(headers)

	// Dados
	for _, r := range results {
		blastRadius := (1.0 / float64(r.NumShards)) * 100.0
		diferenca := r.MaxPercent - r.MinPercent

		row := []string{
			r.Algorithm,
			strconv.Itoa(r.NumShards),
			strconv.Itoa(r.TotalKeys),
			fmt.Sprintf("%.2f", r.StdDev),
			fmt.Sprintf("%.2f", r.Variance),
			fmt.Sprintf("%.2f", r.CoefVariation),
			fmt.Sprintf("%.2f", r.MinPercent),
			fmt.Sprintf("%.2f", r.MaxPercent),
			fmt.Sprintf("%.2f", diferenca),
			r.Quality,
			fmt.Sprintf("%.2f", r.UniformityScore),
			fmt.Sprintf("%.2f", blastRadius),
			r.Duration.String(),
		}
		writer.Write(row)
	}

	fmt.Printf("\n✓ Relatório CSV gerado: %s\n", filename)
	return nil
}

// generateMarkdownTable gera tabela markdown para artigo
func generateMarkdownTable(results []DistributionResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Agrupar por número de shards
	resultsByShards := make(map[int][]DistributionResult)
	for _, result := range results {
		resultsByShards[result.NumShards] = append(resultsByShards[result.NumShards], result)
	}

	var shardCounts []int
	for count := range resultsByShards {
		shardCounts = append(shardCounts, count)
	}
	sort.Ints(shardCounts)

	fmt.Fprintf(file, "# Benchmark de Distribuição de Hash - Arquitetura Celular\n\n")
	fmt.Fprintf(file, "## Metodologia\n\n")
	fmt.Fprintf(file, "- **Dataset**: 1.000.000 UUID v4\n")
	fmt.Fprintf(file, "- **Algoritmos testados**: SHA-512, SHA-256, SHA-1, MD5, MURMUR3\n")
	fmt.Fprintf(file, "- **Configurações de shards**: %v\n", shardCounts)
	fmt.Fprintf(file, "- **Réplicas virtuais**: %d por shard\n\n", results[0].TotalKeys)

	for _, numShards := range shardCounts {
		results := resultsByShards[numShards]
		sort.Slice(results, func(i, j int) bool {
			return results[i].CoefVariation < results[j].CoefVariation
		})

		blastRadius := (1.0 / float64(numShards)) * 100.0

		fmt.Fprintf(file, "## Configuração: %d Shards (Blast Radius: %.2f%%)\n\n", numShards, blastRadius)
		fmt.Fprintf(file, "| Rank | Algoritmo | Coef. Variação | Desvio Padrão | Min%% | Max%% | Δ%% | Qualidade | Score |\n")
		fmt.Fprintf(file, "|------|-----------|----------------|---------------|------|------|-----|-----------|-------|\n")

		for i, r := range results {
			fmt.Fprintf(file, "| %d | %s | %.2f%% | %.2f | %.2f%% | %.2f%% | %.2f%% | %s | %.2f |\n",
				i+1, r.Algorithm, r.CoefVariation, r.StdDev,
				r.MinPercent, r.MaxPercent, r.MaxPercent-r.MinPercent,
				r.Quality, r.UniformityScore)
		}
		fmt.Fprintf(file, "\n")
	}

	// Tabela resumo
	fmt.Fprintf(file, "## Resumo: Melhor Algoritmo por Configuração\n\n")
	fmt.Fprintf(file, "| Num Shards | Algoritmo Ideal | CV%% | Score | Blast Radius | Disponibilidade |\n")
	fmt.Fprintf(file, "|------------|----------------|------|-------|--------------|------------------|\n")

	for _, numShards := range shardCounts {
		results := resultsByShards[numShards]
		sort.Slice(results, func(i, j int) bool {
			return results[i].CoefVariation < results[j].CoefVariation
		})

		best := results[0]
		blastRadius := (1.0 / float64(numShards)) * 100.0
		availability := ((float64(numShards) - 1.0) / float64(numShards)) * 100.0

		fmt.Fprintf(file, "| %d | %s | %.2f%% | %.2f | %.2f%% | %.2f%% |\n",
			numShards, best.Algorithm, best.CoefVariation, best.UniformityScore,
			blastRadius, availability)
	}

	fmt.Fprintf(file, "\n## Conclusões\n\n")
	fmt.Fprintf(file, "1. **Uniformidade de Distribuição**: Análise do coeficiente de variação\n")
	fmt.Fprintf(file, "2. **Trade-off Blast Radius**: Relação entre número de shards e disponibilidade\n")
	fmt.Fprintf(file, "3. **Performance de Algoritmos**: Comparação entre algoritmos criptográficos e não-criptográficos\n")

	fmt.Printf("✓ Tabela Markdown gerada: %s\n", filename)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Uso: %s <arquivo-uuids> [flags]\n\n", os.Args[0])
		fmt.Println("Flags opcionais:")
		fmt.Println("  -shards=3,5,10,25,50,100  Números de shards para testar (padrão: 3,5,10)")
		fmt.Println("  -replicas=150             Número de réplicas virtuais (padrão: 150)")
		fmt.Println("  -csv=results.csv          Gerar relatório CSV")
		fmt.Println("  -md=results.md            Gerar tabela Markdown")
		fmt.Println("  -verbose                  Exibir detalhes completos")
		fmt.Println("\nExemplo:")
		fmt.Println("  go run main.go uuids.txt -shards=3,5,10,25,50,100 -csv=benchmark.csv -md=benchmark.md -verbose")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Configuração padrão
	config := BenchmarkConfig{
		ShardCounts:   []int{3, 5, 10},
		NumReplicas:   150,
		GenerateCSV:   false,
		GenerateTable: false,
		Verbose:       false,
	}

	var csvFile, mdFile string

	// Parse de argumentos
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]

		if strings.HasPrefix(arg, "-shards=") {
			shardsStr := strings.TrimPrefix(arg, "-shards=")
			config.ShardCounts = []int{}
			for _, s := range strings.Split(shardsStr, ",") {
				num, err := strconv.Atoi(strings.TrimSpace(s))
				if err == nil && num > 0 {
					config.ShardCounts = append(config.ShardCounts, num)
				}
			}
		} else if strings.HasPrefix(arg, "-replicas=") {
			replicasStr := strings.TrimPrefix(arg, "-replicas=")
			if num, err := strconv.Atoi(replicasStr); err == nil && num > 0 {
				config.NumReplicas = num
			}
		} else if strings.HasPrefix(arg, "-csv=") {
			csvFile = strings.TrimPrefix(arg, "-csv=")
			config.GenerateCSV = true
		} else if strings.HasPrefix(arg, "-md=") {
			mdFile = strings.TrimPrefix(arg, "-md=")
			config.GenerateTable = true
		} else if arg == "-verbose" {
			config.Verbose = true
		}
	}

	// Verificar se arquivo existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatalf("❌ Arquivo não encontrado: %s", filename)
	}

	fmt.Println("🔬 Iniciando Benchmark de Distribuição de Hash...")
	fmt.Printf("📁 Arquivo: %s\n", filename)

	// Ler chaves do arquivo
	fmt.Print("📖 Lendo UUIDs... ")
	keys, err := readKeysFromFile(filename)
	if err != nil {
		log.Fatalf("\n❌ Erro ao ler arquivo: %v", err)
	}

	if len(keys) == 0 {
		log.Fatalf("\n❌ Nenhuma chave encontrada no arquivo")
	}
	fmt.Printf("✓ %d UUIDs carregados\n", len(keys))

	// Definir funções de hash disponíveis (MD5 removido por questões de segurança)
	hashFunctions := []HashFunction{
		{"SHA-512", hashKeySHA512},
		{"SHA-256", hashKeySHA256},
		{"SHA-1", hashKeySHA1},
		{"MURMUR3", hashKeyMurmur},
	}

	// Executar benchmark
	fmt.Printf("\n🚀 Executando benchmark com %d algoritmos e %d configurações...\n",
		len(hashFunctions), len(config.ShardCounts))

	var allResults []DistributionResult
	totalTests := len(hashFunctions) * len(config.ShardCounts)
	currentTest := 0

	for _, numShards := range config.ShardCounts {
		for _, hashFunc := range hashFunctions {
			currentTest++
			fmt.Printf("\r⏳ Progresso: %d/%d testes [%d shards - %s]          ",
				currentTest, totalTests, numShards, hashFunc.Name)

			result := analyzeDistribution(keys, hashFunc, numShards, config.NumReplicas)
			allResults = append(allResults, result)
		}
	}
	fmt.Println("\n✅ Benchmark concluído!")

	// Exibir resultados
	printResults(allResults, config)

	// Gerar relatórios
	if config.GenerateCSV {
		if csvFile == "" {
			csvFile = "benchmark_results.csv"
		}
		if err := generateCSVReport(allResults, csvFile); err != nil {
			log.Printf("⚠️  Erro ao gerar CSV: %v", err)
		}
	}

	if config.GenerateTable {
		if mdFile == "" {
			mdFile = "benchmark_results.md"
		}
		if err := generateMarkdownTable(allResults, mdFile); err != nil {
			log.Printf("⚠️  Erro ao gerar Markdown: %v", err)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("✨ Benchmark finalizado com sucesso!")
	fmt.Println(strings.Repeat("=", 100))
}
