# 🎓 Benchmark de Distribuição - Arquitetura Celular

## ✨ Implementação Completa para Dissertação de Mestrado

Este documento descreve a implementação completa do benchmark de distribuição de hash para análise acadêmica da arquitetura celular.

---

## 📋 O Que Foi Implementado

### 1. **Código de Benchmark Completo**
   - 📁 `cmd/hashing-distribution/main.go` (atualizado)
   - ✅ Suporte para múltiplos números de shards
   - ✅ 5 algoritmos de hash (SHA-512, SHA-256, SHA-1, MD5, MURMUR3)
   - ✅ Métricas acadêmicas (CV, desvio padrão, variância, score)
   - ✅ Análise de blast radius integrada

### 2. **Scripts de Automação**
   - 📁 `scripts/run-benchmark.sh`
   - ✅ Execução automatizada de múltiplos benchmarks
   - ✅ Geração de relatórios CSV e Markdown
   - ✅ Output colorido e informativo

### 3. **Makefile Integrado**
   - 📁 `Makefile` (atualizado)
   - ✅ Comandos simplificados (`make benchmark-quick`, etc.)
   - ✅ Geração automática de UUIDs
   - ✅ Limpeza de resultados antigos

### 4. **Documentação Completa**
   - 📁 `BENCHMARK_GUIDE.md` - Guia acadêmico completo
   - 📁 `BENCHMARK_QUICKSTART.md` - Início rápido
   - 📁 `README_BENCHMARK.md` - Este arquivo

---

## 🚀 Início Rápido (3 Comandos)

```bash
# 1. Gerar UUIDs (se necessário)
make generate-uuids

# 2. Executar benchmark completo
make benchmark-complete

# 3. Ver resultados
ls -lh benchmark-results/
```

**Tempo estimado**: 15-20 minutos para benchmark completo

---

## 📊 Comandos Principais

### Benchmarks Pré-configurados

| Comando | Shards Testados | Tempo | Uso |
|---------|----------------|-------|-----|
| `make benchmark-quick` | 3, 5, 10 | ~2 min | Testes rápidos |
| `make benchmark-medium` | 10, 25, 50 | ~5 min | Análise intermediária |
| `make benchmark-advanced` | 50, 100, 200 | ~12 min | Escala enterprise |
| `make benchmark-complete` | 3-200 | ~20 min | **Dissertação completa** |
| `make benchmark-full` | Todos | ~40 min | Suite completa |

### Execução Customizada

```bash
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=3,5,10,25,50,100,200 \
    -replicas=150 \
    -csv=results.csv \
    -md=results.md \
    -verbose
```

---

## 📈 Métricas Implementadas

### Métricas de Uniformidade

1. **Coeficiente de Variação (CV)** 
   - Métrica primária de uniformidade
   - CV = (σ / μ) × 100%
   - Quanto menor, melhor

2. **Score de Uniformidade**
   - Métrica normalizada 0-100
   - 100 = distribuição perfeita
   - Score = 100 - min(CV, 100)

3. **Desvio Padrão (σ)**
   - Dispersão absoluta
   - Medida em número de chaves

4. **Variância (σ²)**
   - Quadrado do desvio padrão
   - Útil para análise estatística

### Métricas de Disponibilidade

5. **Blast Radius**
   - Impacto de uma falha
   - BR = (1 / N) × 100%
   - Onde N = número de shards

6. **Disponibilidade**
   - Percentual operacional após falha
   - Disp = ((N-1) / N) × 100%

### Classificação de Qualidade

| CV Range | Classificação | Uso Recomendado |
|----------|---------------|-----------------|
| < 2% | EXCELENTE | ✅ Produção crítica |
| 2-5% | MUITO BOA | ✅ Produção geral |
| 5-10% | BOA | ⚠️ Aceitável |
| 10-20% | REGULAR | ⚠️ Revisar |
| > 20% | RUIM | ❌ Não recomendado |

---

## 📁 Estrutura de Saída

### Diretório de Resultados

```
benchmark-results/
├── quick_20251205_143022.csv      # Dados tabulares
├── quick_20251205_143022.md       # Tabelas formatadas
├── medium_20251205_144515.csv
├── medium_20251205_144515.md
├── complete_20251205_150245.csv   # ⭐ Principal
└── complete_20251205_150245.md    # ⭐ Principal
```

### Formato CSV

```csv
Algoritmo,NumShards,TotalChaves,DesvPadrao,Variancia,CoefVariacao,MinPercent,MaxPercent,Diferenca,Quality,UniformityScore,BlastRadius,Duration
SHA-1,10,1000000,1547.23,2393920.45,1.55,9.85,10.12,0.27,EXCELENTE,98.45,10.00,245ms
```

**Uso**: Importar em Python/R/Excel para análise estatística

### Formato Markdown

```markdown
## Configuração: 10 Shards (Blast Radius: 10.00%)

| Rank | Algoritmo | Coef. Variação | Desvio Padrão | Min% | Max% | Δ% | Qualidade | Score |
|------|-----------|----------------|---------------|------|------|-----|-----------|-------|
| 1 | SHA-1 | 1.55% | 1547.23 | 9.85% | 10.12% | 0.27% | EXCELENTE | 98.45 |
```

**Uso**: Copiar diretamente para o artigo

---

## 🔬 Perguntas de Pesquisa Respondidas

### 1. Qual algoritmo oferece melhor distribuição?

**Comando**:
```bash
make benchmark-quick
```

**Análise**: Compare CV entre algoritmos. Menor CV = melhor.

### 2. Como o número de shards afeta a uniformidade?

**Comando**:
```bash
make benchmark-complete
```

**Análise**: Plote CV vs NumShards. Identifique tendência.

