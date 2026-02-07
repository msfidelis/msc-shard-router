# Benchmark de Distribuição de Hash - Arquitetura Celular

## Visão Geral

Este benchmark foi desenvolvido para análise acadêmica do comportamento de diferentes algoritmos de hash na distribuição de requisições em arquiteturas celulares. O objetivo é identificar o equilíbrio ideal entre algoritmo de hashing e número de shards.

## Metodologia

### Dataset
- **1.000.000 UUID v4**: Identificadores únicos gerados aleatoriamente
- **Réplicas virtuais**: 150 por shard (configurável)
- **Algoritmos testados**:
  - SHA-512 (criptográfico)
  - SHA-256 (criptográfico)
  - SHA-1 (criptográfico)
  - MD5 (criptográfico, deprecated)
  - MURMUR3 (não-criptográfico, alta performance)

### Métricas Avaliadas

1. **Coeficiente de Variação (CV)**: Métrica principal de uniformidade
   - CV < 2%: EXCELENTE
   - CV 2-5%: MUITO BOA
   - CV 5-10%: BOA
   - CV 10-20%: REGULAR
   - CV > 20%: RUIM

2. **Desvio Padrão**: Dispersão absoluta da distribuição
3. **Variância**: Quadrado do desvio padrão
4. **Score de Uniformidade**: 0-100 (100 = distribuição perfeita)
5. **Blast Radius**: Impacto de falha (1/N × 100%)
6. **Performance**: Tempo de processamento

## Uso Rápido

### Executar Benchmark Completo

```bash
# Dar permissão de execução ao script
chmod +x scripts/run-benchmark.sh

# Executar benchmark completo
./scripts/run-benchmark.sh
```

Este script irá:
1. Verificar/gerar os 1 milhão de UUIDs
2. Executar 3 configurações de benchmark
3. Gerar relatórios CSV e Markdown
4. Salvar resultados em `benchmark-results/`

### Executar Benchmark Customizado

```bash
# Sintaxe
go run cmd/hashing-distribution/main.go <arquivo-uuids> [flags]

# Exemplo: Testar 3, 5 e 10 shards
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=3,5,10 \
    -replicas=150 \
    -csv=results.csv \
    -md=results.md \
    -verbose

# Exemplo: Testar configurações de produção
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=50,100,200,500 \
    -replicas=200 \
    -csv=production_benchmark.csv \
    -md=production_benchmark.md
```

### Flags Disponíveis

| Flag | Descrição | Padrão | Exemplo |
|------|-----------|--------|---------|
| `-shards` | Números de shards para testar | `3,5,10` | `-shards=3,5,10,25,50,100` |
| `-replicas` | Réplicas virtuais por shard | `150` | `-replicas=200` |
| `-csv` | Gerar relatório CSV | - | `-csv=results.csv` |
| `-md` | Gerar tabela Markdown | - | `-md=results.md` |
| `-verbose` | Exibir detalhes completos | `false` | `-verbose` |

## Interpretação dos Resultados

### Output do Terminal

O benchmark exibe resultados organizados por número de shards, ranqueados por **Score de Uniformidade**:

```
═════════════════════════════════════════════════════════════
BENCHMARK DE DISTRIBUIÇÃO DE HASH PARA ARQUITETURA CELULAR
═════════════════════════════════════════════════════════════

Total de UUID v4: 1000000
Réplicas virtuais por shard: 150

──────────────────────────────────────────────────────────────
CONFIGURAÇÃO: 10 SHARDS (Esperado: 100000 chaves por shard)
──────────────────────────────────────────────────────────────

[1] SHA-1
    ├─ Qualidade: EXCELENTE (Score: 98.45/100)
    ├─ Coeficiente de Variação: 1.55%
    ├─ Desvio Padrão: 1547.23
    ├─ Variância: 2393920.45
    ├─ Distribuição: 9.85% - 10.12% (Δ: 0.27%)
    └─ Performance: 245ms
```

### Relatório CSV

Arquivo estruturado para análise estatística:

```csv
Algoritmo,NumShards,TotalChaves,DesvPadrao,Variancia,CoefVariacao,MinPercent,MaxPercent,Diferenca,Quality,UniformityScore,BlastRadius,Duration
SHA-1,10,1000000,1547.23,2393920.45,1.55,9.85,10.12,0.27,EXCELENTE,98.45,10.00,245ms
```

### Relatório Markdown

Tabelas formatadas para inclusão direta no artigo acadêmico:

