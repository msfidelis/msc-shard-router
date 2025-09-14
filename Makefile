# Makefile para o MSC Shard Router

.PHONY: test test-verbose test-coverage build run clean help docker-build docker-run docker-compose-up docker-compose-down lint security ci

# Configurações
BINARY_NAME=shard-router
DOCKER_IMAGE=msc-shard-router
PORT=8080

# Cores para output
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Mostra esta mensagem de ajuda
	@echo "$(GREEN)MSC Shard Router - Comandos disponíveis:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

test: ## Executa todos os testes unitários
	@echo "$(GREEN)Executando testes unitários...$(NC)"
	@go test ./... -v

test-coverage: ## Executa testes com coverage
	@echo "$(GREEN)Executando testes com coverage...$(NC)"
	@go test ./... -v -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report gerado em coverage.html$(NC)"

test-verbose: ## Executa testes com output verboso
	@echo "$(GREEN)Executando testes verbosos...$(NC)"
	@go test ./... -v -count=1

build: ## Compila o binário
	@echo "$(GREEN)Compilando $(BINARY_NAME)...$(NC)"
	@go mod tidy
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BINARY_NAME) .
	@echo "$(GREEN)Binário compilado: $(BINARY_NAME)$(NC)"


test-integration: docker-compose-up ## Executa testes de integração
	@echo "$(GREEN)Aguardando serviços subirem...$(NC)"
	@sleep 5
	@echo "$(GREEN)Testando health check...$(NC)"
	@curl -f http://localhost:9090/healthz || (echo "$(RED)Health check falhou$(NC)" && exit 1)
	@echo "$(GREEN)Testando roteamento...$(NC)"
	@curl -H "id_client: test123" http://localhost:9090/ || true
	@echo "$(GREEN)Testando métricas...$(NC)"
	@curl -f http://localhost:9090/metrics | grep -q "shard_router" || (echo "$(RED)Métricas não encontradas$(NC)" && exit 1)
	@echo "$(GREEN)Testes de integração concluídos$(NC)"
	@$(MAKE) docker-compose-down

benchmark: ## Executa benchmarks
	@echo "$(GREEN)Executando benchmarks...$(NC)"
	@go test ./... -bench=. -benchmem

ci: lint test ## Pipeline de CI (format, lint, test)
	@echo "$(GREEN)Pipeline de CI concluída com sucesso!$(NC)"


ci-full: lint security test build ## Pipeline de CI completa com segurança
	@echo "$(GREEN)Pipeline de CI completa concluída!$(NC)"

lint-install: ## Instala golangci-lint
	@echo "$(GREEN)Instalando golangci-lint...$(NC)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

lint: ## Executa linting do código
	@echo "$(GREEN)Executando lint...$(NC)"
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "$(YELLOW)golangci-lint não encontrado. Instalando...$(NC)"; \
		$(MAKE) lint-install; \
	fi
	@golangci-lint run --timeout=5m

security: ## Executa verificações de segurança
	@echo "$(GREEN)Executando verificações de segurança...$(NC)"
	@if ! command -v gosec &> /dev/null; then \
		echo "$(YELLOW)Instalando gosec...$(NC)"; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	@gosec ./...
	@if ! command -v govulncheck &> /dev/null; then \
		echo "$(YELLOW)Instalando govulncheck...$(NC)"; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	@govulncheck ./...

.DEFAULT_GOAL := help