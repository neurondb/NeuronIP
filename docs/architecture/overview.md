# 🏗️ System Architecture Overview

<div align="center">

**Understanding NeuronIP's Architecture**

[Backend →](backend.md) • [Frontend →](frontend.md) • [Database →](database.md) • [Data Flow →](data-flow.md)

</div>

---

## 📋 Table of Contents

- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Component Diagram](#component-diagram)
- [Technology Stack](#technology-stack)
- [Design Principles](#design-principles)
- [Deployment Architecture](#deployment-architecture)
- [Integration Points](#integration-points)

---

## 🎯 Overview

NeuronIP is built as a modern, microservices-inspired architecture with clear separation of concerns. The system consists of:

1. **Backend API** - Go-based REST API server
2. **Frontend** - Next.js web application
3. **Database** - PostgreSQL with NeuronDB extension
4. **External Services** - NeuronDB, NeuronAgent, NeuronMCP

### High-Level Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        Web[Web Browser]
        API_Client[API Clients]
        Mobile[Mobile Apps]
    end
    
    subgraph "Frontend Layer"
        NextJS[Next.js Frontend<br/>Port 3001]
    end
    
    subgraph "API Layer"
        API[Go API Server<br/>Port 8082]
        Auth[Authentication<br/>Middleware]
        RateLimit[Rate Limiting<br/>Middleware]
    end
    
    subgraph "Service Layer"
        Semantic[Semantic<br/>Service]
        Warehouse[Warehouse<br/>Service]
        Support[Support<br/>Service]
        Compliance[Compliance<br/>Service]
        Workflows[Workflow<br/>Service]
    end
    
    subgraph "Data Layer"
        PG[(PostgreSQL<br/>Database)]
        NeuronDB[(NeuronDB<br/>Extension)]
    end
    
    subgraph "External Services"
        NeuronAgent[NeuronAgent<br/>Service]
        NeuronMCP[NeuronMCP<br/>Tools]
    end
    
    Web --> NextJS
    API_Client --> API
    Mobile --> API
    
    NextJS --> API
    
    API --> Auth
    Auth --> RateLimit
    RateLimit --> Semantic
    RateLimit --> Warehouse
    RateLimit --> Support
    RateLimit --> Compliance
    RateLimit --> Workflows
    
    Semantic --> PG
    Warehouse --> PG
    Support --> PG
    Compliance --> PG
    Workflows --> PG
    
    PG --> NeuronDB
    
    Semantic --> NeuronAgent
    Warehouse --> NeuronAgent
    Support --> NeuronAgent
    Workflows --> NeuronAgent
    
    Workflows --> NeuronMCP
```

---

## 🧩 Component Diagram

### Detailed Component Architecture

```mermaid
graph LR
    subgraph "API Server"
        Router[Gorilla Mux<br/>Router]
        Middleware[Middleware<br/>Stack]
        Handlers[HTTP<br/>Handlers]
    end
    
    subgraph "Core Services"
        SemanticSvc[Semantic<br/>Service]
        WarehouseSvc[Warehouse<br/>Service]
        SupportSvc[Support<br/>Service]
        ComplianceSvc[Compliance<br/>Service]
        WorkflowSvc[Workflow<br/>Service]
    end
    
    subgraph "Infrastructure"
        DB[Database<br/>Pool]
        Logger[Logger]
        Metrics[Prometheus<br/>Metrics]
        Cache[Cache<br/>Layer]
    end
    
    subgraph "External"
        Agent[NeuronAgent<br/>Client]
        MCP[NeuronMCP<br/>Client]
    end
    
    Router --> Middleware
    Middleware --> Handlers
    Handlers --> SemanticSvc
    Handlers --> WarehouseSvc
    Handlers --> SupportSvc
    Handlers --> ComplianceSvc
    Handlers --> WorkflowSvc
    
    SemanticSvc --> DB
    WarehouseSvc --> DB
    SupportSvc --> DB
    ComplianceSvc --> DB
    WorkflowSvc --> DB
    
    SemanticSvc --> Agent
    WarehouseSvc --> Agent
    SupportSvc --> Agent
    WorkflowSvc --> Agent
    WorkflowSvc --> MCP
    
    DB --> Logger
    Handlers --> Metrics
    SemanticSvc --> Cache
```

---

## 🛠️ Technology Stack

### Backend

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.24+ | Programming language |
| **Gorilla Mux** | Latest | HTTP router |
| **pgx/v5** | Latest | PostgreSQL driver |
| **Prometheus** | Latest | Metrics collection |

### Frontend

| Technology | Version | Purpose |
|------------|---------|---------|
| **Next.js** | 14+ | React framework |
| **TypeScript** | Latest | Type safety |
| **Tailwind CSS** | Latest | Styling |
| **React Query** | Latest | Data fetching |

### Database

| Technology | Version | Purpose |
|------------|---------|---------|
| **PostgreSQL** | 16+ | Primary database |
| **NeuronDB** | Latest | AI-native extensions |
| **pgcrypto** | Latest | Encryption |
| **uuid-ossp** | Latest | UUID generation |

### Infrastructure

| Technology | Purpose |
|------------|---------|
| **Docker** | Containerization |
| **Docker Compose** | Local development |
| **Prometheus** | Metrics |
| **Grafana** | Visualization (optional) |

---

## 🎨 Design Principles

### 1. Separation of Concerns

Each layer has a distinct responsibility:

- **Handlers** - HTTP request/response handling
- **Services** - Business logic
- **Database** - Data persistence
- **External Clients** - Third-party integrations

### 2. Dependency Injection

Services receive dependencies through constructors:

```go
// Example: Service with injected dependencies
service := semantic.NewService(
    queries,      // Database queries
    pool,         // Connection pool
    neurondbClient, // NeuronDB client
)
```

### 3. Interface-Based Design

Services implement interfaces for testability:

```go
type SemanticService interface {
    Search(ctx context.Context, req SearchRequest) ([]SearchResult, error)
    CreateDocument(ctx context.Context, doc *Document) error
}
```

### 4. Error Handling

Consistent error handling across the application:

```go
// Structured errors
if err != nil {
    return nil, errors.NotFound("Document not found")
}
```

### 5. Configuration Management

Environment-based configuration with validation:

```go
cfg := config.Load()
if err := cfg.Validate(); err != nil {
    // Handle validation error
}
```

---

## 🚀 Deployment Architecture

### Development Environment

```mermaid
graph TB
    subgraph "Local Machine"
        Dev[Developer]
        Docker[Docker Compose]
    end
    
    subgraph "Containers"
        Frontend[Frontend<br/>Container]
        API[API<br/>Container]
        DB[PostgreSQL<br/>Container]
    end
    
    Dev --> Docker
    Docker --> Frontend
    Docker --> API
    Docker --> DB
```

### Production Environment

```mermaid
graph TB
    subgraph "Load Balancer"
        LB[NGINX/HAProxy]
    end
    
    subgraph "Application Tier"
        API1[API Instance 1]
        API2[API Instance 2]
        API3[API Instance N]
    end
    
    subgraph "Frontend Tier"
        Frontend1[Frontend Instance 1]
        Frontend2[Frontend Instance 2]
    end
    
    subgraph "Database Tier"
        Primary[(PostgreSQL<br/>Primary)]
        Replica[(PostgreSQL<br/>Replica)]
    end
    
    subgraph "External Services"
        NeuronAgent[NeuronAgent<br/>Cluster]
        NeuronMCP[NeuronMCP<br/>Service]
    end
    
    LB --> API1
    LB --> API2
    LB --> API3
    
    LB --> Frontend1
    LB --> Frontend2
    
    API1 --> Primary
    API2 --> Primary
    API3 --> Primary
    
    API1 --> Replica
    API2 --> Replica
    API3 --> Replica
    
    API1 --> NeuronAgent
    API2 --> NeuronAgent
    API3 --> NeuronAgent
    
    API1 --> NeuronMCP
    API2 --> NeuronMCP
    API3 --> NeuronMCP
```

---

## 🔌 Integration Points

### NeuronDB Integration

NeuronDB provides AI-native capabilities:

- **Vector Operations** - Semantic search and embeddings
- **ML Functions** - Machine learning inference
- **RAG Tools** - Retrieval-augmented generation

### NeuronAgent Integration

NeuronAgent provides AI agent capabilities:

- **Session Management** - Long-term memory
- **Workflow Execution** - Multi-step agent workflows
- **Evaluation** - Agent performance metrics

### NeuronMCP Integration

NeuronMCP provides Model Context Protocol tools:

- **Vector Tools** - Vector operations
- **Embedding Tools** - Text embeddings
- **RAG Tools** - RAG pipeline operations
- **PostgreSQL Tools** - Database operations

---

## 📊 Request Flow

### Typical API Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Router
    participant Middleware
    participant Handler
    participant Service
    participant Database
    participant NeuronDB
    
    Client->>Router: HTTP Request
    Router->>Middleware: Request
    Middleware->>Middleware: Auth Check
    Middleware->>Middleware: Rate Limit
    Middleware->>Handler: Validated Request
    Handler->>Service: Business Logic
    Service->>Database: Query Data
    Database->>NeuronDB: Vector/ML Operations
    NeuronDB-->>Database: Results
    Database-->>Service: Data
    Service-->>Handler: Response
    Handler-->>Middleware: HTTP Response
    Middleware-->>Router: Response
    Router-->>Client: JSON Response
```

---

## 🔐 Security Architecture

### Authentication Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Auth
    participant Database
    
    Client->>API: Request with API Key
    API->>Auth: Validate API Key
    Auth->>Database: Check Key
    Database-->>Auth: Key Valid
    Auth->>Auth: Generate JWT
    Auth-->>API: Authenticated User
    API->>API: Process Request
    API-->>Client: Response
```

### Authorization Model

- **API Keys** - Service-to-service authentication
- **JWT Tokens** - User session tokens
- **RBAC** - Role-based access control
- **RLS** - Row-level security in database

---

## 📈 Scalability Considerations

### Horizontal Scaling

- **Stateless API** - API instances can scale horizontally
- **Database Pooling** - Connection pooling for efficiency
- **Caching** - Redis/Memcached for frequently accessed data
- **Load Balancing** - Distribute traffic across instances

### Vertical Scaling

- **Database Optimization** - Indexes, query optimization
- **Connection Pooling** - Efficient connection management
- **Resource Limits** - CPU and memory limits per service

---

## 🔍 Monitoring and Observability

### Metrics

- **Prometheus** - Metrics collection
- **Business Metrics** - Custom application metrics
- **System Metrics** - CPU, memory, disk usage

### Logging

- **Structured Logging** - JSON-formatted logs
- **Log Levels** - Debug, Info, Warn, Error
- **Request Tracing** - Request ID tracking

### Tracing

- **Request IDs** - Track requests across services
- **Performance Metrics** - Response time tracking
- **Error Tracking** - Error aggregation and alerting

---

## 📚 Related Documentation

- [Backend Architecture](backend.md) - Detailed backend design
- [Frontend Architecture](frontend.md) - Frontend structure
- [Database Design](database.md) - Schema and data modeling
- [Data Flow](data-flow.md) - Request/response flows
- [Deployment Guide](../deployment/production.md) - Production deployment

---

<div align="center">

[← Back to Architecture](README.md) • [Next: Backend Architecture →](backend.md)

</div>
