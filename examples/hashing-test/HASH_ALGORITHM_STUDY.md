# Estudo Comparativo de Algoritmos de Hashing para Shard Router

## Introdução

Este documento apresenta um estudo comparativo de algoritmos de hashing para distribuição de chaves em um sistema de sharding consistente. O objetivo é determinar qual algoritmo oferece a melhor combinação de **distribuição uniforme** e **performance** para o MSC Shard Router.

## Metodologia

### Configuração do Teste
- **Chaves de teste**: 1000 UUIDs v4 aleatórios
- **Número de shards**: 3 (shard01, shard02, shard03)
- **Réplicas virtuais**: 10 por shard
- **Distribuição ideal**: ~333.33 UUIDs por shard (33.33% cada)
- **Plataforma**: Apple M4, macOS, Go 1.23

### Critérios de Avaliação
- **EXCELENTE**: Desvio médio ≤ 10%
- **BOA**: Desvio médio ≤ 15% 
- **REGULAR**: Desvio médio ≤ 25%
- **RUIM**: Desvio médio > 25%

### Algoritmos Testados
1. **SHA-512** (algoritmo atual do sistema)
2. **SHA-256** (padrão da indústria)
3. **MD5** (algoritmo legado)
4. **FNV-1a** (algoritmo não-criptográfico)

## Resultados dos Testes

### Distribuição de UUIDs

| Algoritmo | Shard01 | Shard02 | Shard03 | Desvio Médio | Qualidade |
|-----------|---------|---------|---------|--------------|-----------|
| **SHA-512** | 359 (35.9%) | 265 (26.5%) | 376 (37.6%) | 45.6 UUIDs (13.7%) | ✅ **BOA** |
| **SHA-256** | 286 (28.6%) | 445 (44.5%) | 269 (26.9%) | 74.4 UUIDs (22.3%) | ⚠️ **REGULAR** |
| **MD5** | 305 (30.5%) | 429 (42.9%) | 266 (26.6%) | 63.8 UUIDs (19.1%) | ⚠️ **REGULAR** |
| **FNV-1a** | 910 (91.0%) | 32 (3.2%) | 58 (5.8%) | 384.4 UUIDs (115.3%) | ❌ **RUIM** |

### Performance Benchmark

| Algoritmo | ns/operação | Alocações/op | Bytes/op | Velocidade Relativa |
|-----------|-------------|--------------|----------|-------------------|
| **FNV-1a** | 16.03 | 0 | 0 | 🚀 **6.7x mais rápido** |
| **SHA-256** | 46.67 | 1 | 32 | 🏃 **2.3x mais rápido** |
| **MD5** | 90.19 | 1 | 16 | ⚡ **1.2x mais rápido** |
| **SHA-512** | 107.3 | 1 | 64 | 📊 **Baseline** |

## Análise Detalhada

### 🏆 SHA-512 (Atual)
**Distribuição**: ✅ BOA (13.7% desvio)
**Performance**: 107.3 ns/op

**Pontos Positivos:**
- Melhor distribuição entre todos os algoritmos testados
- Máxima segurança criptográfica
- Resistente a ataques de colisão
- Consistência comprovada em produção

**Pontos de Atenção:**
- Maior custo computacional
- Maior uso de memória (64 bytes/op)

### 🥈 SHA-256
**Distribuição**: ⚠️ REGULAR (22.3% desvio)
**Performance**: 46.67 ns/op (2.3x mais rápido)

**Pontos Positivos:**
- Boa performance (2.3x mais rápido que SHA-512)
- Segurança criptográfica adequada
- Padrão da indústria
- Menor uso de memória que SHA-512

**Pontos de Atenção:**
- Distribuição menos uniforme que SHA-512
- Concentração de chaves no shard02 (44.5%)

### 🥉 MD5
**Distribuição**: ⚠️ REGULAR (19.1% desvio)
**Performance**: 90.19 ns/op (1.2x mais rápido)

**Pontos Positivos:**
- Performance razoável
- Menor uso de memória (16 bytes/op)
- Distribuição melhor que SHA-256

**Pontos de Atenção:**
- Vulnerabilidades de segurança conhecidas
- Não recomendado para novos sistemas
- Concentração de chaves no shard02 (42.9%)

