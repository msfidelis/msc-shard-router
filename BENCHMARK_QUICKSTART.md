# Guia de Execução - Benchmark de Distribuição

## Início Rápido

### 1. Gerar UUIDs (se necessário)

```bash
make generate-uuids
```

Isso irá gerar `examples/hashing-test/1kk_uuids.txt` com 1 milhão de UUIDs v4.

### 2. Executar Benchmark Rápido

```bash
make benchmark-quick
```

Este comando testa 3, 5 e 10 shards e leva aproximadamente 1-2 minutos.

### 3. Ver Resultados

Os resultados são salvos em `benchmark-results/` com timestamp:
- `quick_YYYYMMDD_HHMMSS.csv` - Dados tabulares
- `quick_YYYYMMDD_HHMMSS.md` - Tabelas formatadas

## Comandos Disponíveis

### Benchmarks Individuais

```bash
# Benchmark rápido: 3, 5, 10 shards (1-2 min)
make benchmark-quick

# Benchmark médio: 10, 25, 50 shards (3-5 min)
make benchmark-medium

# Benchmark avançado: 50, 100, 200 shards (8-12 min)
make benchmark-advanced

# Benchmark completo para dissertação (15-20 min)
make benchmark-complete
```

### Benchmark Automatizado

```bash
# Executa todos os benchmarks em sequência
make benchmark-full
```

### Utilitários

```bash
# Verificar se UUIDs existem (gera se necessário)
make benchmark-check

# Limpar resultados antigos
make benchmark-clean

# Gerar novos UUIDs
make generate-uuids
```

## Execução Manual Customizada

### Sintaxe

```bash
go run cmd/hashing-distribution/main.go <arquivo-uuids> [flags]
```

### Exemplos Práticos

#### 1. Análise de Blast Radius

Testar configurações focadas em disponibilidade:

```bash
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=10,50,100,500,1000 \
    -replicas=150 \
    -csv=blast_radius_analysis.csv \
    -md=blast_radius_analysis.md
```

**Objetivo**: Analisar relação entre número de shards e blast radius (1/N).

#### 2. Comparação de Algoritmos

Testar algoritmos com poucos shards:

```bash
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=3,5,10 \
    -replicas=200 \
    -csv=algorithm_comparison.csv \
    -md=algorithm_comparison.md \
    -verbose
```

**Objetivo**: Identificar qual algoritmo oferece melhor distribuição.

#### 3. Análise de Escalabilidade

Testar muitos shards:

```bash
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=100,200,500,1000 \
    -replicas=150 \
    -csv=scalability_test.csv \
    -md=scalability_test.md
```

**Objetivo**: Avaliar comportamento em escala enterprise.

#### 4. Teste de Réplicas Virtuais

Comparar diferentes números de réplicas:

```bash
# Poucas réplicas
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=10 -replicas=50 -csv=replicas_50.csv

# Muitas réplicas  
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=10 -replicas=200 -csv=replicas_200.csv
```

**Objetivo**: Avaliar impacto do número de réplicas virtuais na uniformidade.

## Interpretando Resultados

### Output do Terminal

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
    ├─ Performance: 245ms
    └─ Detalhes por shard:
       ├─ shard01: 99850 (9.98%) [Desvio: 150 / 0.2%]
       ├─ shard02: 100120 (10.01%) [Desvio: 120 / 0.1%]
       ...
```

### Métricas Chave

1. **Coeficiente de Variação (CV)**: Principal métrica
   - Menor CV = melhor uniformidade
   - CV < 2% é EXCELENTE para produção

2. **Score de Uniformidade**: Métrica normalizada (0-100)
   - 100 = distribuição perfeita
   - > 95 = excelente para produção

3. **Distribuição Min-Max**: Faixa percentual
   - Δ pequeno = boa distribuição
   - Δ < 0.5% é ideal

4. **Blast Radius**: Implícito no número de shards
   - 10 shards = 10% blast radius
   - 100 shards = 1% blast radius

### Arquivo CSV

Estrutura do CSV gerado:

```csv
Algoritmo,NumShards,TotalChaves,DesvPadrao,Variancia,CoefVariacao,MinPercent,MaxPercent,Diferenca,Quality,UniformityScore,BlastRadius,Duration
SHA-1,10,1000000,1547.23,2393920.45,1.55,9.85,10.12,0.27,EXCELENTE,98.45,10.00,245ms
SHA-256,10,1000000,2341.67,5483419.11,2.34,9.72,10.35,0.63,MUITO BOA,97.66,10.00,198ms
```

**Uso**: Importar no Excel, Python (pandas), R para análise estatística.

### Arquivo Markdown

Tabelas formatadas prontas para o artigo:

```markdown
## Configuração: 10 Shards (Blast Radius: 10.00%)

