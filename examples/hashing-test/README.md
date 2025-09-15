# Teste de Distribuição de Hash - 1 Milhão de UUIDs

Este diretório contém os testes de distribuição de hash com 1 milhão de UUIDs aleatórios para avaliar a performance e qualidade de diferentes algoritmos de hash em um sistema de sharding consistente.

## Arquivos

- `1kk_uuids.txt` - Arquivo com 1.000.000 UUIDs aleatórios para teste
- `results.md` - Resultados detalhados da análise
- `generate_uuids.go` - Script para gerar novos UUIDs (se necessário)

## Como Executar

```bash
# A partir da raiz do projeto
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt
```

## Configuração do Teste

- **Número de Shards**: 3 (shard01, shard02, shard03)
- **Réplicas Virtuais**: 10 por shard
- **Total de Chaves**: 1.000.000 UUIDs
- **Distribuição Ideal**: ~333.333 chaves por shard (33.3% cada)

## Métricas Avaliadas

1. **Distribuição por Shard**: Número e percentual de chaves em cada shard
2. **Desvio**: Diferença absoluta da distribuição ideal
3. **Desvio Padrão**: Medida da variabilidade da distribuição
4. **Variância**: Quadrado do desvio padrão
5. **Performance**: Tempo total de processamento
6. **Qualidade**: Classificação baseada no desvio padrão
   - EXCELENTE: < 1000
   - BOA: 1000-10000
   - REGULAR: 10000-100000
   - RUIM: > 100000

## Algoritmos Testados

- **SHA-512**: Hash criptográfico de 512 bits
- **SHA-256**: Hash criptográfico de 256 bits  
- **SHA-1**: Hash criptográfico de 160 bits
- **MD5**: Hash de 128 bits (deprecated para segurança)
- **MURMUR**: Hash não-criptográfico otimizado para performance