-- NeuronIP Demo Data Seeding Script
-- This SQL script seeds demo data directly into the database
-- Run with: psql -d neuronip -f scripts/seed-demo.sql

BEGIN;

-- Seed Demo Users
INSERT INTO neuronip.users (email, password_hash, name, role, metadata) VALUES
    ('demo@example.com', '$2a$10$dummy', 'Demo User', 'admin', '{"seeded": true}'::jsonb),
    ('john@example.com', '$2a$10$dummy', 'John Doe', 'user', '{"seeded": true}'::jsonb),
    ('jane@example.com', '$2a$10$dummy', 'Jane Smith', 'user', '{"seeded": true}'::jsonb),
    ('admin@example.com', '$2a$10$dummy', 'Admin User', 'admin', '{"seeded": true}'::jsonb)
ON CONFLICT (email) DO UPDATE SET
    name = EXCLUDED.name,
    role = EXCLUDED.role,
    metadata = EXCLUDED.metadata;

-- Seed API Keys
INSERT INTO neuronip.api_keys (name, key_hash, rate_limit_per_hour, metadata) VALUES
    ('Demo API Key', '$2a$10$demo', 1000, '{"seeded": true}'::jsonb),
    ('Development Key', '$2a$10$dev', 500, '{"seeded": true}'::jsonb)
ON CONFLICT DO NOTHING;

-- Seed Support Tickets
INSERT INTO neuronip.support_tickets (
    id, customer_id, subject, description, priority, status, metadata
) VALUES
    (
        'ticket-001',
        'customer-123',
        'Password reset not working',
        'I tried to reset my password but did not receive the email. This is the second time this has happened.',
        'high',
        'open',
        '{"seeded": true, "customer_email": "john.doe@example.com", "customer_name": "John Doe"}'::jsonb
    ),
    (
        'ticket-002',
        'customer-456',
        'API rate limit exceeded',
        'I''m getting 429 errors when making API calls. I thought my plan included 10,000 requests per day.',
        'medium',
        'open',
        '{"seeded": true, "customer_email": "jane.smith@example.com", "customer_name": "Jane Smith"}'::jsonb
    ),
    (
        'ticket-003',
        'customer-789',
        'Feature request: Bulk operations',
        'It would be great to have bulk operations for managing multiple resources at once.',
        'low',
        'open',
        '{"seeded": true, "customer_email": "support@example.com", "customer_name": "Support Team"}'::jsonb
    )
ON CONFLICT (id) DO NOTHING;

-- Seed Semantic Collections
INSERT INTO neuronip.semantic_collections (id, name, metadata) VALUES
    ('kb-general', 'General Documentation', '{"seeded": true}'::jsonb),
    ('kb-faq', 'Frequently Asked Questions', '{"seeded": true}'::jsonb)
ON CONFLICT (id) DO NOTHING;

-- Seed Knowledge Documents
INSERT INTO neuronip.knowledge_documents (
    id, collection_id, title, content, content_type, metadata
) VALUES
    (
        'doc-001',
        'kb-general',
        'Getting Started with NeuronIP',
        'NeuronIP is an AI-native intelligence platform built on PostgreSQL. To get started, you''ll need PostgreSQL 16+ with the NeuronDB extension installed. Once installed, you can start the API server and frontend. The platform provides five core capabilities: Semantic Knowledge Search, Data Warehouse Q&A, Customer Support Memory, Compliance & Audit Analytics, and Agent Workflows.',
        'documentation',
        '{"seeded": true, "tags": ["getting-started", "installation", "overview"]}'::jsonb
    ),
    (
        'doc-002',
        'kb-general',
        'Authentication and API Keys',
        'NeuronIP supports multiple authentication methods. API keys are the primary method for service-to-service communication. To create an API key, navigate to Settings > API Keys and click Create. API keys support scopes for fine-grained permissions. You can set rate limits, expiration dates, and rotation policies.',
        'documentation',
        '{"seeded": true, "tags": ["authentication", "api-keys", "security"]}'::jsonb
    ),
    (
        'doc-003',
        'kb-faq',
        'How do I reset my password?',
        'To reset your password, click the "Forgot Password" link on the login page. Enter your email address and check your inbox for reset instructions. If you don''t receive the email, check your spam folder or contact support.',
        'faq',
        '{"seeded": true, "tags": ["password", "authentication", "help"]}'::jsonb
    ),
    (
        'doc-004',
        'kb-faq',
        'What are API rate limits?',
        'API rate limits control how many requests you can make per hour. Default limits are 1,000 requests per hour. You can view your current usage in the dashboard and request higher limits if needed. Rate limit headers are included in all API responses.',
        'faq',
        '{"seeded": true, "tags": ["api", "rate-limits", "limits"]}'::jsonb
    )
ON CONFLICT (id) DO NOTHING;

