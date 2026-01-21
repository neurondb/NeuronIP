# Changelog

All notable changes to NeuronIP will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of NeuronIP
- Semantic Knowledge Search capability
- Data Warehouse Q&A functionality
- Customer Support Memory with AI agents
- Compliance & Audit Analytics
- Agent Workflows with NeuronDB integration
- REST API backend built with Go
- Next.js frontend with TypeScript
- Docker Compose setup for easy deployment
- PostgreSQL with NeuronDB extension support
- Integration with NeuronAgent and NeuronMCP

### Enterprise-Grade Features
- **Product Positioning**: Support Memory Hub as killer wedge, demo datasets and scripts, "Why NeuronIP" page with 6 hard claims
- **Authentication**: OIDC SSO, SCIM 2.0 provisioning, enhanced MFA with TOTP, session management with concurrent limits
- **RBAC**: Organization, workspace, and resource-level permissions, custom roles, permission inheritance
- **Multi-Tenancy**: Per-tenant schema/database isolation with strict isolation tests
- **API Keys**: Scopes, rotation, expiry, per-key rate limits, usage analytics
- **Data Plane**: Connectors UI with schedule editor and credentials vault, ingestion jobs with retries, backpressure, dead-letter queue
- **Pipelines**: Versioned, replayable chunking and embedding pipelines
- **Hybrid Search**: Saved searches with SQL filters + vector queries
- **Semantic Layer**: Metrics, dimensions, definitions, owners, lineage
- **Query Governance**: Function allow-list, read-only mode, sandbox role
- **Result Caching**: TTL-based caching with invalidation rules
- **Agent Hub**: Templates library, tools registry, memory policies
- **Agent Evaluation**: Golden sets, regression tests, scoring, drift tracking
- **Human-in-the-Loop**: Approve/edit/send workflow with learning from edits
- **Agent Audit Trail**: Complete logging of every agent action and tool call
- **Observability**: Request ID propagation across UI→API→DB, p50/p95/p99 latency metrics, error rates, token usage, embedding cost tracking, distributed tracing
- **Security**: Secrets scanning, dependency scanning, signed builds, SBOM generation
- **CI/CD**: Integration tests with docker-compose, migration runner with rollback, deterministic seed data loader, automated release workflow

### Changed
- N/A (initial release)

### Deprecated
- N/A (initial release)

### Removed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- N/A (initial release)

## [1.0.0] - YYYY-MM-DD

### Added
- Initial release of NeuronIP AI-Native Enterprise Intelligence Platform
- Backend API with Go (v1.24+)
- Frontend UI with Next.js 14 and TypeScript
- Docker support for containerized deployment
- Comprehensive documentation
- CI/CD pipeline with GitHub Actions

---

## Release Types

- **Major** (X.0.0): Breaking changes
- **Minor** (0.X.0): New features (backward compatible)
- **Patch** (0.0.X): Bug fixes (backward compatible)

## Contributing

When adding entries to the changelog, please follow these guidelines:

1. Add entries under the `[Unreleased]` section
2. Group changes by type (Added, Changed, Deprecated, Removed, Fixed, Security)
3. Link to relevant issues/PRs when possible
4. Use present tense ("Add feature" not "Added feature")
5. Don't include internal implementation details unless they affect users
