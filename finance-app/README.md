# 💰 Finance App — GoLang

Aplicação de finanças pessoais construída em **Go**, com foco em aprendizado de conceitos avançados de programação e as tecnologias usadas no dia a dia de um desenvolvedor back-end.

## 🎯 Objetivo

Explorar na prática:
- **Goroutines** e **channels** para processamento assíncrono de transações
- **Maps** e estruturas de dados nativas do Go
- Interfaces e **composição** ao invés de herança (OO em Go)
- Camadas de arquitetura: domain → service → repository → handler

## 🛠️ Tecnologias

| Tech | Uso |
|------|-----|
| **Go 1.24** | Linguagem principal |
| **PostgreSQL** | Persistência principal |
| **Redis** | Cache de contas e sessões |
| **Apache Kafka** | Publicação de eventos (account.created, transaction.created) |
| **gRPC** | API interna / inter-serviços |
| **GraphQL** | API de leitura flexível |
| **REST (Gin)** | API HTTP principal |
| **Docker Compose** | Ambiente local completo |

## 📁 Estrutura

```
finance-app/
├── cmd/api/            # Entrypoint da aplicação
├── internal/
│   ├── domain/         # Entidades de negócio + interfaces de repositório
│   ├── service/        # Lógica de negócio (goroutines, channels aqui!)
│   ├── repository/
│   │   ├── postgres/   # Implementação SQL (database/sql)
│   │   └── redis/      # Cache com go-redis
│   ├── handler/
│   │   ├── rest/       # Handlers Gin + router
│   │   ├── grpc/       # Servidor gRPC
│   │   └── graphql/    # Schema + resolvers GraphQL
│   └── event/kafka/    # Producer e Consumer Kafka
├── migrations/         # SQL de criação das tabelas
├── proto/              # Definição .proto do gRPC
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## 🚀 Como rodar

### Pré-requisitos
- Docker + Docker Compose
- Go 1.24+

### 1. Subir infraestrutura
```bash
make docker-up
```
Isso sobe: **PostgreSQL**, **Redis**, **Kafka** + **Zookeeper** e a própria API.

### 2. Rodar localmente (sem Docker para a API)
```bash
# Variáveis de ambiente (opcional, há defaults)
export POSTGRES_DSN="host=localhost user=postgres password=postgres dbname=financeapp port=5432 sslmode=disable"
export REDIS_ADDR="localhost:6379"
export KAFKA_BROKERS="localhost:9092"

make run
```

### 3. Rodar os testes
```bash
make test
```

---

## 📡 API REST

Base URL: `http://localhost:8080/api/v1`

### Contas

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/accounts` | Criar conta |
| GET | `/accounts/:id` | Buscar por ID |
| GET | `/accounts?user_id=X` | Listar por usuário |
| PUT | `/accounts/:id` | Atualizar |
| DELETE | `/accounts/:id` | Remover |

**Exemplo — criar conta:**
```json
POST /api/v1/accounts
{
  "user_id": 1,
  "name": "Conta Corrente",
  "type": "checking",
  "currency": "BRL"
}
```

### Transações

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/transactions` | Criar transação (síncrono) |
| POST | `/transactions/async` | Criar via worker pool (goroutine) |
| GET | `/transactions/:id` | Buscar por ID |
| GET | `/transactions?account_id=X` | Listar com filtros |
| GET | `/transactions/summary/:account_id` | Resumo receita/despesa |

**Tipos de transação:** `income`, `expense`, `transfer`

**Exemplo — criar despesa:**
```json
POST /api/v1/transactions
{
  "account_id": 1,
  "type": "expense",
  "amount": 150.50,
  "description": "Mercado",
  "category_id": 2
}
```

### Categorias

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/categories` | Criar categoria |
| GET | `/categories?user_id=X` | Listar categorias |

### Orçamentos

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/budgets` | Criar orçamento |
| GET | `/budgets?user_id=X&year=2024&month=3` | Listar orçamentos do mês |

---

## 🔮 GraphQL

Endpoint: `POST /api/v1/graphql`

```graphql
# Buscar conta
query {
  account(id: 1) {
    id name balance currency
  }
}

# Resumo financeiro
query {
  summary(account_id: 1) {
    income expense net
  }
}

# Criar conta
mutation {
  createAccount(user_id: 1, name: "Poupança", type: "savings") {
    id balance
  }
}
```

---

## 🔗 gRPC

Porta: `9090`

Serviços disponíveis (ver `proto/finance.proto`):
- `AccountService.GetAccount`
- `AccountService.ListAccounts`
- `TransactionService.CreateTransaction`
- `TransactionService.GetSummary`

---

## ⚡ Conceitos Go explorados

### Goroutines + Channels no Worker Pool
```go
// TransactionService usa um pool de goroutines para processar transações
func (s *TransactionService) startWorkers(n int) {
    for i := 0; i < n; i++ {
        go func() {
            for job := range s.jobQueue { // channel como fila
                job.Result <- s.processTransaction(ctx, job.Transaction)
            }
        }()
    }
}
```

### Goroutines concorrentes para leitura em paralelo
```go
// Summary busca income e expense em paralelo usando goroutines + channels
go func() { incomeCh <- s.txRepo.SumByType(ctx, accountID, Income) }()
go func() { expenseCh <- s.txRepo.SumByType(ctx, accountID, Expense) }()

ir := <-incomeCh
er := <-expenseCh
```

### Interfaces para desacoplamento
```go
// A service depende de interfaces, não de implementações concretas
type AccountRepository interface {
    Create(ctx context.Context, a *Account) error
    GetByID(ctx context.Context, id int64) (*Account, error)
    // ...
}
```

---

## 🗺️ Roadmap

- [ ] Autenticação JWT
- [ ] Relatórios mensais com gráficos
- [ ] Notificações por e-mail (AWS SES)
- [ ] Deploy na AWS Lambda
- [ ] FIX Protocol adapter para cotações de ativos
- [ ] Frontend simples (HTMX ou React)
