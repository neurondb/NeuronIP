# 💻 Development Setup

<div align="center">

**Set up your development environment**

[Contributing →](contributing.md) • [Coding Standards →](coding-standards.md)

</div>

---

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Backend Setup](#backend-setup)
- [Frontend Setup](#frontend-setup)
- [Database Setup](#database-setup)
- [Running Tests](#running-tests)
- [Development Workflow](#development-workflow)

---

## ✅ Prerequisites

- Go 1.24+
- Node.js 18+
- PostgreSQL 16+
- Docker (optional)

---

## 🔧 Backend Setup

```bash
cd api
go mod download
go run cmd/server/main.go
```

---

## 🎨 Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

---

## 💾 Database Setup

```bash
psql -d neuronip -f ../neuronip.sql
```

---

## 🧪 Running Tests

```bash
# Backend
cd api
go test ./...

# Frontend
cd frontend
npm test
```

---

## 📚 Related Documentation

- [Contributing Guide](contributing.md) - How to contribute
- [Coding Standards](coding-standards.md) - Code style guide

---

<div align="center">

[← Back to Documentation](../README.md)

</div>
