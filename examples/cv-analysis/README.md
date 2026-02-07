# Análise de Coeficiente de Variação vs Número de Shards

Este script analisa o comportamento do **Coeficiente de Variação (CV%)** em relação ao **número de shards** para diferentes algoritmos de hash, permitindo visualizar como a uniformidade da distribuição varia conforme o número de células aumenta.

## Objetivo

Gerar dados tabelados e em CSV para criar gráficos que ilustrem o comportamento do CV% mediante ao número de células (shards) no contexto de arquitetura celular com consistent hashing.

## Algoritmos Analisados

- **MURMUR3**: Hash não-criptográfico de alta performance
- **SHA-1**: Hash criptográfico SHA-1 (160 bits)
- **SHA-256**: Hash criptográfico SHA-256 (256 bits)
- **SHA-512**: Hash criptográfico SHA-512 (512 bits)

## Uso

### Execução Básica

```bash
go run cmd/cv-analysis/main.go examples/hashing-test/1kk_uuids.txt
```

### Parâmetros Disponíveis

```bash
go run cmd/cv-analysis/main.go <arquivo_uuids> [flags]

Flags:
  -min int        Número mínimo de shards (default: 3)
  -max int        Número máximo de shards (default: 20)
  -replicas int   Número de réplicas virtuais por shard (default: 150)
  -output string  Arquivo de saída CSV (default: "cv_analysis.csv")
```

### Exemplos

**Análise de 3 a 20 shards:**
```bash
go run cmd/cv-analysis/main.go examples/hashing-test/1kk_uuids.txt \
    -min=3 -max=20 -replicas=150
```

**Análise de 3 a 50 shards:**
```bash
go run cmd/cv-analysis/main.go examples/hashing-test/1kk_uuids.txt \
    -min=3 -max=50 -replicas=150 -output=cv_analysis_50.csv
```

**Análise com menos réplicas:**
```bash
go run cmd/cv-analysis/main.go examples/hashing-test/1kk_uuids.txt \
    -min=3 -max=30 -replicas=100
```

## Output

### 1. Tabela Detalhada por Algoritmo

Para cada algoritmo, exibe:
- Número de shards
- Coeficiente de Variação (CV%)
- Desvio Padrão
- Média de chaves por shard
- Delta (Amplitude)
- Percentuais mínimo e máximo

Exemplo:
```
┌─ MURMUR3 ──────────────────────────────────────────────────────────────
│
│  Shards | CV%      | Desvio Padrão | Média       | Delta (Amp.)     
│  ───────────────────────────────────────────────────────────────────
│  3      |    6.16 |      20548.44 |   333333.33 |  49328 ( 4.93%) 
│  5      |    4.05 |       8103.22 |   200000.00 |  24576 ( 2.46%) 
│  10     |    8.72 |       8721.35 |   100000.00 |  26896 ( 2.69%) 
│
└───────────────────────────────────────────────────────────────────────
```

### 2. Tabela Comparativa

Tabela lado-a-lado comparando CV% de todos os algoritmos:

```
Shards   | MURMUR3      | SHA-1        | SHA-256      | SHA-512     
-------------------------------------------------------------------------
3        |       6.16% |       8.71% |       7.68% |       6.89%
5        |       4.05% |      11.69% |       5.95% |       5.44%
10       |       8.72% |       7.50% |       9.46% |      10.05%
```

### 3. Arquivo CSV

Formato do CSV gerado:

```csv
Algorithm,NumShards,CV%,StdDev,Mean,MinPercent,MaxPercent,Delta,DeltaPercent,TotalKeys
MURMUR3,3,6.1623,20548.44,333333.33,16.53,21.47,49328,4.9328,1000000
MURMUR3,5,4.0516,8103.22,200000.00,18.92,21.38,24576,2.4576,1000000
...
```

## Uso para Gráficos

### Com Excel/Google Sheets

1. Importe o arquivo `cv_analysis.csv`
2. Crie um gráfico de linhas com:
   - **Eixo X**: NumShards
   - **Eixo Y**: CV%
   - **Séries**: Uma linha por Algorithm

### Com Python/Matplotlib

```python
import pandas as pd
import matplotlib.pyplot as plt

# Carregar dados
df = pd.read_csv('cv_analysis.csv')

# Plotar
fig, ax = plt.subplots(figsize=(12, 6))

for algo in ['MURMUR3', 'SHA-1', 'SHA-256', 'SHA-512']:
    data = df[df['Algorithm'] == algo]
    ax.plot(data['NumShards'], data['CV%'], marker='o', label=algo)

ax.set_xlabel('Número de Shards')
ax.set_ylabel('Coeficiente de Variação (%)')
ax.set_title('CV% vs Número de Shards por Algoritmo')
ax.legend()
ax.grid(True, alpha=0.3)
plt.tight_layout()
plt.savefig('cv_analysis.png', dpi=300)
```

### Com R/ggplot2

```r
library(ggplot2)
library(readr)

# Carregar dados
df <- read_csv('cv_analysis.csv')

# Plotar
ggplot(df, aes(x = NumShards, y = `CV%`, color = Algorithm)) +
  geom_line(size = 1) +
  geom_point(size = 2) +
  labs(
    title = "Coeficiente de Variação vs Número de Shards",
    x = "Número de Shards",
    y = "Coeficiente de Variação (%)"
  ) +
  theme_minimal() +
  theme(legend.position = "bottom")

ggsave("cv_analysis.png", width = 12, height = 6, dpi = 300)
```

## Interpretação dos Resultados

### Coeficiente de Variação (CV%)

- **CV < 5%**: Distribuição muito uniforme (EXCELENTE)
- **CV 5-10%**: Distribuição boa (BOA)
- **CV 10-15%**: Distribuição aceitável (REGULAR)
- **CV > 15%**: Distribuição ruim (INADEQUADA)

### Observações Esperadas

1. **Tendência Geral**: CV% pode variar não-linearmente com número de shards
2. **Algoritmos Não-Criptográficos**: MURMUR3 tende a ter melhor uniformidade
3. **Algoritmos Criptográficos**: SHA-256/512 podem ter overhead maior mas distribuição aceitável
4. **Pontos Ótimos**: Identificar número de shards onde CV% é minimizado

## Aplicação Acadêmica

Este script é útil para:

- **Dissertações**: Gerar gráficos comparativos de algoritmos
- **Análise de Escalabilidade**: Identificar como distribuição degrada/melhora com escala
- **Trade-offs**: Balancear performance vs uniformidade vs número de células
- **Validação Teórica**: Comparar resultados empíricos com modelos matemáticos

## Referências

- Consistent Hashing and Random Trees (Karger et al., 1997)
- NIST FIPS 180-4: Secure Hash Standard (SHA)
- MurmurHash3 (Austin Appleby, 2016)