| Rank | Algoritmo | Coef. Variação | Desvio Padrão | Min% | Max% | Δ% | Qualidade | Score |
|------|-----------|----------------|---------------|------|------|-----|-----------|-------|
| 1 | SHA-1 | 1.55% | 1547.23 | 9.85% | 10.12% | 0.27% | EXCELENTE | 98.45 |
| 2 | SHA-256 | 2.34% | 2341.67 | 9.72% | 10.35% | 0.63% | MUITO BOA | 97.66 |
```

**Uso**: Copiar diretamente para o artigo LaTeX/Markdown.

## Análise para Dissertação

### 1. Análise de Uniformidade

**Pergunta**: Qual algoritmo tem melhor distribuição?

```bash
make benchmark-quick
```

**Análise**:
- Compare CV entre algoritmos
- Escolha o de menor CV
- Verifique se performance é aceitável

### 2. Análise de Blast Radius

**Pergunta**: Quantos shards para 99% disponibilidade?

```bash
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=50,100,200,500 \
    -csv=blast_analysis.csv
```

**Análise**:
- 100 shards = 1% blast radius = 99% disponível
- Avaliar trade-off CV vs blast radius

### 3. Análise de Escalabilidade

**Pergunta**: A distribuição melhora com mais shards?

```bash
make benchmark-complete
```

**Análise**:
- Plotar CV vs NumShards
- Identificar ponto de retorno decrescente
- Determinar configuração ótima

### 4. Comparação com Literatura

**Benchmark de referência**:

| Estudo | Algoritmo | CV | Shards |
|--------|-----------|-----|--------|
| DeCandia et al. (2007) | MD5 | ~3% | 100 |
| Este trabalho | SHA-1 | ~1.5% | 100 |

## Gráficos Recomendados

### 1. CV vs Número de Shards

```python
import pandas as pd
import matplotlib.pyplot as plt

df = pd.read_csv('benchmark_complete.csv')

for algo in df['Algoritmo'].unique():
    data = df[df['Algoritmo'] == algo]
    plt.plot(data['NumShards'], data['CoefVariacao'], 
             marker='o', label=algo)

plt.xlabel('Número de Shards')
plt.ylabel('Coeficiente de Variação (%)')
plt.title('Uniformidade vs Número de Shards')
plt.legend()
plt.grid(True)
plt.savefig('cv_vs_shards.png', dpi=300)
```

### 2. Score de Uniformidade

```python
pivot = df.pivot(index='NumShards', 
                 columns='Algoritmo', 
                 values='UniformityScore')

pivot.plot(kind='bar', figsize=(12, 6))
plt.ylabel('Score de Uniformidade')
plt.title('Comparação de Algoritmos por Configuração')
plt.savefig('uniformity_comparison.png', dpi=300)
```

### 3. Blast Radius vs Disponibilidade

```python
df['Disponibilidade'] = 100 - df['BlastRadius']

plt.figure(figsize=(10, 6))
plt.plot(df['NumShards'], df['BlastRadius'], 
         'r-', marker='o', label='Blast Radius')
plt.plot(df['NumShards'], df['Disponibilidade'], 
         'g-', marker='s', label='Disponibilidade')

plt.xlabel('Número de Shards')
plt.ylabel('Percentual (%)')
plt.title('Trade-off: Blast Radius vs Disponibilidade')
plt.legend()
plt.grid(True)
plt.savefig('blast_radius_tradeoff.png', dpi=300)
```

## Troubleshooting

### Erro: Arquivo não encontrado

```bash
# Solução
make generate-uuids
```

### Erro: Out of memory

Reduza o dataset ou aumente swap:

```bash
# Usar 100k UUIDs em vez de 1M
head -n 100000 examples/hashing-test/1kk_uuids.txt > test_100k.txt
go run cmd/hashing-distribution/main.go test_100k.txt -shards=3,5,10
```

### Performance lenta

Reduza réplicas virtuais:

```bash
# Usar menos réplicas (trade-off: menos uniforme)
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=100 -replicas=50
```

## Próximos Passos

1. ✅ Execute `make benchmark-complete`
2. ✅ Analise os arquivos `.csv` gerados
3. ✅ Copie tabelas `.md` para o artigo
4. ✅ Gere gráficos com Python/R
5. ✅ Interprete resultados à luz da literatura
6. ✅ Documente conclusões

## Suporte

Para dúvidas ou problemas:
- Abra uma issue no GitHub
- Consulte `BENCHMARK_GUIDE.md` para documentação completa
- Revise o código em `cmd/hashing-distribution/main.go`
