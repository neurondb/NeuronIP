# 🔄 Data Flow Architecture

<div align="center">

**How Data Moves Through NeuronIP**

[← Database](database.md) • [Back to Architecture](README.md)

</div>

---

## 📋 Table of Contents

- [Overview](#overview)
- [Request Flow](#request-flow)
- [Semantic Search Flow](#semantic-search-flow)
- [Warehouse Q&A Flow](#warehouse-qa-flow)
- [Support Ticket Flow](#support-ticket-flow)
- [Workflow Execution Flow](#workflow-execution-flow)
- [Data Processing Pipelines](#data-processing-pipelines)

---

## 🎯 Overview

Understanding data flow is crucial for debugging, optimization, and system design. This document describes how data moves through NeuronIP's various components.

---

## 📥 Request Flow

### Standard API Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Router
    participant Middleware
    participant Handler
    participant Service
    participant Database
    participant NeuronDB
    participant NeuronAgent
    
    Client->>Router: HTTP Request
    Router->>Middleware: Apply Middleware
    Middleware->>Middleware: Authentication
    Middleware->>Middleware: Rate Limiting
    Middleware->>Handler: Validated Request
    Handler->>Handler: Parse & Validate
    Handler->>Service: Business Logic Call
    Service->>Database: Query Data
    Database->>NeuronDB: Vector/ML Operations
    NeuronDB-->>Database: Results
    Database-->>Service: Data
    Service->>NeuronAgent: AI Operations (if needed)
    NeuronAgent-->>Service: AI Results
    Service-->>Handler: Response Data
    Handler-->>Middleware: HTTP Response
    Middleware-->>Router: Response
    Router-->>Client: JSON Response
```

### Request Processing Steps

1. **Client Request** - HTTP request arrives at router
2. **Middleware Stack** - Authentication, logging, CORS, rate limiting
3. **Handler** - Request parsing and validation
4. **Service** - Business logic execution
5. **Database** - Data retrieval/storage
6. **NeuronDB** - Vector/ML operations (if needed)
7. **NeuronAgent** - AI agent operations (if needed)
8. **Response** - JSON response back to client

---

## 🔍 Semantic Search Flow

### Search Request Flow

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant API
    participant SemanticService
    participant NeuronDB
    participant Database
    
    User->>Frontend: Enter search query
    Frontend->>API: POST /api/v1/semantic/search
    API->>SemanticService: Search(query, limit)
    SemanticService->>NeuronDB: GenerateEmbedding(query)
    NeuronDB-->>SemanticService: Query embedding
    SemanticService->>NeuronDB: VectorSearch(embedding, limit)
    NeuronDB->>Database: Query document_chunks
    Database-->>NeuronDB: Chunk data
    NeuronDB-->>SemanticService: Similar chunks
    SemanticService->>Database: Get full documents
    Database-->>SemanticService: Documents
    SemanticService-->>API: Search results
    API-->>Frontend: JSON response
    Frontend-->>User: Display results
```

### Document Creation Flow

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant API
    participant SemanticService
    participant Database
    participant NeuronDB
    
    User->>Frontend: Upload document
    Frontend->>API: POST /api/v1/semantic/documents
    API->>SemanticService: CreateDocument(doc)
    SemanticService->>Database: Insert document
    Database-->>SemanticService: Document ID
    SemanticService->>SemanticService: Chunk document
    loop For each chunk
        SemanticService->>NeuronDB: GenerateEmbedding(chunk)
        NeuronDB-->>SemanticService: Chunk embedding
        SemanticService->>Database: Insert chunk + embedding
    end
    Database-->>SemanticService: Success
    SemanticService-->>API: Document created
    API-->>Frontend: Success response
    Frontend-->>User: Document indexed
```

---

## 💬 Warehouse Q&A Flow

### Query Processing Flow

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant API
    participant WarehouseService
    participant NeuronAgent
    participant Database
    participant Warehouse
    
    User->>Frontend: Ask question
    Frontend->>API: POST /api/v1/warehouse/query
    API->>WarehouseService: Query(nlQuery, schemaId)
    WarehouseService->>Database: Get schema definition
    Database-->>WarehouseService: Schema
    WarehouseService->>NeuronAgent: GenerateSQL(nlQuery, schema)
    NeuronAgent-->>WarehouseService: SQL query
    WarehouseService->>Warehouse: Execute SQL
    Warehouse-->>WarehouseService: Query results
    WarehouseService->>NeuronAgent: ExplainQuery(sql, results)
    NeuronAgent-->>WarehouseService: Explanation
    WarehouseService->>Database: Save query history
    Database-->>WarehouseService: Query saved
    WarehouseService-->>API: Query result + explanation
    API-->>Frontend: JSON response
    Frontend-->>User: Display results + chart
```

### Schema Discovery Flow

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant API
    participant WarehouseService
    participant Warehouse
    participant NeuronAgent
    
    User->>Frontend: Request schema discovery
    Frontend->>API: POST /api/v1/warehouse/schemas
    API->>WarehouseService: DiscoverSchema(connection)
    WarehouseService->>Warehouse: Get table metadata
    Warehouse-->>WarehouseService: Tables, columns, types
    WarehouseService->>NeuronAgent: AnalyzeSchema(metadata)
    NeuronAgent-->>WarehouseService: Schema description
    WarehouseService->>Database: Save schema
    Database-->>WarehouseService: Schema ID
    WarehouseService-->>API: Schema created
    API-->>Frontend: Schema response
    Frontend-->>User: Display schema
```

---

## 🎫 Support Ticket Flow

### Ticket Creation Flow

```mermaid
sequenceDiagram
    participant Customer
    participant Frontend
    participant API
    participant SupportService
    participant NeuronAgent
    participant Database
    
    Customer->>Frontend: Create ticket
    Frontend->>API: POST /api/v1/support/tickets
    API->>SupportService: CreateTicket(ticket)
    SupportService->>Database: Insert ticket
    Database-->>SupportService: Ticket ID
    SupportService->>NeuronAgent: FindSimilarCases(ticket)
    NeuronAgent->>Database: Search past tickets
    Database-->>NeuronAgent: Similar tickets
    NeuronAgent-->>SupportService: Similar cases
    SupportService->>Database: Link similar cases
    Database-->>SupportService: Success
    SupportService-->>API: Ticket created
    API-->>Frontend: Ticket response
    Frontend-->>Customer: Ticket number
```

### Conversation Flow

```mermaid
sequenceDiagram
    participant Customer
    participant Frontend
    participant API
    participant SupportService
    participant NeuronAgent
    participant Database
    
    Customer->>Frontend: Send message
    Frontend->>API: POST /api/v1/support/tickets/{id}/conversations
    API->>SupportService: AddConversation(ticketId, message)
    SupportService->>Database: Insert conversation
    Database-->>SupportService: Conversation ID
    SupportService->>NeuronAgent: ProcessMessage(ticket, message)
    NeuronAgent->>Database: Get ticket history
    Database-->>NeuronAgent: History
    NeuronAgent->>NeuronAgent: Generate response
    NeuronAgent-->>SupportService: AI response
    SupportService->>Database: Save AI response
    Database-->>SupportService: Success
    SupportService-->>API: Conversation added
    API-->>Frontend: Response
    Frontend-->>Customer: Display AI response
```

---

## ⚙️ Workflow Execution Flow

### Workflow Execution

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant API
    participant WorkflowService
    participant NeuronAgent
    participant NeuronMCP
    participant Database
    
    User->>Frontend: Execute workflow
    Frontend->>API: POST /api/v1/workflows/{id}/execute
    API->>WorkflowService: ExecuteWorkflow(workflowId, input)
    WorkflowService->>Database: Get workflow definition
    Database-->>WorkflowService: Workflow
    WorkflowService->>NeuronAgent: StartWorkflow(definition, input)
    NeuronAgent-->>WorkflowService: Execution ID
    loop For each step
        NeuronAgent->>NeuronMCP: Execute tool
        NeuronMCP-->>NeuronAgent: Tool result
        NeuronAgent->>Database: Save step result
    end
    NeuronAgent-->>WorkflowService: Workflow complete
    WorkflowService->>Database: Update execution status
    Database-->>WorkflowService: Success
    WorkflowService-->>API: Execution result
    API-->>Frontend: Response
    Frontend-->>User: Display results
```

---

## 🔄 Data Processing Pipelines

### Document Indexing Pipeline

```mermaid
graph LR
    A[Document Upload] --> B[Validate Document]
    B --> C[Extract Text]
    C --> D[Chunk Document]
    D --> E[Generate Embeddings]
    E --> F[Store in Database]
    F --> G[Index Complete]
    
    style A fill:#e1f5ff
    style G fill:#c8e6c9
```

### Query Processing Pipeline

```mermaid
graph LR
    A[Natural Language Query] --> B[Parse Query]
    B --> C[Generate Embedding]
    C --> D[Vector Search]
    D --> E[Rank Results]
    E --> F[Format Response]
    F --> G[Return Results]
    
    style A fill:#e1f5ff
    style G fill:#c8e6c9
```

### Compliance Check Pipeline

```mermaid
graph LR
    A[Data Input] --> B[Load Policies]
    B --> C[Check Compliance]
    C --> D{Compliant?}
    D -->|Yes| E[Log Success]
    D -->|No| F[Generate Alert]
    F --> G[Notify Stakeholders]
    E --> H[Complete]
    G --> H
    
    style A fill:#e1f5ff
    style H fill:#c8e6c9
    style F fill:#ffcdd2
```

---

## 📊 Data Transformation

### Embedding Generation

```mermaid
graph TD
    A[Text Input] --> B[Text Preprocessing]
    B --> C[Tokenization]
    C --> D[NeuronDB Embedding Model]
    D --> E[Vector Output]
    E --> F[Store in Database]
    
    style A fill:#e1f5ff
    style E fill:#fff9c4
    style F fill:#c8e6c9
```

### SQL Generation

```mermaid
graph TD
    A[Natural Language] --> B[Schema Context]
    B --> C[NeuronAgent LLM]
    C --> D[SQL Query]
    D --> E[Validate SQL]
    E --> F{Valid?}
    F -->|Yes| G[Execute Query]
    F -->|No| H[Error Response]
    G --> I[Return Results]
    
    style A fill:#e1f5ff
    style D fill:#fff9c4
    style I fill:#c8e6c9
    style H fill:#ffcdd2
```

---

## 🔐 Security Flow

### Authentication Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant AuthMiddleware
    participant Database
    
    Client->>API: Request with API Key
    API->>AuthMiddleware: Validate Request
    AuthMiddleware->>Database: Lookup API Key
    Database-->>AuthMiddleware: Key Data
    AuthMiddleware->>AuthMiddleware: Verify Key
    AuthMiddleware->>AuthMiddleware: Check Expiration
    AuthMiddleware->>AuthMiddleware: Check Rate Limit
    AuthMiddleware-->>API: Authenticated User
    API->>API: Process Request
    API-->>Client: Response
```

---

## 📈 Performance Optimization

### Caching Flow

```mermaid
graph LR
    A[Request] --> B{Cache Hit?}
    B -->|Yes| C[Return Cached]
    B -->|No| D[Process Request]
    D --> E[Store in Cache]
    E --> F[Return Result]
    C --> G[Response]
    F --> G
    
    style A fill:#e1f5ff
    style C fill:#fff9c4
    style G fill:#c8e6c9
```

### Connection Pooling

```mermaid
graph TD
    A[Service Request] --> B[Get Connection]
    B --> C{Pool Available?}
    C -->|Yes| D[Use Existing]
    C -->|No| E[Create New]
    D --> F[Execute Query]
    E --> F
    F --> G[Return Connection]
    G --> H[Back to Pool]
    
    style A fill:#e1f5ff
    style D fill:#fff9c4
    style H fill:#c8e6c9
```

---

## 📚 Related Documentation

- [Backend Architecture](backend.md) - Backend services
- [Database Design](database.md) - Database schema
- [API Reference](../api/endpoints.md) - API endpoints

---

<div align="center">

[← Back to Architecture](README.md)

</div>