### 3. Qual é o trade-off entre blast radius e uniformidade?

**Comando**:
```bash
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=10,50,100,500,1000 -csv=tradeoff.csv
```

**Análise**: Compare CV vs BlastRadius. Encontre ponto ótimo.

### 4. Quantos shards para 99% disponibilidade?

**Resposta**: 100 shards → 1% blast radius → 99% disponível

**Verificação**:
```bash
go run cmd/hashing-distribution/main.go examples/hashing-test/1kk_uuids.txt \
    -shards=100 -verbose
```

---

## 📊 Análise Estatística Recomendada

### Python (Pandas)

```python
import pandas as pd
import matplotlib.pyplot as plt

# Carregar dados
df = pd.read_csv('benchmark-results/complete_*.csv')

# Análise por algoritmo
print(df.groupby('Algoritmo')['CoefVariacao'].describe())

# Plot: CV vs NumShards
for algo in df['Algoritmo'].unique():
    data = df[df['Algoritmo'] == algo]
    plt.plot(data['NumShards'], data['CoefVariacao'], 
             marker='o', label=algo)

plt.xlabel('Número de Shards')
plt.ylabel('Coeficiente de Variação (%)')
plt.title('Uniformidade vs Escalabilidade')
plt.legend()
plt.grid(True)
plt.savefig('cv_analysis.png', dpi=300)
plt.show()

# Encontrar configuração ótima
best = df.loc[df['UniformityScore'].idxmax()]
print(f"Melhor configuração: {best['Algoritmo']} com {best['NumShards']} shards")
print(f"CV: {best['CoefVariacao']:.2f}%")
print(f"Blast Radius: {best['BlastRadius']:.2f}%")
```

### R

```r
library(tidyverse)

# Carregar dados
df <- read_csv("benchmark-results/complete_*.csv")

# Análise por algoritmo
df %>%
  group_by(Algoritmo) %>%
  summarise(
    MediaCV = mean(CoefVariacao),
    MinCV = min(CoefVariacao),
    MaxCV = max(CoefVariacao)
  )

# Plot: Heatmap
df %>%
  ggplot(aes(x = NumShards, y = Algoritmo, fill = CoefVariacao)) +
  geom_tile() +
  scale_fill_gradient(low = "green", high = "red") +
  labs(title = "Coeficiente de Variação por Configuração",
       x = "Número de Shards",
       y = "Algoritmo",
       fill = "CV (%)") +
  theme_minimal()
```

---

## 📝 Inclusão no Artigo

### Seção de Metodologia

```markdown
### 4.1 Metodologia de Avaliação

Para validar a eficácia dos diferentes algoritmos de hash na distribuição 
uniforme de chaves, foram realizados experimentos com **1 milhão de chaves 
UUID v4** distribuídas entre N shards (N ∈ {3, 5, 10, 25, 50, 100, 200}). 

As chaves UUID v4 foram escolhidas por sua natureza aleatória e 
representatividade em cenários reais de produção. Cada configuração 
utilizou 150 réplicas virtuais por shard para otimizar a distribuição 
no hash ring consistente.

Os critérios de avaliação incluíram:

- **Uniformidade de distribuição**: Medida pelo Coeficiente de Variação (CV)
- **Desvio padrão**: Dispersão absoluta da distribuição
- **Score de uniformidade**: Métrica normalizada (0-100)
- **Blast radius**: Impacto de falha calculado como 1/N × 100%
```

### Seção de Resultados

```markdown
### 4.2 Resultados Experimentais

A Tabela X apresenta os resultados comparativos dos algoritmos testados 
para diferentes configurações de shards.
```

**Cole aqui as tabelas do arquivo `.md` gerado**

---

## ✅ Checklist para Dissertação

- [ ] Executar `make benchmark-complete`
- [ ] Revisar arquivos CSV gerados
- [ ] Copiar tabelas Markdown para o artigo
- [ ] Gerar gráficos (CV vs Shards, Blast Radius)
- [ ] Analisar estatísticas descritivas
- [ ] Identificar configuração ótima
- [ ] Comparar com literatura (Dynamo, Cassandra)
- [ ] Documentar conclusões
- [ ] Revisar metodologia no artigo
- [ ] Incluir referências aos dados brutos

---

## 🔗 Documentação Adicional

- **Guia Completo**: `BENCHMARK_GUIDE.md`
- **Início Rápido**: `BENCHMARK_QUICKSTART.md`
- **Código Fonte**: `cmd/hashing-distribution/main.go`
- **Script Automatizado**: `scripts/run-benchmark.sh`

---

## 💡 Dicas Importantes

1. **Execute uma vez**: Gere UUIDs uma vez, reutilize sempre
2. **Seja consistente**: Use mesmos parâmetros para comparações
3. **Documente hardware**: CPU, RAM influenciam performance
4. **Salve tudo**: Resultados são únicos, backup é essencial
5. **Analise iterativamente**: Comece com quick, depois complete

---

## 🎯 Resultado Esperado

Após executar o benchmark completo, você terá:

✅ **Dados quantitativos** sobre uniformidade de distribuição  
✅ **Análise comparativa** de 5 algoritmos de hash  
✅ **Trade-off documentado** entre blast radius e complexidade  
✅ **Evidência estatística** para recomendações arquiteturais  
✅ **Tabelas e gráficos** prontos para publicação  

---

## 🚀 Próximo Passo

Execute agora:

```bash
make benchmark-complete
```

Aguarde 15-20 minutos e seus dados estarão prontos! 🎓📊

---

**Autor**: Mestrado em Arquitetura Celular  
**Data**: Dezembro 2025  
**Versão**: 1.0