```markdown
## Configuração: 10 Shards (Blast Radius: 10.00%)

| Rank | Algoritmo | Coef. Variação | Desvio Padrão | Min% | Max% | Δ% | Qualidade | Score |
|------|-----------|----------------|---------------|------|------|-----|-----------|-------|
| 1 | SHA-1 | 1.55% | 1547.23 | 9.85% | 10.12% | 0.27% | EXCELENTE | 98.45 |
```

## Análise Acadêmica

### Perguntas de Pesquisa

1. **Qual algoritmo oferece melhor uniformidade de distribuição?**
   - Compare os Coeficientes de Variação
   - Analise o Score de Uniformidade

2. **Como o número de shards afeta a distribuição?**
   - Observe a tendência do CV conforme N aumenta
   - Avalie o trade-off entre uniformidade e complexidade

3. **Qual é o equilíbrio ideal entre blast radius e uniformidade?**
   - Cross-reference: CV baixo + Blast Radius aceitável
   - Considere requisitos de disponibilidade

4. **Algoritmos criptográficos são superiores aos não-criptográficos?**
   - Compare SHA-* vs MURMUR3
   - Avalie trade-off segurança vs performance

### Hipóteses Testáveis

- **H1**: Algoritmos criptográficos (SHA-*) oferecem distribuição mais uniforme que não-criptográficos (MURMUR3)
- **H2**: O aumento do número de shards melhora a uniformidade (menor CV)
- **H3**: SHA-1 apresenta melhor equilíbrio entre uniformidade e performance
- **H4**: A partir de N shards, a melhoria de uniformidade se torna marginal

## Estrutura de Diretórios

```
msc-shard-router/
├── cmd/
│   └── hashing-distribution/
│       └── main.go              # Código do benchmark
├── examples/
│   └── hashing-test/
│       ├── 1kk_uuids.txt       # 1M UUIDs (gerado)
│       └── generate_uuids.go   # Gerador de UUIDs
├── scripts/
│   └── run-benchmark.sh        # Script automatizado
├── benchmark-results/           # Resultados gerados
│   ├── *.csv                   # Dados tabulares
│   └── *.md                    # Tabelas formatadas
└── BENCHMARK_GUIDE.md          # Este arquivo
```

## Configurações Recomendadas

### Para Dissertação de Mestrado

```bash
# Benchmark completo para análise estatística
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=3,5,10,25,50,100,200,500,1000 \
    -replicas=200 \
    -csv=mestrado_completo.csv \
    -md=mestrado_completo.md \
    -verbose
```

### Para Análise de Blast Radius

```bash
# Foco em disponibilidade
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=10,50,100,500,1000 \
    -replicas=150 \
    -csv=blast_radius_analysis.csv \
    -md=blast_radius_analysis.md
```

### Para Comparação de Algoritmos

```bash
# Foco em uniformidade com poucos shards
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=3,5,10 \
    -replicas=200 \
    -csv=algorithm_comparison.csv \
    -md=algorithm_comparison.md \
    -verbose
```

## Performance Esperada

### Tempos Aproximados (MacBook Pro M1)

| Configuração | Tempo Estimado |
|--------------|----------------|
| 3 shards | ~1 segundo |
| 10 shards | ~2 segundos |
| 50 shards | ~5 segundos |
| 100 shards | ~10 segundos |
| 1000 shards | ~60 segundos |

**Total para benchmark completo**: ~15-20 minutos

## Troubleshooting

### Arquivo de UUIDs não encontrado

```bash
# Gerar UUIDs manualmente
cd examples/hashing-test
go run generate_uuids.go
```

### Memória insuficiente

Se o sistema não tem memória para 1M UUIDs, reduza o dataset:

```go
// Em generate_uuids.go, altere:
const numUUIDs = 100000  // 100k em vez de 1M
```

### Resultados inconsistentes

Certifique-se de usar o mesmo arquivo de UUIDs entre execuções:
- UUIDs devem ser gerados uma única vez
- Reutilize o mesmo arquivo para comparações

## Contribuindo com Resultados

Para contribuir com a pesquisa:

1. Execute o benchmark completo
2. Salve os resultados CSV
3. Documente seu hardware (CPU, RAM)
4. Compartilhe via pull request ou issue

## Citação

Se utilizar este benchmark em sua pesquisa acadêmica, cite:

```
@mastersthesis{fidelis2025cellular,
  author = {Fidelis, Matheus},
  title = {Roteamento Celular Inteligente: Análise de Algoritmos de Hash para Arquiteturas Distribuídas},
  school = {Universidade},
  year = {2025},
  type = {Dissertação de Mestrado}
}
```

## Licença

MIT License - Veja LICENSE para detalhes

## Contato

Para dúvidas sobre o benchmark ou colaborações acadêmicas:
- GitHub: @msfidelis
- Repository: https://github.com/msfidelis/msc-shard-router
