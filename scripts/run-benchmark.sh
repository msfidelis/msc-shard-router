#!/bin/bash

# Script para executar benchmark completo de distribuição de hash
# Autor: Mestrado em Arquitetura Celular
# Data: Dezembro 2025

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
UUID_FILE="$PROJECT_ROOT/examples/hashing-test/1kk_uuids.txt"
RESULTS_DIR="$PROJECT_ROOT/benchmark-results"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  BENCHMARK DE DISTRIBUIÇÃO - ARQUITETURA CELULAR${NC}"
echo -e "${BLUE}  Análise de Algoritmos de Hash vs. Número de Shards${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
echo ""

# Verificar se arquivo de UUIDs existe
if [ ! -f "$UUID_FILE" ]; then
    echo -e "${RED}❌ Arquivo de UUIDs não encontrado: $UUID_FILE${NC}"
    echo -e "${YELLOW}💡 Gerando 1 milhão de UUIDs...${NC}"
    
    cd "$PROJECT_ROOT/examples/hashing-test"
    go run generate_uuids.go
    
    if [ ! -f "$UUID_FILE" ]; then
        echo -e "${RED}❌ Falha ao gerar UUIDs${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ UUIDs gerados com sucesso${NC}"
fi

# Criar diretório de resultados
mkdir -p "$RESULTS_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo -e "${YELLOW}📊 Configurações do Benchmark:${NC}"
echo "   • Dataset: 1.000.000 UUID v4"
echo "   • Algoritmos: SHA-512, SHA-256, SHA-1, MD5, MURMUR3"
echo "   • Réplicas virtuais: 150"
echo "   • Diretório de saída: $RESULTS_DIR"
echo ""

# Executar benchmarks com diferentes configurações

echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}[1/3] Benchmark Básico: 3, 5, 10 shards${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

cd "$PROJECT_ROOT"
go run cmd/hashing-distribution/main.go "$UUID_FILE" \
    -shards=3,5,10 \
    -replicas=150 \
    -csv="$RESULTS_DIR/benchmark_basic_${TIMESTAMP}.csv" \
    -md="$RESULTS_DIR/benchmark_basic_${TIMESTAMP}.md" \
    -verbose

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}[2/3] Benchmark Médio: 10, 25, 50 shards${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

go run cmd/hashing-distribution/main.go "$UUID_FILE" \
    -shards=10,25,50 \
    -replicas=150 \
    -csv="$RESULTS_DIR/benchmark_medium_${TIMESTAMP}.csv" \
    -md="$RESULTS_DIR/benchmark_medium_${TIMESTAMP}.md" \
    -verbose

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}[3/3] Benchmark Avançado: 50, 100, 200 shards${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

go run cmd/hashing-distribution/main.go "$UUID_FILE" \
    -shards=50,100,200 \
    -replicas=150 \
    -csv="$RESULTS_DIR/benchmark_advanced_${TIMESTAMP}.csv" \
    -md="$RESULTS_DIR/benchmark_advanced_${TIMESTAMP}.md"

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}[EXTRA] Benchmark Completo: Todas as configurações${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

go run cmd/hashing-distribution/main.go "$UUID_FILE" \
    -shards=3,5,10,25,50,100,200 \
    -replicas=150 \
    -csv="$RESULTS_DIR/benchmark_complete_${TIMESTAMP}.csv" \
    -md="$RESULTS_DIR/benchmark_complete_${TIMESTAMP}.md"

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✨ Benchmark Concluído com Sucesso!${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}📁 Resultados salvos em:${NC}"
echo "   $RESULTS_DIR"
echo ""
echo -e "${YELLOW}📊 Arquivos gerados:${NC}"
ls -lh "$RESULTS_DIR"/*${TIMESTAMP}* | awk '{print "   " $9 " (" $5 ")"}'
echo ""
echo -e "${YELLOW}💡 Próximos passos:${NC}"
echo "   1. Analise os arquivos .md para inclusão no artigo"
echo "   2. Importe os arquivos .csv para análise estatística"
echo "   3. Compare os resultados entre diferentes configurações"
echo ""