-- Seed Warehouse Schemas
INSERT INTO neuronip.warehouse_schemas (
    id, name, description, schema_definition, metadata
) VALUES
    (
        'sales-schema',
        'Sales Data Warehouse',
        'E-commerce sales data with products, customers, orders, and revenue',
        '{
            "tables": [
                {
                    "name": "products",
                    "description": "Product catalog with pricing and categories",
                    "columns": [
                        {"name": "product_id", "type": "uuid", "description": "Unique product identifier"},
                        {"name": "name", "type": "text", "description": "Product name"},
                        {"name": "category", "type": "text", "description": "Product category"},
                        {"name": "price", "type": "decimal", "description": "Product price in USD"},
                        {"name": "stock_quantity", "type": "integer", "description": "Available stock"}
                    ]
                },
                {
                    "name": "customers",
                    "description": "Customer information and demographics",
                    "columns": [
                        {"name": "customer_id", "type": "uuid", "description": "Unique customer identifier"},
                        {"name": "email", "type": "text", "description": "Customer email address"},
                        {"name": "name", "type": "text", "description": "Customer full name"},
                        {"name": "country", "type": "text", "description": "Customer country"}
                    ]
                },
                {
                    "name": "orders",
                    "description": "Order transactions with line items",
                    "columns": [
                        {"name": "order_id", "type": "uuid", "description": "Unique order identifier"},
                        {"name": "customer_id", "type": "uuid", "description": "Customer who placed the order"},
                        {"name": "order_date", "type": "timestamp", "description": "Order date and time"},
                        {"name": "total_amount", "type": "decimal", "description": "Total order amount in USD"}
                    ]
                }
            ]
        }'::jsonb,
        '{"seeded": true}'::jsonb
    )
ON CONFLICT (id) DO NOTHING;

-- Seed Saved Searches
INSERT INTO neuronip.saved_searches (
    id, name, description, query, is_public, metadata
) VALUES
    (
        'search-001',
        'Top Products by Revenue',
        'Find products with highest revenue',
        'SELECT name, SUM(total_amount) as revenue FROM orders JOIN order_items ON orders.order_id = order_items.order_id GROUP BY name ORDER BY revenue DESC LIMIT 10',
        true,
        '{"seeded": true}'::jsonb
    ),
    (
        'search-002',
        'Active Customers',
        'List all active customers',
        'SELECT * FROM customers WHERE registration_date > NOW() - INTERVAL ''30 days''',
        false,
        '{"seeded": true}'::jsonb
    ),
    (
        'search-003',
        'Monthly Sales Summary',
        'Monthly sales breakdown',
        'SELECT DATE_TRUNC(''month'', order_date) as month, SUM(total_amount) as total FROM orders GROUP BY month ORDER BY month DESC',
        true,
        '{"seeded": true}'::jsonb
    )
ON CONFLICT (id) DO NOTHING;

-- Seed Workflows
INSERT INTO neuronip.workflows (
    id, name, description, definition, metadata
) VALUES
    (
        'workflow-001',
        'Data Processing Pipeline',
        'Processes incoming data and generates reports',
        '{
            "nodes": [
                {"id": "node-1", "type": "agent", "name": "Process Data"},
                {"id": "node-2", "type": "script", "name": "Generate Report"}
            ],
            "edges": [
                {"from": "node-1", "to": "node-2"}
            ]
        }'::jsonb,
        '{"seeded": true}'::jsonb
    )
ON CONFLICT (id) DO NOTHING;

-- Seed Metrics
INSERT INTO neuronip.metrics (
    id, name, description, definition, metadata
) VALUES
    (
        'metric-001',
        'Total Revenue',
        'Sum of all order amounts',
        '{"type": "sum", "field": "total_amount", "table": "orders"}'::jsonb,
        '{"seeded": true}'::jsonb
    ),
    (
        'metric-002',
        'Customer Count',
        'Number of unique customers',
        '{"type": "count_distinct", "field": "customer_id", "table": "customers"}'::jsonb,
        '{"seeded": true}'::jsonb
    ),
    (
        'metric-003',
        'Average Order Value',
        'Average amount per order',
        '{"type": "avg", "field": "total_amount", "table": "orders"}'::jsonb,
        '{"seeded": true}'::jsonb
    )
ON CONFLICT (id) DO NOTHING;

COMMIT;

-- Summary
SELECT 'Demo data seeding completed!' as status;
SELECT COUNT(*) as users_count FROM neuronip.users WHERE metadata->>'seeded' = 'true';
SELECT COUNT(*) as tickets_count FROM neuronip.support_tickets WHERE metadata->>'seeded' = 'true';
SELECT COUNT(*) as documents_count FROM neuronip.knowledge_documents WHERE metadata->>'seeded' = 'true';
SELECT COUNT(*) as searches_count FROM neuronip.saved_searches WHERE metadata->>'seeded' = 'true';
