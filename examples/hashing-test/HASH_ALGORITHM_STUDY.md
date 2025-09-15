# Estudo Comparativo de Algoritmos de Hashing

## Resumo Executivo

Este estudo compara diferentes algoritmos de hashing para distribuição de chaves entre shards,
avaliando a uniformidade da distribuição e performance dos algoritmos.

## Resultados Comparativos

| Algoritmo | Desvio Padrão | Variância | Melhor Shard (%) | Pior Shard (%) | Diferença |
|-----------|---------------|-----------|------------------|----------------|----------|
| SHA512 | 28.31 | 801.67 | 30.5% | 37.2% | 6.7% |
| SHA256 | 64.60 | 4173.67 | 26.3% | 41.9% | 15.6% |
| SHA1 | 11.03 | 121.67 | 32.0% | 34.7% | 2.7% |
| MD5 | 42.05 | 1768.33 | 28.2% | 38.5% | 10.3% |
| FNV64 | 395.12 | 156116.33 | 2.4% | 89.1% | 86.7% |
| SimpleHash | 471.40 | 222222.33 | 0.0% | 100.0% | 100.0% |

### SHA512

**Distribuição por Shard:**

| Shard | Quantidade | Percentual |
|-------|------------|------------|
| shard-01 | 323 | 32.3% |
| shard-02 | 305 | 30.5% |
| shard-03 | 372 | 37.2% |

**Estatísticas:**
- Total de chaves: 1000
- Esperado por shard: 333
- Desvio padrão: 28.31
- Variância: 801.67

### SHA256

**Distribuição por Shard:**

| Shard | Quantidade | Percentual |
|-------|------------|------------|
| shard-01 | 419 | 41.9% |
| shard-02 | 263 | 26.3% |
| shard-03 | 318 | 31.8% |

**Estatísticas:**
- Total de chaves: 1000
- Esperado por shard: 333
- Desvio padrão: 64.60
- Variância: 4173.67

### SHA1

**Distribuição por Shard:**

| Shard | Quantidade | Percentual |
|-------|------------|------------|
| shard-01 | 347 | 34.7% |
| shard-02 | 320 | 32.0% |
| shard-03 | 333 | 33.3% |

**Estatísticas:**
- Total de chaves: 1000
- Esperado por shard: 333
- Desvio padrão: 11.03
- Variância: 121.67

### MD5

**Distribuição por Shard:**

| Shard | Quantidade | Percentual |
|-------|------------|------------|
| shard-03 | 385 | 38.5% |
| shard-01 | 333 | 33.3% |
| shard-02 | 282 | 28.2% |

**Estatísticas:**
- Total de chaves: 1000
- Esperado por shard: 333
- Desvio padrão: 42.05
- Variância: 1768.33

### FNV64

**Distribuição por Shard:**

| Shard | Quantidade | Percentual |
|-------|------------|------------|
| shard-01 | 24 | 2.4% |
| shard-02 | 891 | 89.1% |
| shard-03 | 85 | 8.5% |

**Estatísticas:**
- Total de chaves: 1000
- Esperado por shard: 333
- Desvio padrão: 395.12
- Variância: 156116.33

### SimpleHash

**Distribuição por Shard:**

| Shard | Quantidade | Percentual |
|-------|------------|------------|
| shard-01 | 1000 | 100.0% |
| shard-02 | 0 | 0.0% |
| shard-03 | 0 | 0.0% |

**Estatísticas:**
- Total de chaves: 1000
- Esperado por shard: 333
- Desvio padrão: 471.40
- Variância: 222222.33

## Análise e Recomendações

**Melhor algoritmo:** SHA1

**Critérios de avaliação:**
1. **Menor desvio padrão** - indica distribuição mais uniforme
2. **Menor variância** - confirma consistência da distribuição
3. **Menor diferença** entre melhor e pior shard

**Considerações de segurança:**
- SHA-512 e SHA-256 são criptograficamente seguros
- MD5 e SHA-1 são considerados deprecados para uso criptográfico
- FNV64 e SimpleHash são rápidos mas não criptograficamente seguros

**Recomendação final:**
Para aplicações que requerem segurança criptográfica, use **SHA-256** ou **SHA-512**.
Para aplicações onde a performance é crítica e segurança criptográfica não é necessária, considere **FNV64**.
