# Resultados da Análise de Distribuição de Hash

**Data do Teste**: 15 de Setembro de 2025  
**Total de Chaves**: 1.000.000 UUIDs  
**Configuração**: 3 shards, 10 réplicas virtuais cada  
**Distribuição Ideal**: 333.333 chaves por shard (33.3%)

## Resumo dos Resultados

### 🏆 Ranking por Qualidade de Distribuição (Desvio Padrão)

1. **SHA-256** - 70.786,6 (REGULAR)
2. **MD5** - 74.979,5 (REGULAR) 
3. **MURMUR** - 83.999,9 (REGULAR)
4. **SHA-512** - 84.770,2 (REGULAR)
5. **SHA-1** - 105.210,2 (REGULAR)

### ⚡ Ranking por Performance

1. **MURMUR** - 118,67ms
2. **SHA-256** - 152,54ms  
3. **SHA-1** - 155,27ms
4. **MD5** - 202,10ms
5. **SHA-512** - 213,82ms

## Resultados Detalhados

### SHA-512
```
Distribuição:
  shard01: 362.644 chaves (36.3%) - desvio: 29.310,7
  shard02: 419.349 chaves (41.9%) - desvio: 86.015,7  
  shard03: 218.007 chaves (21.8%) - desvio: 115.326,3

Estatísticas:
  Desvio padrão: 84.770,2
  Variância: 7.185.991.084,2
  Performance: 213,82ms
  Qualidade: REGULAR
```

### SHA-256
```
Distribuição:
  shard01: 309.824 chaves (31.0%) - desvio: 23.509,3
  shard02: 429.359 chaves (42.9%) - desvio: 96.025,7
  shard03: 260.817 chaves (26.1%) - desvio: 72.516,3

Estatísticas:
  Desvio padrão: 70.786,6
  Variância: 5.010.745.337,6
  Performance: 152,54ms
  Qualidade: REGULAR
```

### SHA-1
```
Distribuição:
  shard01: 311.486 chaves (31.1%) - desvio: 21.847,3
  shard02: 471.716 chaves (47.2%) - desvio: 138.382,7
  shard03: 216.798 chaves (21.7%) - desvio: 116.535,3

Estatísticas:
  Desvio padrão: 105.210,2
  Variância: 11.069.184.107,6
  Performance: 155,27ms
  Qualidade: REGULAR
```

### MD5
```
Distribuição:
  shard01: 294.112 chaves (29.4%) - desvio: 39.221,3
  shard02: 438.262 chaves (43.8%) - desvio: 104.928,7
  shard03: 267.626 chaves (26.8%) - desvio: 65.707,3

Estatísticas:
  Desvio padrão: 74.979,5
  Variância: 5.621.930.576,9
  Performance: 202,10ms
  Qualidade: REGULAR
```

### MURMUR
```
Distribuição:
  shard01: 233.036 chaves (23.3%) - desvio: 100.297,3
  shard02: 438.612 chaves (43.9%) - desvio: 105.278,7
  shard03: 328.352 chaves (32.8%) - desvio: 4.981,3

Estatísticas:
  Desvio padrão: 83.999,9
  Variância: 7.055.988.803,6
  Performance: 118,67ms
  Qualidade: REGULAR
```

## Análise e Conclusões

### Distribuição de Chaves

Todos os algoritmos apresentaram qualidade **REGULAR** de distribuição, com desvios padrão na faixa de 70k-105k chaves. Nenhum algoritmo conseguiu uma distribuição próxima da ideal (33.3% por shard).

### Observações Importantes

1. **SHA-256** teve a melhor distribuição (menor desvio padrão)
2. **MURMUR** foi o mais rápido, mas teve distribuição irregular  
3. **SHA-1** teve a pior distribuição (maior desvio padrão)
4. **SHA-512** foi o mais lento, com distribuição mediana
5. **MD5** apresentou equilíbrio entre performance e distribuição

### Recomendações

Para um sistema de sharding com estas características:

- **Para melhor distribuição**: SHA-256
- **Para melhor performance**: MURMUR  
- **Para equilíbrio**: MD5 (considerando apenas aspectos técnicos)

**Nota**: MD5 não é recomendado para uso em produção devido a vulnerabilidades de segurança conhecidas.

### Próximos Passos

1. Testar com diferentes números de réplicas virtuais
2. Avaliar com datasets de diferentes tamanhos
3. Considerar algoritmos adicionais (xxHash, CityHash)
4. Implementar testes de redistribuição (add/remove shards)