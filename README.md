# MSC Notes - Shard Router

[![CI/CD Pipeline](https://github.com/msfidelis/msc-shard-router/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/msfidelis/msc-shard-router/actions/workflows/ci-cd.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/msfidelis/msc-shard-router)](https://go.dev/)
[![Coverage](https://img.shields.io/badge/coverage-87%25-green)](https://github.com/msfidelis/msc-shard-router/actions)
[![Security Scan](https://img.shields.io/badge/security-passing-green)](https://github.com/msfidelis/msc-shard-router/security)
[![Performance](https://img.shields.io/badge/performance-tested-blue)](https://github.com/msfidelis/msc-shard-router/actions/workflows/performance.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Um proxy router baseado em hash consistente para distribuição de requisições entre shards, desenvolvido como parte das POCs do mestrado em arquitetura celular.

## Visão Geral

O **MSC Shard Router** é um componente essencial em arquiteturas distribuídas que implementa os conceitos de:

- **Sharding**: Particionamento horizontal de dados/serviços
- **Hash Consistente**: Distribuição uniforme e estável de chaves entre shards
- **Proxy/Load Balancer**: Roteamento transparente e metrificável de requisições
- **Bulkheads**: Isolamento entre componentes através de shards independentes
- **Resiliência**: Tolerância a falhas através de distribuição de carga consistente e isolamento

## Arquitetura

O projeto implementa um padrão de proxy reverso que utiliza hash consistente para determinar o shard de destino baseado em um header HTTP específico:

```
Cliente → Shard Router → [Hash Consistente] → Shard N
```

### Arquitetura Detalhada

```mermaid
graph TB
    subgraph "Cliente"
        C1[Aplicação Cliente - Header: id_client]
    end
    
    subgraph "Shard Router"
        SR1[HTTP Server :8080]
        SR2[Extrair valor do Header: id_client]
        SR3[Calcula o Hash Ring Engine]
        SR4[Proxy Reverso para o shard correto]
        SR5[Métricas Prometheus]
        
        SR1 --> SR2
        SR2 --> SR3
        SR3 --> SR4
        SR4 --> SR5
    end
    
    subgraph "Shards Backend"
        S1[Shard 01:8080]
        S2[Shard 02:8080]
        S3[Shard 03:8080]
        SN[Shard N:8080]
    end
        
    C1 --> SR1
    
    SR4 --> S1
    SR4 --> S2
    SR4 --> S3
    SR4 --> SN
```

## 🔧 Configuração

### Docker Image 

> Buildx arm64, amd64

```bash
docker pull ghcr.io/msfidelis/msc-shard-router:latest
```

### Variáveis de Ambiente

| Variável | Descrição | Exemplo | Padrão |
|----------|-----------|---------|---------|
| `ROUTER_PORT` | Porta do servidor router | `8080` | `8080` |
| `SHARDING_KEY` | Nome do header HTTP usado como shard key | `id_client` | `id_client` |
| `HASHING_ALGORITHM` | Algoritmo de hash para consistent hashing | `SHA1, SHA256, SHA512, MURMUR3` | `SHA512` |
| `SHARD_01_URL` | URL do primeiro shard | `http://shard01:80` | - |
| `SHARD_02_URL` | URL do segundo shard | `http://shard02:80` | - |
| `SHARD_N_URL` | URLs adicionais seguindo o padrão | `http://shardN:80` | - |
| `CB_FAILURE_THRESHOLD` | Falhas consecutivas para abrir o circuito | `5` | `5` |
| `CB_OPEN_TIMEOUT_SEC` | Tempo em segundos para tentar half-open | `30` | `30` |
| `CB_HALF_OPEN_SUCCESS_THRESHOLD` | Sucessos consecutivos em half-open para fechar | `2` | `2` |

### Algoritmos de Hash Suportados

| Algoritmo | Variável | Segurança | Performance | Recomendação |
|-----------|----------|-----------|-------------|--------------|
| **SHA-512** | `SHA512` | 🔒 Máxima | ⚡ Boa | 🎯 **Produção** |
| **SHA-256** | `SHA256` | 🔒 Alta | ⚡ Muito Boa | 🚀 **Performance** |
| **SHA-1** | `SHA1` | ⚠️ Moderada | ⚡ Boa | 🧪 **Legado** |
| **MD5** | `MD5` | ❌ Baixa | ⚡ Muito Boa | 🧪 **Desenvolvimento** |
| **Murmur3** | `MURMUR` | ❌ Nenhuma | 🚀 Máxima | ⚡ **Não-criptográfico** |

**Exemplo de configuração:**
```bash
export HASHING_ALGORITHM=SHA256  # Para melhor performance
export HASHING_ALGORITHM=MURMUR  # Para máxima velocidade
export HASHING_ALGORITHM=SHA512  # Para máxima segurança (padrão)
```


### Descoberta Dinâmica de Shards

O sistema automaticamente descobre shards através de regex pattern matching das variáveis de ambiente que seguem o padrão `SHARD_(\d+)_URL`.

## Algoritmo de Hash Consistente

### Implementação

O sistema utiliza **SHA-512** para geração de hashes, convertidos para `uint64` para posicionamento no anel. Características:

- **Réplicas Virtuais**: Cada shard físico possui múltiplas réplicas virtuais no anel
- **Distribuição Uniforme**: Minimiza hotspots através de múltiplos pontos no anel
- **Busca Binária**: Localização eficiente O(log n) do shard de destino

### Fluxo de Roteamento

1. **Extração**: Captura do valor do header definido em `SHARDING_KEY`
2. **Hashing**: Cálculo SHA-512 do valor + conversão para uint64
3. **Lookup**: Busca binária no anel ordenado pelo hash
4. **Roteamento**: Proxy da requisição para o shard selecionado

### Diagrama do Hash Consistente

#### 1. Fluxo de Processamento da Requisição

```mermaid
graph TD
    A[Requisição HTTP] --> B{Header SHARDING_KEY existe?}
    B -->|Não| C[Erro: Header obrigatório]
    B -->|Sim| D[Extrair valor do header]
    
    D --> E[Hash SHA-512 do valor]
    E --> F[Converter para uint64]
    F --> G[Busca binária no anel hash]
    
    G --> H{Posição encontrada?}
    H -->|Não encontrada| I[Retorna primeiro nó do anel]
    H -->|Encontrada| J[Retorna nó na posição]
    
    I --> K[Proxy para shard selecionado]
    J --> K
    
    K --> L[Registrar métricas]
    L --> M[Retornar resposta]
    
    style E fill:#e3f2fd
    style G fill:#f3e5f5
    style K fill:#e8f5e8
```

#### 2. Estrutura das Réplicas Virtuais

```mermaid
graph TB
    subgraph "Shards Físicos"
        SA[Shard A<br/>shard01:80]
        SB[Shard B<br/>shard02:80] 
        SC[Shard C<br/>shard03:80]
    end
    
    subgraph "Réplicas Virtuais no Hash Ring"
        SA --> RA1[A-0: hash_A0]
        SA --> RA2[A-1: hash_A1]
        SA --> RA3[A-2: hash_A2]
        
        SB --> RB1[B-0: hash_B0]
        SB --> RB2[B-1: hash_B1]
        SB --> RB3[B-2: hash_B2]
        
        SC --> RC1[C-0: hash_C0]
        SC --> RC2[C-1: hash_C1]
        SC --> RC3[C-2: hash_C2]
    end
    
    style SA fill:#ffebee
    style SB fill:#e8f5e8
    style SC fill:#e3f2fd
```

#### 3. Anel Hash Consistente (Visão Circular)

```mermaid
graph LR
    subgraph "Anel Hash Ordenado por Valor uint64"
        direction TB
        H1[hash_A0: 1234...] --> H2[hash_B1: 2456...]
        H2 --> H3[hash_C0: 3789...]
        H3 --> H4[hash_A1: 4567...]
        H4 --> H5[hash_B2: 5890...]
        H5 --> H6[hash_C1: 6123...]
        H6 --> H7[hash_A2: 7456...]
        H7 --> H8[hash_B0: 8789...]
        H8 --> H9[hash_C2: 9012...]
        H9 --> H1
    end
    
    subgraph "Exemplo de Lookup"
        KEY[user123<br/>hash: 5500...] -.-> H6
        H6 -.-> RESULT[Rota para Shard C]
    end
    
    style KEY fill:#fff3e0
    style H6 fill:#e8f5e8
    style RESULT fill:#e8f5e8
```

### Algoritmo de Distribuição

O hash consistente implementado segue os seguintes princípios:

1. **Múltiplas Réplicas Virtuais**: Cada shard físico é representado por múltiplas posições no anel hash
2. **Distribuição Uniforme**: As réplicas virtuais minimizam hotspots e garantem distribuição equilibrada
3. **Estabilidade**: Adição/remoção de shards afeta apenas os nós adjacentes no anel
4. **Eficiência**: Busca binária O(log n) para localização do shard de destino

#### Processo de Inicialização do Hash Ring

```mermaid
flowchart TD
    A[Descobrir Shards via ENV] --> B[Criar Hash Ring vazio]
    B --> C[Para cada Shard encontrado]
    C --> D[Gerar N réplicas virtuais]
    D --> E[Calcular hash SHA-512 + índice]
    E --> F[Adicionar réplica ao anel]
    F --> G{Mais shards?}
    G -->|Sim| C
    G -->|Não| H[Ordenar anel por hash]
    H --> I[Hash Ring pronto]
    
    style A fill:#e3f2fd
    style I fill:#e8f5e8
```

#### Processo de Roteamento de Requisições

```mermaid
flowchart TD
    A[Request + Header] --> B[Extrair chave de sharding]
    B --> C[Calcular hash SHA-512]
    C --> D[Converter para uint64]
    D --> E[Busca binária no anel]
    E --> F{Encontrou posição >= hash?}
    F -->|Sim| G[Retornar shard da posição]
    F -->|Não| H[Retornar primeiro shard do anel]
    G --> I[Fazer proxy da requisição]
    H --> I
    I --> J[Registrar métricas]
    
    style C fill:#fff3e0
    style E fill:#f3e5f5
    style I fill:#e8f5e8
```

## Endpoints

### Proxy Principal
- **Endpoint**: `/*` - Aceitando qualquer path ou método, todos os componentes do request serão repassados para o shard
- **Método**: Todos os métodos HTTP
- **Funcionalidade**: Roteamento baseado em hash consistente

### Health Check
- **Endpoint**: `/healthz`
- **Método**: GET
- **Resposta**: Status 200 OK

### Métricas Prometheus
- **Endpoint**: `/metrics`
- **Método**: GET
- **Métricas Disponíveis**:
  - `shard_router_requests_total`: Contador de requisições por shard
  - `shard_router_responses_total`: Contador de respostas por shard e status

## Monitoramento

### Métricas Prometheus

```prometheus
# Requisições totais por shard
shard_router_requests_total{shard="http://shard01:80"}

# Respostas por shard e código de status
shard_router_responses_total{shard="http://shard01:80",status="200"}
```

### Logs Estruturados

O sistema produz logs estruturados incluindo:
- Mapeamento de shards durante inicialização
- Roteamento de chaves para hosts específicos
- Status de saúde do servidor

## Exemplo de Uso

```bash
# Requisição com header de sharding
curl -H "id_client: user123" http://localhost:9090/

# A requisição será sempre roteada para o mesmo shard baseado no hash de "user123"
```

## Conceitos Acadêmicos Implementados

### Arquitetura Celular
- **Isolamento**: Cada shard opera independentemente
- **Escalabilidade**: Adição dinâmica de novos shards
- **Tolerância a Falhas**: Falha de um shard não afeta outros

### Bulkheads Pattern
- **Compartimentalização**: Recursos isolados por shard
- **Contenção de Falhas**: Problemas localizados não se propagam

### Consistent Hashing
- **Estabilidade**: Mudanças mínimas na distribuição ao adicionar/remover shards
- **Performance**: Lookup O(log n) com distribuição uniforme

## Stack Utilizada

- **Go 1.25**: Runtime e linguagem
- **Gorilla Mux**: Roteamento HTTP
- **Prometheus**: Métricas e observabilidade
- **Docker**: Containerização
- **Air**: Hot reload para desenvolvimento

## 🏗️ Arquitetura Testável

O projeto foi refatorado seguindo princípios de **Clean Architecture** e **Dependency Injection** para maximizar a testabilidade:

### Interfaces e Abstrações

```go
// Principais interfaces para testabilidade
type HashRing interface {
    AddNode(nodeID string)
    GetNode(key string) string
}

type ShardRouter interface {
    GetShardingKey(r *http.Request) string
    GetShardHost(key string) string
    InitHashRing(size int)
    AddShard(shardHost string)
}

type ConfigManager interface {
    LoadShards() ([]Shard, error)
    GetShardingKey() string
}
```

### Benefícios da Arquitetura

- **Testabilidade**: Mocking fácil de dependências através de interfaces
- **Injeção de Dependência**: Componentes desacoplados e testáveis isoladamente
- **Separation of Concerns**: Cada package tem responsabilidade única
- **Facilidade de Manutenção**: Código modular e bem estruturado

### Padrões Implementados

- **Repository Pattern**: Para configuração de shards
- **Strategy Pattern**: Para algoritmos de hash
- **Dependency Injection**: Para testabilidade
- **Interface Segregation**: Interfaces pequenas e focadas


## Execução Local

### Docker Compose (Recomendado)

```bash
# Subir todos os serviços
make docker-compose-up

# Ou manualmente:
docker-compose up -d
```

### Build Manual

```bash
# Build da aplicação
make build

# Ou manualmente:
go mod tidy
go build -o shard-router .

# Configuração das variáveis
export ROUTER_PORT=8080
export SHARDING_KEY=id_client
export SHARD_01_URL=http://localhost:8081
export SHARD_02_URL=http://localhost:8082
export SHARD_03_URL=http://localhost:8083

# Execução
make run
```

## Testes

O projeto possui uma suite completa de testes unitários com alta cobertura:

### Executar Testes

```bash
# Todos os testes
make test

# Testes com coverage
make test-coverage

# Testes verbosos
make test-verbose

# Benchmarks
make benchmark
```

### Estrutura de Testes

- **`pkg/hashring/main_test.go`**: Testes do algoritmo de hash consistente
- **`pkg/sharding/main_test.go`**: Testes do roteamento de shards
- **`pkg/setup/main_test.go`**: Testes da configuração e descoberta de shards
- **`main_test.go`**: Testes dos handlers HTTP e integração


### Testes de Integração

```bash
# Testa o sistema completo com Docker Compose
make test-integration
```

## Referências Acadêmicas

- [Consistent Hashing and Random Trees](https://www.akamai.com/us/en/multimedia/documents/technical-publication/consistent-hashing-and-random-trees-distributed-caching-protocols-for-relieving-hot-spots-on-the-world-wide-web-technical-publication.pdf)
- [Building Microservices - Sam Newman](https://samnewman.io/books/building_microservices/)
- [Site Reliability Engineering - Google](https://sre.google/books/)


## Contribuição

Este projeto faz parte de uma pesquisa acadêmica de mestrado sobre arquitetura celular. Contribuições e discussões sobre os conceitos implementados são bem-vindas.