# Analisador de Distribuição de Hashing

## Descrição

Este utilitário analisa a distribuição de chaves entre shards usando todos os algoritmos de hash disponíveis no MSC Shard Router.

## Uso

### Compilação
```bash
make build-distribution
```

### Execução Direta
```bash
go run cmd/hashing-distribution/main.go <arquivo-de-chaves>
```

### Execução com Binário Compilado
```bash
./hashing-distribution <arquivo-de-chaves>
```

### Com Makefile
```bash
# Analisar arquivo específico
make analyze-distribution FILE=1kk_uuids.txt

# Teste com 100 chaves aleatórias
make test-distribution
```

## Formato do Arquivo de Entrada

O arquivo deve conter uma chave por linha:
```
uuid-1
uuid-2
uuid-3
...
```

## Formato de Saída

Para cada algoritmo de hash, mostra:

```
SHA-1
  shard01 : 311486 chaves ( 31.1%) - desvio: 21847.3
  shard02 : 471716 chaves ( 47.2%) - desvio: 138382.7
  shard03 : 216798 chaves ( 21.7%) - desvio: 116535.3
  Estatísticas:
    Desvio padrão: 105210.2
    Variância: 11069184107.6
    Performance: 155.701833ms
    Qualidade: RUIM
```

## Algoritmos Analisados

- **SHA-512**: Algoritmo padrão do sistema
- **SHA-256**: Alternativa com boa performance
- **SHA-1**: Algoritmo legado
- **MD5**: Algoritmo rápido mas inseguro
- **MURMUR**: Algoritmo não-criptográfico de alta performance

## Métricas

- **Desvio**: Distância da distribuição ideal (33.33% por shard)
- **Desvio Padrão**: Medida da variabilidade da distribuição
- **Variância**: Quadrado do desvio padrão
- **Performance**: Tempo total para processar todas as chaves
- **Qualidade**: Classificação baseada no desvio médio
  - EXCELENTE: ≤ 5%
  - MUITO BOA: ≤ 10%
  - BOA: ≤ 15%
  - REGULAR: ≤ 25%
  - RUIM: > 25%

## Exemplos

### Arquivo pequeno (5 chaves)
```bash
echo -e "key1\nkey2\nkey3\nkey4\nkey5" > test.txt
go run cmd/hashing-distribution/main.go test.txt
```

### Arquivo grande (1M de UUIDs)
```bash
go run cmd/hashing-distribution/main.go 1kk_uuids.txt
```

## Uso em Pesquisa

Este utilitário é ideal para:
- Validar algoritmos de hash para consistent hashing
- Comparar distribuições entre diferentes algoritmos
- Análise de performance de algoritmos de hash
- Geração de dados para papers acadêmicos
- Validação empírica de escolhas arquiteturais