### ❌ FNV-1a
**Distribuição**: ❌ RUIM (115.3% desvio)
**Performance**: 16.03 ns/op (6.7x mais rápido)

**Pontos Positivos:**
- Máxima performance (6.7x mais rápido)
- Zero alocações de memória
- Ideal para cases não-criptográficos

**Pontos Críticos:**
- Distribuição completamente desigual
- 91% das chaves concentradas em um único shard
- Inviável para uso em produção com consistent hashing

## Conclusões e Recomendações

### Ranking Final

1. **🏆 SHA-512** - Melhor balanço distribuição/segurança
2. **🥈 SHA-256** - Boa opção para performance/segurança
3. **🥉 MD5** - Apenas se segurança não for crítica
4. **❌ FNV-1a** - Inadequado para consistent hashing

### Recomendações por Cenário

#### 🎯 **Produção (Recomendado)**
**Manter SHA-512**
- Melhor distribuição uniforme (13.7% desvio)
- Máxima segurança para dados sensíveis
- Performance aceitável para a maioria dos casos

#### ⚡ **Performance Crítica**
**Migrar para SHA-256**
- 2.3x mais rápido que SHA-512
- Distribuição aceitável (22.3% desvio)
- Segurança adequada para a maioria dos sistemas

#### 🧪 **Desenvolvimento/Teste**
**SHA-256 ou MD5**
- Maior velocidade para ciclos de desenvolvimento
- MD5 apenas se segurança não for requisito

#### ❌ **Não Recomendado**
**FNV-1a para Consistent Hashing**
- Distribuição completamente desigual
- Pode funcionar para outros tipos de hash tables
- Inadequado para sharding distribuído

### Implementação Sugerida

```go
// Configuração flexível de algoritmo
type HashAlgorithm int

const (
    SHA512 HashAlgorithm = iota  // Produção
    SHA256                       // Performance/Segurança
    MD5                          // Desenvolvimento
)

func (ring *ConsistentHashRing) SetHashAlgorithm(algo HashAlgorithm) {
    switch algo {
    case SHA512:
        ring.hashFunc = hashKeySHA512
    case SHA256:
        ring.hashFunc = hashKeySHA256
    case MD5:
        ring.hashFunc = hashKeyMD5
    }
}
```

## Considerações Acadêmicas

### Consistent Hashing vs Performance
Este estudo demonstra que **nem sempre o algoritmo mais rápido é o melhor** para consistent hashing. O FNV-1a, apesar de ser 6.7x mais rápido, produz uma distribuição completamente inadequada.

### Trade-offs Observados
- **Segurança vs Performance**: SHA-256 oferece 2.3x mais performance com perda aceitável de segurança
- **Distribuição vs Velocidade**: SHA-512 mantém a melhor distribuição mesmo sendo mais lento
- **Memória vs Performance**: FNV-1a usa zero alocações mas falha na distribuição

### Implicações para Arquitetura Celular
- **Isolamento**: Distribuição desigual pode quebrar o isolamento entre células
- **Scalabilidade**: Algoritmos mal distribuídos criam hotspots
- **Resiliência**: Concentração de carga prejudica a tolerância a falhas

## Próximos Passos

1. **Configurabilidade**: Implementar seleção dinâmica de algoritmo
2. **Monitoramento**: Adicionar métricas de distribuição em produção
3. **Teste de Carga**: Validar resultados com cargas reais
4. **Número de Réplicas**: Estudar impacto de diferentes números de réplicas virtuais

---

*Estudo realizado como parte da pesquisa de Mestrado em Arquitetura Celular*  
*MSC Shard Router - Setembro 2024*

## Anexo: Dados Brutos

### UUIDs de Teste (Amostra)
```
Primeiros 10 UUIDs utilizados nos testes:
[Lista seria gerada dinamicamente durante o teste]
```

### Comandos para Reprodução
```bash
# Executar teste de distribuição
go test ./pkg/hashring -v -run TestCompareHashingAlgorithms

# Executar benchmark de performance  
go test ./pkg/hashring -bench=BenchmarkHashAlgorithms -run=^$ -benchmem
```