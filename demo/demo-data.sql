/*-------------------------------------------------------------------------
 *
 * demo-data.sql
 *    Comprehensive Demo Data for NeuronIP Platform
 *
 * This script populates the NeuronIP database with realistic demo data
 * that makes all pages show meaningful information. The data simulates
 * a production-like environment with users, searches, queries, workflows,
 * compliance checks, and more.
 *
 * Run this after all migrations have been executed.
 *
 * Usage:
 *   psql -h localhost -U postgres -d neuronip -f demo/demo-data.sql
 *
 *-------------------------------------------------------------------------
 */

-- Set timezone for consistent timestamps
SET timezone = 'UTC';

-- ============================================================================
-- SECTION 1: USERS AND PROFILES
-- ============================================================================

-- Insert demo users
INSERT INTO neuronip.users (id, email, email_verified, name, role, last_login_at, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'admin@acmecorp.com', true, 'Sarah Johnson', 'admin', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '90 days', NOW() - INTERVAL '2 hours'),
('550e8400-e29b-41d4-a716-446655440001', 'alex.chen@acmecorp.com', true, 'Alex Chen', 'analyst', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '60 days', NOW() - INTERVAL '30 minutes'),
('550e8400-e29b-41d4-a716-446655440002', 'maria.rodriguez@acmecorp.com', true, 'Maria Rodriguez', 'analyst', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '45 days', NOW() - INTERVAL '1 hour'),
('550e8400-e29b-41d4-a716-446655440003', 'david.kim@acmecorp.com', true, 'David Kim', 'developer', NOW() - INTERVAL '15 minutes', NOW() - INTERVAL '30 days', NOW() - INTERVAL '15 minutes'),
('550e8400-e29b-41d4-a716-446655440004', 'emily.taylor@acmecorp.com', true, 'Emily Taylor', 'support', NOW() - INTERVAL '45 minutes', NOW() - INTERVAL '20 days', NOW() - INTERVAL '45 minutes'),
('550e8400-e29b-41d4-a716-446655440005', 'james.wilson@acmecorp.com', true, 'James Wilson', 'analyst', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 hours'),
('550e8400-e29b-41d4-a716-446655440006', 'lisa.anderson@acmecorp.com', true, 'Lisa Anderson', 'analyst', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '10 days', NOW() - INTERVAL '5 hours'),
('550e8400-e29b-41d4-a716-446655440007', 'michael.brown@acmecorp.com', true, 'Michael Brown', 'developer', NOW() - INTERVAL '1 day', NOW() - INTERVAL '8 days', NOW() - INTERVAL '1 day')
ON CONFLICT (email) DO NOTHING;

-- Insert user profiles
INSERT INTO neuronip.user_profiles (user_id, bio, company, job_title, location, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'Platform administrator with 10+ years in data engineering', 'Acme Corporation', 'Senior Platform Administrator', 'San Francisco, CA', NOW() - INTERVAL '90 days', NOW()),
('550e8400-e29b-41d4-a716-446655440001', 'Data analyst specializing in business intelligence and analytics', 'Acme Corporation', 'Senior Data Analyst', 'New York, NY', NOW() - INTERVAL '60 days', NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'Business analyst focused on customer insights and metrics', 'Acme Corporation', 'Business Intelligence Analyst', 'Austin, TX', NOW() - INTERVAL '45 days', NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'Full-stack developer building data platforms', 'Acme Corporation', 'Senior Software Engineer', 'Seattle, WA', NOW() - INTERVAL '30 days', NOW()),
('550e8400-e29b-41d4-a716-446655440004', 'Customer support specialist with expertise in AI-powered solutions', 'Acme Corporation', 'Customer Support Lead', 'Boston, MA', NOW() - INTERVAL '20 days', NOW()),
('550e8400-e29b-41d4-a716-446655440005', 'Data scientist working on predictive analytics', 'Acme Corporation', 'Data Scientist', 'Chicago, IL', NOW() - INTERVAL '15 days', NOW()),
('550e8400-e29b-41d4-a716-446655440006', 'Analytics engineer building data pipelines', 'Acme Corporation', 'Analytics Engineer', 'Denver, CO', NOW() - INTERVAL '10 days', NOW()),
('550e8400-e29b-41d4-a716-446655440007', 'Platform engineer maintaining infrastructure', 'Acme Corporation', 'Platform Engineer', 'Portland, OR', NOW() - INTERVAL '8 days', NOW())
ON CONFLICT (user_id) DO NOTHING;

-- ============================================================================
-- SECTION 2: DATA SOURCES AND WAREHOUSE SCHEMAS
-- ============================================================================

-- Insert data sources (must be inserted before ingestion_jobs due to foreign key)
INSERT INTO neuronip.data_sources (id, name, source_type, connection_string, enabled, last_accessed_at, created_at, updated_at) VALUES
('660e8400-e29b-41d4-a716-446655440000', 'Production PostgreSQL', 'postgresql', 'postgresql://prod-db.acmecorp.com:5432/analytics', true, NOW() - INTERVAL '10 minutes', NOW() - INTERVAL '60 days', NOW() - INTERVAL '10 minutes'),
('660e8400-e29b-41d4-a716-446655440001', 'Customer Data Warehouse', 'postgresql', 'postgresql://warehouse.acmecorp.com:5432/customer_data', true, NOW() - INTERVAL '5 minutes', NOW() - INTERVAL '45 days', NOW() - INTERVAL '5 minutes'),
('660e8400-e29b-41d4-a716-446655440002', 'Salesforce API', 'api', 'https://api.salesforce.com/v1', true, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '30 days', NOW() - INTERVAL '1 hour'),
('660e8400-e29b-41d4-a716-446655440003', 'Marketing Analytics DB', 'mysql', 'mysql://marketing-db.acmecorp.com:3306/marketing', true, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '20 days', NOW() - INTERVAL '2 hours'),
('660e8400-e29b-41d4-a716-446655440004', 'Product Database', 'postgresql', 'postgresql://product-db.acmecorp.com:5432/products', true, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '15 days', NOW() - INTERVAL '30 minutes')
ON CONFLICT (name) DO NOTHING;

-- Insert warehouse schemas
INSERT INTO neuronip.warehouse_schemas (id, schema_name, database_name, description, tables, last_synced_at, created_at, updated_at) VALUES
('770e8400-e29b-41d4-a716-446655440000', 'analytics', 'analytics', 'Main analytics schema with business metrics', 
 '[
   {"name": "sales_transactions", "columns": ["id", "customer_id", "product_id", "amount", "transaction_date", "region"], "row_count": 1250000},
   {"name": "customer_profiles", "columns": ["id", "name", "email", "segment", "lifetime_value", "created_at"], "row_count": 45000},
   {"name": "product_catalog", "columns": ["id", "name", "category", "price", "in_stock", "supplier_id"], "row_count": 8500},
   {"name": "marketing_campaigns", "columns": ["id", "name", "channel", "budget", "start_date", "end_date", "conversions"], "row_count": 320},
   {"name": "user_activity", "columns": ["id", "user_id", "event_type", "timestamp", "properties"], "row_count": 2500000}
 ]'::jsonb,
 NOW() - INTERVAL '1 hour', NOW() - INTERVAL '60 days', NOW() - INTERVAL '1 hour'),

('770e8400-e29b-41d4-a716-446655440001', 'customer_data', 'customer_data', 'Customer data warehouse schema',
 '[
   {"name": "customers", "columns": ["id", "email", "name", "signup_date", "status", "tier"], "row_count": 125000},
   {"name": "orders", "columns": ["id", "customer_id", "order_date", "total", "status", "shipping_address"], "row_count": 890000},
   {"name": "subscriptions", "columns": ["id", "customer_id", "plan_type", "start_date", "renewal_date", "status"], "row_count": 45000},
   {"name": "support_tickets", "columns": ["id", "customer_id", "subject", "status", "created_at", "resolved_at"], "row_count": 12500}
 ]'::jsonb,
 NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '45 days', NOW() - INTERVAL '30 minutes'),

('770e8400-e29b-41d4-a716-446655440002', 'marketing', 'marketing', 'Marketing analytics schema',
 '[
   {"name": "campaigns", "columns": ["id", "name", "channel", "budget", "start_date", "end_date"], "row_count": 450},
   {"name": "ad_performance", "columns": ["id", "campaign_id", "impressions", "clicks", "conversions", "spend", "date"], "row_count": 125000},
   {"name": "email_events", "columns": ["id", "campaign_id", "recipient_id", "event_type", "timestamp"], "row_count": 2500000},
   {"name": "social_media_metrics", "columns": ["id", "platform", "post_id", "likes", "shares", "comments", "date"], "row_count": 85000}
 ]'::jsonb,
 NOW() - INTERVAL '2 hours', NOW() - INTERVAL '30 days', NOW() - INTERVAL '2 hours')
ON CONFLICT (schema_name, database_name) DO NOTHING;

-- ============================================================================
-- SECTION 3: SEMANTIC SEARCHES AND SEARCH HISTORY
-- ============================================================================

-- Insert search history (last 30 days)
INSERT INTO neuronip.search_history (id, user_id, query_text, result_count, created_at) VALUES
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'customer retention strategies', 12, NOW() - INTERVAL '2 hours'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'Q4 sales performance analysis', 8, NOW() - INTERVAL '5 hours'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'product recommendation algorithms', 15, NOW() - INTERVAL '1 day'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'marketing campaign ROI', 10, NOW() - INTERVAL '2 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440003', 'API rate limiting best practices', 7, NOW() - INTERVAL '3 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'customer churn prediction models', 9, NOW() - INTERVAL '4 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440005', 'data pipeline optimization', 11, NOW() - INTERVAL '5 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'A/B testing methodology', 13, NOW() - INTERVAL '6 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'revenue forecasting models', 6, NOW() - INTERVAL '7 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440006', 'ETL pipeline monitoring', 14, NOW() - INTERVAL '8 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440003', 'database indexing strategies', 8, NOW() - INTERVAL '9 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'customer segmentation analysis', 12, NOW() - INTERVAL '10 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'pricing optimization strategies', 9, NOW() - INTERVAL '11 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440005', 'machine learning model deployment', 10, NOW() - INTERVAL '12 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'email marketing effectiveness', 7, NOW() - INTERVAL '13 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'supply chain optimization', 11, NOW() - INTERVAL '14 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440003', 'microservices architecture patterns', 13, NOW() - INTERVAL '15 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'social media engagement metrics', 8, NOW() - INTERVAL '16 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'inventory management systems', 9, NOW() - INTERVAL '17 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440005', 'real-time analytics dashboards', 12, NOW() - INTERVAL '18 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'customer feedback analysis', 10, NOW() - INTERVAL '19 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'financial reporting automation', 8, NOW() - INTERVAL '20 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440003', 'container orchestration best practices', 11, NOW() - INTERVAL '21 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'content marketing strategies', 9, NOW() - INTERVAL '22 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'user behavior analytics', 13, NOW() - INTERVAL '23 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440005', 'data quality monitoring', 7, NOW() - INTERVAL '24 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'conversion rate optimization', 10, NOW() - INTERVAL '25 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'predictive maintenance models', 12, NOW() - INTERVAL '26 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440003', 'API security best practices', 8, NOW() - INTERVAL '27 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'brand sentiment analysis', 9, NOW() - INTERVAL '28 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'sales forecasting techniques', 11, NOW() - INTERVAL '29 days'),
(gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440005', 'data governance frameworks', 9, NOW() - INTERVAL '30 days');

-- Insert saved searches
INSERT INTO neuronip.saved_searches (id, name, description, query, schema_id, is_public, owner_id, tags, created_at, updated_at) VALUES
(gen_random_uuid(), 'Customer Analytics', 'Search for customer-related documents and insights', 'customer analytics and behavior', '770e8400-e29b-41d4-a716-446655440000', true, '550e8400-e29b-41d4-a716-446655440001', ARRAY['customer', 'analytics', 'business'], NOW() - INTERVAL '20 days', NOW() - INTERVAL '5 days'),
(gen_random_uuid(), 'Sales Performance', 'Q4 sales performance and trends', 'Q4 sales performance metrics', '770e8400-e29b-41d4-a716-446655440000', true, '550e8400-e29b-41d4-a716-446655440002', ARRAY['sales', 'performance', 'metrics'], NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 days'),
(gen_random_uuid(), 'Marketing Campaigns', 'Marketing campaign effectiveness and ROI', 'marketing campaign ROI and effectiveness', '770e8400-e29b-41d4-a716-446655440002', false, '550e8400-e29b-41d4-a716-446655440002', ARRAY['marketing', 'campaigns', 'roi'], NOW() - INTERVAL '10 days', NOW() - INTERVAL '1 day'),
(gen_random_uuid(), 'Product Documentation', 'Technical product documentation and guides', 'product documentation and technical guides', NULL, true, '550e8400-e29b-41d4-a716-446655440003', ARRAY['product', 'documentation', 'technical'], NOW() - INTERVAL '8 days', NOW()),
(gen_random_uuid(), 'Data Pipeline Issues', 'Common data pipeline issues and solutions', 'data pipeline errors and troubleshooting', NULL, false, '550e8400-e29b-41d4-a716-446655440005', ARRAY['data', 'pipeline', 'troubleshooting'], NOW() - INTERVAL '5 days', NOW() - INTERVAL '2 days');

-- ============================================================================
-- SECTION 4: WAREHOUSE QUERIES AND RESULTS
-- ============================================================================

-- Insert warehouse queries
WITH query_data AS (
  SELECT 
    gen_random_uuid() as id,
    '550e8400-e29b-41d4-a716-446655440001' as user_id,
    'What are the top 10 customers by revenue in the last quarter?' as query,
    'SELECT c.name, SUM(o.total) as revenue FROM customers c JOIN orders o ON c.id = o.customer_id WHERE o.order_date >= CURRENT_DATE - INTERVAL ''3 months'' GROUP BY c.id, c.name ORDER BY revenue DESC LIMIT 10;' as sql,
    '770e8400-e29b-41d4-a716-446655440001'::uuid as schema_id,
    'completed' as status,
    NOW() - INTERVAL '2 hours' as created_at,
    NOW() - INTERVAL '2 hours' + INTERVAL '1.5 seconds' as executed_at
  UNION ALL
  SELECT gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'Show me monthly sales trends for the past year', 
    'SELECT DATE_TRUNC(''month'', transaction_date) as month, SUM(amount) as total_sales FROM sales_transactions WHERE transaction_date >= CURRENT_DATE - INTERVAL ''12 months'' GROUP BY month ORDER BY month;',
    '770e8400-e29b-41d4-a716-446655440000'::uuid, 'completed', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours' + INTERVAL '2.3 seconds'
  UNION ALL
  SELECT gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'What is the average order value by customer segment?',
    'SELECT c.segment, AVG(o.total) as avg_order_value FROM customers c JOIN orders o ON c.id = o.customer_id GROUP BY c.segment ORDER BY avg_order_value DESC;',
    '770e8400-e29b-41d4-a716-446655440001'::uuid, 'completed', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '1.8 seconds'
  UNION ALL
  SELECT gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'Which marketing campaigns had the highest conversion rate?',
    'SELECT name, channel, (conversions::numeric / NULLIF(impressions, 0)) * 100 as conversion_rate FROM marketing_campaigns ORDER BY conversion_rate DESC LIMIT 10;',
    '770e8400-e29b-41d4-a716-446655440002'::uuid, 'completed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '3.1 seconds'
  UNION ALL
  SELECT gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'Show customer churn rate by month',
    'SELECT DATE_TRUNC(''month'', cancelled_at) as month, COUNT(*) as churned_customers FROM subscriptions WHERE status = ''cancelled'' GROUP BY month ORDER BY month DESC;',
    '770e8400-e29b-41d4-a716-446655440001'::uuid, 'completed', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days' + INTERVAL '2.5 seconds'
  UNION ALL
  SELECT gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440005', 'What are the top performing products by revenue?',
    'SELECT p.name, p.category, SUM(st.amount) as total_revenue FROM products p JOIN sales_transactions st ON p.id = st.product_id GROUP BY p.id, p.name, p.category ORDER BY total_revenue DESC LIMIT 20;',
    '770e8400-e29b-41d4-a716-446655440000'::uuid, 'completed', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days' + INTERVAL '4.2 seconds'
  UNION ALL
  SELECT gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'Show regional sales distribution',
    'SELECT region, SUM(amount) as total_sales, COUNT(*) as transaction_count FROM sales_transactions GROUP BY region ORDER BY total_sales DESC;',
    '770e8400-e29b-41d4-a716-446655440000'::uuid, 'completed', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days' + INTERVAL '1.9 seconds'
  UNION ALL
  SELECT gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440001', 'What is the customer lifetime value by acquisition channel?',
    'SELECT acquisition_channel, AVG(lifetime_value) as avg_ltv, COUNT(*) as customer_count FROM customers GROUP BY acquisition_channel ORDER BY avg_ltv DESC;',
    '770e8400-e29b-41d4-a716-446655440001'::uuid, 'completed', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days' + INTERVAL '2.7 seconds'
  UNION ALL
  SELECT gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440002', 'Show email campaign performance metrics',
    'SELECT campaign_id, COUNT(*) FILTER (WHERE event_type = ''open'') as opens, COUNT(*) FILTER (WHERE event_type = ''click'') as clicks FROM email_events GROUP BY campaign_id ORDER BY opens DESC;',
    '770e8400-e29b-41d4-a716-446655440002'::uuid, 'completed', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days' + INTERVAL '3.4 seconds'
  UNION ALL
  SELECT gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440005', 'What are the peak usage hours for our platform?',
    'SELECT EXTRACT(HOUR FROM timestamp) as hour, COUNT(*) as event_count FROM user_activity GROUP BY hour ORDER BY event_count DESC;',
    '770e8400-e29b-41d4-a716-446655440000'::uuid, 'completed', NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days' + INTERVAL '2.1 seconds'
)
INSERT INTO neuronip.warehouse_queries (id, user_id, natural_language_query, generated_sql, schema_id, status, created_at, executed_at)
SELECT id, user_id, query, sql, schema_id, status, created_at, executed_at FROM query_data;

-- Insert query results for completed queries
INSERT INTO neuronip.query_results (id, query_id, result_data, chart_config, chart_type, execution_time_ms, row_count, created_at)
SELECT 
  gen_random_uuid(),
  wq.id,
  '[
    {"name": "Acme Corp", "revenue": 125000.50},
    {"name": "TechStart Inc", "revenue": 98000.25},
    {"name": "Global Solutions", "revenue": 87500.75},
    {"name": "Innovation Labs", "revenue": 76500.00},
    {"name": "Digital Ventures", "revenue": 72000.30}
  ]'::jsonb,
  '{"xAxis": "name", "yAxis": "revenue", "title": "Top Customers by Revenue"}'::jsonb,
  'bar',
  1500,
  5,
  wq.executed_at
FROM neuronip.warehouse_queries wq
WHERE wq.status = 'completed'
LIMIT 10;

-- Insert query explanations
INSERT INTO neuronip.query_explanations (id, query_id, explanation_text, explanation_type, created_at)
SELECT 
  gen_random_uuid(),
  wq.id,
  'This query identifies the top 10 customers by total revenue in the last quarter. It joins the customers and orders tables, filters for orders in the last 3 months, groups by customer, and orders by total revenue in descending order.',
  'sql',
  wq.executed_at
FROM neuronip.warehouse_queries wq
WHERE wq.status = 'completed'
LIMIT 10;

-- ============================================================================
-- SECTION 5: WORKFLOWS AND EXECUTIONS
-- ============================================================================

-- Insert workflows
INSERT INTO neuronip.workflows (id, name, description, workflow_definition, enabled, created_by, created_at, updated_at) VALUES
('880e8400-e29b-41d4-a716-446655440000', 'Daily Sales Report', 'Generates daily sales reports and sends to stakeholders', 
 '{"steps": [{"id": "1", "type": "query", "config": {"sql": "SELECT * FROM sales_transactions WHERE date = CURRENT_DATE"}}, {"id": "2", "type": "transform", "config": {}}, {"id": "3", "type": "notify", "config": {"recipients": ["sales-team@acmecorp.com"]}}]}'::jsonb,
 true, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '30 days', NOW() - INTERVAL '1 day'),

('880e8400-e29b-41d4-a716-446655440001', 'Customer Churn Analysis', 'Analyzes customer churn patterns and generates alerts',
 '{"steps": [{"id": "1", "type": "query", "config": {"sql": "SELECT * FROM subscriptions WHERE status = ''cancelled''"}}, {"id": "2", "type": "analyze", "config": {"model": "churn_prediction"}}, {"id": "3", "type": "alert", "config": {"threshold": 0.7}}]}'::jsonb,
 true, '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '25 days', NOW() - INTERVAL '2 days'),

('880e8400-e29b-41d4-a716-446655440002', 'Marketing Campaign Performance', 'Tracks marketing campaign metrics and ROI',
 '{"steps": [{"id": "1", "type": "query", "config": {"sql": "SELECT * FROM marketing_campaigns"}}, {"id": "2", "type": "calculate", "config": {"metrics": ["roi", "conversion_rate"]}}, {"id": "3", "type": "dashboard", "config": {"dashboard_id": "marketing"}}]}'::jsonb,
 true, '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '20 days', NOW() - INTERVAL '3 days'),

('880e8400-e29b-41d4-a716-446655440003', 'Data Quality Check', 'Runs data quality checks on all data sources',
 '{"steps": [{"id": "1", "type": "validate", "config": {"checks": ["completeness", "accuracy", "consistency"]}}, {"id": "2", "type": "report", "config": {"format": "json"}}, {"id": "3", "type": "notify", "config": {"on_failure": true}}]}'::jsonb,
 true, '550e8400-e29b-41d4-a716-446655440005', NOW() - INTERVAL '15 days', NOW() - INTERVAL '4 days'),

('880e8400-e29b-41d4-a716-446655440004', 'Weekly Analytics Summary', 'Generates weekly analytics summary report',
 '{"steps": [{"id": "1", "type": "aggregate", "config": {"period": "week"}}, {"id": "2", "type": "visualize", "config": {"charts": ["bar", "line"]}}, {"id": "3", "type": "export", "config": {"format": "pdf"}}]}'::jsonb,
 true, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '10 days', NOW() - INTERVAL '5 days')
ON CONFLICT (name) DO NOTHING;

-- Insert workflow executions (last 30 days)
WITH execution_data AS (
  SELECT 
    gen_random_uuid() as id,
    '880e8400-e29b-41d4-a716-446655440000'::uuid as workflow_id,
    'completed' as status,
    '{"date": "2024-01-20"}'::jsonb as input_data,
    '{"total_sales": 125000, "transaction_count": 450}'::jsonb as output_data,
    NOW() - INTERVAL '1 day' as started_at,
    NOW() - INTERVAL '1 day' + INTERVAL '45 seconds' as completed_at,
    NOW() - INTERVAL '1 day' as created_at
  UNION ALL SELECT gen_random_uuid(), '880e8400-e29b-41d4-a716-446655440000'::uuid, 'completed', '{"date": "2024-01-19"}'::jsonb, '{"total_sales": 118000, "transaction_count": 420}'::jsonb, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '42 seconds', NOW() - INTERVAL '2 days'
  UNION ALL SELECT gen_random_uuid(), '880e8400-e29b-41d4-a716-446655440000'::uuid, 'completed', '{"date": "2024-01-18"}'::jsonb, '{"total_sales": 132000, "transaction_count": 480}'::jsonb, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days' + INTERVAL '38 seconds', NOW() - INTERVAL '3 days'
  UNION ALL SELECT gen_random_uuid(), '880e8400-e29b-41d4-a716-446655440001'::uuid, 'completed', '{}'::jsonb, '{"churn_risk_customers": 45, "high_risk": 12}'::jsonb, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '120 seconds', NOW() - INTERVAL '1 day'
  UNION ALL SELECT gen_random_uuid(), '880e8400-e29b-41d4-a716-446655440001'::uuid, 'completed', '{}'::jsonb, '{"churn_risk_customers": 38, "high_risk": 8}'::jsonb, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '115 seconds', NOW() - INTERVAL '2 days'
  UNION ALL SELECT gen_random_uuid(), '880e8400-e29b-41d4-a716-446655440002'::uuid, 'completed', '{}'::jsonb, '{"total_roi": 3.2, "top_campaign": "Summer Sale 2024"}'::jsonb, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '85 seconds', NOW() - INTERVAL '1 day'
  UNION ALL SELECT gen_random_uuid(), '880e8400-e29b-41d4-a716-446655440003'::uuid, 'completed', '{}'::jsonb, '{"checks_passed": 45, "checks_failed": 2}'::jsonb, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '200 seconds', NOW() - INTERVAL '1 day'
  UNION ALL SELECT gen_random_uuid(), '880e8400-e29b-41d4-a716-446655440004'::uuid, 'completed', '{"week": "2024-01-15"}'::jsonb, '{"report_generated": true, "charts": 8}'::jsonb, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days' + INTERVAL '180 seconds', NOW() - INTERVAL '7 days'
  UNION ALL SELECT gen_random_uuid(), '880e8400-e29b-41d4-a716-446655440000'::uuid, 'running', '{"date": "2024-01-21"}'::jsonb, '{}'::jsonb, NOW() - INTERVAL '10 minutes', NULL, NOW() - INTERVAL '10 minutes'
  UNION ALL SELECT gen_random_uuid(), '880e8400-e29b-41d4-a716-446655440001'::uuid, 'pending', '{}'::jsonb, '{}'::jsonb, NULL, NULL, NOW()
)
INSERT INTO neuronip.workflow_executions (id, workflow_id, status, input_data, output_data, started_at, completed_at, created_at)
SELECT id, workflow_id, status, input_data, output_data, started_at, completed_at, created_at FROM execution_data;

-- ============================================================================
-- SECTION 6: AGENTS
-- ============================================================================

-- Insert agents
INSERT INTO neuronip.agents (id, name, agent_type, config, status, performance_metrics, created_at, updated_at) VALUES
('990e8400-e29b-41d4-a716-446655440000', 'Sales Analytics Agent', 'analytics', 
 '{"model": "gpt-4", "temperature": 0.3, "max_tokens": 2000, "tools": ["query_warehouse", "generate_charts"]}'::jsonb,
 'active', '{"queries_processed": 1250, "avg_response_time_ms": 850, "success_rate": 0.96}'::jsonb,
 NOW() - INTERVAL '60 days', NOW() - INTERVAL '1 hour'),

('990e8400-e29b-41d4-a716-446655440001', 'Customer Support Agent', 'support',
 '{"model": "gpt-4", "temperature": 0.7, "max_tokens": 1500, "tools": ["search_knowledge", "create_ticket", "escalate"]}'::jsonb,
 'active', '{"tickets_resolved": 890, "avg_resolution_time_min": 12, "customer_satisfaction": 4.6}'::jsonb,
 NOW() - INTERVAL '45 days', NOW() - INTERVAL '30 minutes'),

('990e8400-e29b-41d4-a716-446655440002', 'Data Quality Agent', 'automation',
 '{"model": "gpt-4", "temperature": 0.2, "max_tokens": 1000, "tools": ["validate_data", "detect_anomalies", "generate_alerts"]}'::jsonb,
 'active', '{"checks_performed": 4500, "anomalies_detected": 23, "false_positive_rate": 0.05}'::jsonb,
 NOW() - INTERVAL '30 days', NOW() - INTERVAL '2 hours'),

('990e8400-e29b-41d4-a716-446655440003', 'Marketing Insights Agent', 'analytics',
 '{"model": "gpt-4", "temperature": 0.4, "max_tokens": 1800, "tools": ["analyze_campaigns", "predict_performance", "optimize_budget"]}'::jsonb,
 'active', '{"campaigns_analyzed": 320, "recommendations_generated": 145, "roi_improvement": 0.15}'::jsonb,
 NOW() - INTERVAL '20 days', NOW() - INTERVAL '1 hour'),

('990e8400-e29b-41d4-a716-446655440004', 'Workflow Orchestrator', 'workflow',
 '{"model": "gpt-4", "temperature": 0.1, "max_tokens": 1200, "tools": ["execute_workflow", "monitor_status", "handle_errors"]}'::jsonb,
 'active', '{"workflows_executed": 567, "success_rate": 0.94, "avg_execution_time_sec": 45}'::jsonb,
 NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 minutes')
ON CONFLICT (name) DO NOTHING;

-- Insert agent performance metrics
INSERT INTO neuronip.agent_performance (id, agent_id, metric_name, metric_value, timestamp)
SELECT 
  gen_random_uuid(),
  '990e8400-e29b-41d4-a716-446655440000'::uuid,
  'queries_processed',
  (RANDOM() * 50 + 10)::numeric,
  NOW() - (n || ' hours')::interval
FROM generate_series(0, 167) n; -- Last 7 days, hourly

INSERT INTO neuronip.agent_performance (id, agent_id, metric_name, metric_value, timestamp)
SELECT 
  gen_random_uuid(),
  '990e8400-e29b-41d4-a716-446655440001'::uuid,
  'tickets_resolved',
  (RANDOM() * 20 + 5)::numeric,
  NOW() - (n || ' hours')::interval
FROM generate_series(0, 167) n;

-- ============================================================================
-- SECTION 7: SUPPORT AGENTS, TICKETS, AND CONVERSATIONS
-- ============================================================================

-- Insert support agents
INSERT INTO neuronip.support_agents (id, name, description, enabled, config, created_at, updated_at) VALUES
('aa0e8400-e29b-41d4-a716-446655440000', 'Primary Support Agent', 'Main customer support agent handling general inquiries', true,
 '{"response_time_target": 5, "escalation_threshold": 0.8, "languages": ["en", "es"]}'::jsonb,
 NOW() - INTERVAL '60 days', NOW()),
('aa0e8400-e29b-41d4-a716-446655440001', 'Technical Support Agent', 'Handles technical issues and product questions', true,
 '{"response_time_target": 10, "escalation_threshold": 0.7, "specializations": ["api", "integrations"]}'::jsonb,
 NOW() - INTERVAL '45 days', NOW()),
('aa0e8400-e29b-41d4-a716-446655440002', 'Billing Support Agent', 'Assists with billing and subscription questions', true,
 '{"response_time_target": 3, "escalation_threshold": 0.9, "access_level": "billing"}'::jsonb,
 NOW() - INTERVAL '30 days', NOW())
ON CONFLICT (name) DO NOTHING;

-- Insert support tickets
INSERT INTO neuronip.support_tickets (id, ticket_number, customer_id, customer_email, subject, status, priority, assigned_agent_id, created_at, updated_at, resolved_at) VALUES
(gen_random_uuid(), 'TKT-2024-001234', 'CUST-001', 'customer1@example.com', 'Unable to access dashboard after login', 'resolved', 'high', 'aa0e8400-e29b-41d4-a716-446655440000'::uuid, NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(gen_random_uuid(), 'TKT-2024-001235', 'CUST-002', 'customer2@example.com', 'API rate limit exceeded error', 'in_progress', 'medium', 'aa0e8400-e29b-41d4-a716-446655440001'::uuid, NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 hour', NULL),
(gen_random_uuid(), 'TKT-2024-001236', 'CUST-003', 'customer3@example.com', 'Billing question about subscription renewal', 'resolved', 'low', 'aa0e8400-e29b-41d4-a716-446655440002'::uuid, NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
(gen_random_uuid(), 'TKT-2024-001237', 'CUST-004', 'customer4@example.com', 'Data export not working correctly', 'open', 'high', 'aa0e8400-e29b-41d4-a716-446655440001'::uuid, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NULL),
(gen_random_uuid(), 'TKT-2024-001238', 'CUST-005', 'customer5@example.com', 'Feature request: Custom dashboard widgets', 'open', 'low', NULL, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours', NULL),
(gen_random_uuid(), 'TKT-2024-001239', 'CUST-006', 'customer6@example.com', 'Integration with Salesforce not syncing', 'in_progress', 'high', 'aa0e8400-e29b-41d4-a716-446655440001'::uuid, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '2 hours', NULL),
(gen_random_uuid(), 'TKT-2024-001240', 'CUST-007', 'customer7@example.com', 'Question about data retention policies', 'resolved', 'medium', 'aa0e8400-e29b-41d4-a716-446655440000'::uuid, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours'),
(gen_random_uuid(), 'TKT-2024-001241', 'CUST-008', 'customer8@example.com', 'Report generation taking too long', 'open', 'medium', NULL, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NULL),
(gen_random_uuid(), 'TKT-2024-001242', 'CUST-009', 'customer9@example.com', 'Need help with custom query builder', 'in_progress', 'low', 'aa0e8400-e29b-41d4-a716-446655440001'::uuid, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour', NULL),
(gen_random_uuid(), 'TKT-2024-001243', 'CUST-010', 'customer10@example.com', 'Account upgrade request', 'resolved', 'low', 'aa0e8400-e29b-41d4-a716-446655440002'::uuid, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes')
ON CONFLICT (ticket_number) DO NOTHING;

-- Insert support conversations (only if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'neuronip' AND table_name = 'support_conversations') THEN
        INSERT INTO neuronip.support_conversations (id, ticket_id, message_text, sender_type, sender_id, created_at)
SELECT 
  gen_random_uuid(),
  st.id,
  CASE 
    WHEN n = 1 THEN 'Hello, I am having trouble accessing my dashboard after logging in. I keep getting an error message.'
    WHEN n = 2 THEN 'Thank you for contacting support. I can help you with that. Can you please tell me what error message you are seeing?'
    WHEN n = 3 THEN 'The error says "Session expired" but I just logged in a minute ago.'
    WHEN n = 4 THEN 'I understand. This might be a session cookie issue. Let me check your account settings and reset your session.'
    WHEN n = 5 THEN 'That would be great, thank you!'
    WHEN n = 6 THEN 'I have reset your session. Please try logging out and logging back in. This should resolve the issue.'
    WHEN n = 7 THEN 'Perfect! It works now. Thank you so much for your help!'
    ELSE 'Message ' || n
  END,
  CASE WHEN n % 2 = 1 THEN 'customer' ELSE 'agent' END,
  CASE WHEN n % 2 = 1 THEN st.customer_id ELSE 'AGENT-001' END,
  st.created_at + (n || ' minutes')::interval
FROM neuronip.support_tickets st
CROSS JOIN generate_series(1, 7) n
WHERE st.ticket_number = 'TKT-2024-001234'
LIMIT 7;

        -- Insert more conversations for other tickets
        INSERT INTO neuronip.support_conversations (id, ticket_id, message_text, sender_type, sender_id, created_at)
SELECT 
  gen_random_uuid(),
  st.id,
  CASE 
    WHEN n = 1 THEN 'I am getting "API rate limit exceeded" errors when trying to use the API.'
    WHEN n = 2 THEN 'I can help you with that. What is your current API usage and what rate limit tier are you on?'
    WHEN n = 3 THEN 'I am on the basic tier with 1000 requests per hour, but I only made about 500 requests.'
    WHEN n = 4 THEN 'Let me check your API key configuration and usage logs to investigate this issue.'
    ELSE 'Continuing investigation...'
  END,
  CASE WHEN n % 2 = 1 THEN 'customer' ELSE 'agent' END,
  CASE WHEN n % 2 = 1 THEN st.customer_id ELSE 'AGENT-002' END,
  st.created_at + (n || ' minutes')::interval
FROM neuronip.support_tickets st
CROSS JOIN generate_series(1, 5) n
WHERE st.ticket_number = 'TKT-2024-001235'
LIMIT 5;
    END IF;
END $$;

-- ============================================================================
-- SECTION 8: METRICS AND BUSINESS METRICS
-- ============================================================================

-- Insert business metrics
INSERT INTO neuronip.business_metrics (id, name, display_name, description, metric_type, formula, unit, owner_id, tags, created_at, updated_at) VALUES
(gen_random_uuid(), 'monthly_recurring_revenue', 'Monthly Recurring Revenue', 'Total monthly recurring revenue from all active subscriptions', 'sum',
 'SELECT SUM(monthly_price) FROM subscriptions WHERE status = ''active''', 'USD', '550e8400-e29b-41d4-a716-446655440001', ARRAY['revenue', 'subscriptions', 'kpi'], NOW() - INTERVAL '60 days', NOW()),
(gen_random_uuid(), 'customer_acquisition_cost', 'Customer Acquisition Cost', 'Average cost to acquire a new customer', 'avg',
 'SELECT AVG(acquisition_cost) FROM customers WHERE signup_date >= CURRENT_DATE - INTERVAL ''30 days''', 'USD', '550e8400-e29b-41d4-a716-446655440002', ARRAY['marketing', 'cost', 'cac'], NOW() - INTERVAL '50 days', NOW()),
(gen_random_uuid(), 'churn_rate', 'Monthly Churn Rate', 'Percentage of customers who cancel their subscription each month', 'custom',
 'SELECT (COUNT(*) FILTER (WHERE status = ''cancelled'')::numeric / NULLIF(COUNT(*), 0)) * 100 FROM subscriptions', 'percentage', '550e8400-e29b-41d4-a716-446655440001', ARRAY['churn', 'retention', 'kpi'], NOW() - INTERVAL '45 days', NOW()),
(gen_random_uuid(), 'average_order_value', 'Average Order Value', 'Average value of customer orders', 'avg',
 'SELECT AVG(total) FROM orders', 'USD', '550e8400-e29b-41d4-a716-446655440002', ARRAY['sales', 'orders', 'aov'], NOW() - INTERVAL '40 days', NOW()),
(gen_random_uuid(), 'conversion_rate', 'Conversion Rate', 'Percentage of visitors who become customers', 'custom',
 'SELECT (COUNT(DISTINCT customer_id)::numeric / NULLIF(COUNT(DISTINCT visitor_id), 0)) * 100 FROM website_events', 'percentage', '550e8400-e29b-41d4-a716-446655440002', ARRAY['marketing', 'conversion', 'kpi'], NOW() - INTERVAL '35 days', NOW()),
(gen_random_uuid(), 'customer_lifetime_value', 'Customer Lifetime Value', 'Predicted total revenue from a customer over their lifetime', 'avg',
 'SELECT AVG(lifetime_value) FROM customers', 'USD', '550e8400-e29b-41d4-a716-446655440001', ARRAY['customer', 'ltv', 'kpi'], NOW() - INTERVAL '30 days', NOW())
ON CONFLICT (name) DO NOTHING;

-- Insert metric catalog entries
INSERT INTO neuronip.metric_catalog (id, name, display_name, description, sql_expression, metric_type, unit, category, tags, owner_id, status, created_at, updated_at) VALUES
(gen_random_uuid(), 'total_revenue', 'Total Revenue', 'Sum of all transaction amounts', 'SELECT SUM(amount) FROM sales_transactions', 'sum', 'USD', 'revenue', ARRAY['revenue', 'sales'], '550e8400-e29b-41d4-a716-446655440001', 'approved', NOW() - INTERVAL '60 days', NOW()),
(gen_random_uuid(), 'active_customers', 'Active Customers', 'Count of customers with active subscriptions', 'SELECT COUNT(DISTINCT customer_id) FROM subscriptions WHERE status = ''active''', 'count', 'customers', 'customers', ARRAY['customers', 'subscriptions'], '550e8400-e29b-41d4-a716-446655440001', 'approved', NOW() - INTERVAL '55 days', NOW()),
(gen_random_uuid(), 'transaction_count', 'Transaction Count', 'Total number of transactions', 'SELECT COUNT(*) FROM sales_transactions', 'count', 'transactions', 'sales', ARRAY['transactions', 'sales'], '550e8400-e29b-41d4-a716-446655440002', 'approved', NOW() - INTERVAL '50 days', NOW()),
(gen_random_uuid(), 'campaign_roi', 'Campaign ROI', 'Return on investment for marketing campaigns', 'SELECT (SUM(revenue) - SUM(cost)) / NULLIF(SUM(cost), 0) * 100 FROM marketing_campaigns', 'custom', 'percentage', 'marketing', ARRAY['roi', 'marketing'], '550e8400-e29b-41d4-a716-446655440002', 'approved', NOW() - INTERVAL '45 days', NOW())
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- SECTION 9: AUDIT LOGS AND COMPLIANCE
-- ============================================================================

-- Insert audit logs (last 30 days)
INSERT INTO neuronip.audit_logs (id, user_id, action_type, resource_type, resource_id, action, details, ip_address, user_agent, status, duration_ms, created_at)
SELECT 
  gen_random_uuid(),
  CASE (n % 8)
    WHEN 0 THEN '550e8400-e29b-41d4-a716-446655440000'
    WHEN 1 THEN '550e8400-e29b-41d4-a716-446655440001'
    WHEN 2 THEN '550e8400-e29b-41d4-a716-446655440002'
    WHEN 3 THEN '550e8400-e29b-41d4-a716-446655440003'
    WHEN 4 THEN '550e8400-e29b-41d4-a716-446655440004'
    WHEN 5 THEN '550e8400-e29b-41d4-a716-446655440005'
    WHEN 6 THEN '550e8400-e29b-41d4-a716-446655440006'
    ELSE '550e8400-e29b-41d4-a716-446655440007'
  END,
  CASE (n % 7)
    WHEN 0 THEN 'query'
    WHEN 1 THEN 'agent_execution'
    WHEN 2 THEN 'workflow_execution'
    WHEN 3 THEN 'data_access'
    WHEN 4 THEN 'config_change'
    WHEN 5 THEN 'user_action'
    ELSE 'ai_action'
  END,
  CASE (n % 4)
    WHEN 0 THEN 'warehouse_query'
    WHEN 1 THEN 'workflow'
    WHEN 2 THEN 'agent'
    ELSE 'data_source'
  END,
  gen_random_uuid()::text,
  CASE (n % 5)
    WHEN 0 THEN 'executed_query'
    WHEN 1 THEN 'created_workflow'
    WHEN 2 THEN 'accessed_data'
    WHEN 3 THEN 'updated_config'
    ELSE 'ran_agent'
  END,
  jsonb_build_object('query_id', gen_random_uuid(), 'rows_returned', (RANDOM() * 1000 + 10)::int),
  '192.168.1.' || (n % 255 + 1)::text,
  'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
  CASE WHEN n % 20 = 0 THEN 'failure' ELSE 'success' END,
  (RANDOM() * 5000 + 100)::int,
  NOW() - (n || ' hours')::interval
FROM generate_series(0, 719) n; -- Last 30 days, hourly

-- Insert policies
INSERT INTO neuronip.policies (id, name, description, policy_type, policy_definition, enabled, priority, applies_to, conditions, actions, created_by, created_at, updated_at) VALUES
(gen_random_uuid(), 'GDPR Data Access Policy', 'Ensures GDPR compliance for data access', 'compliance',
 '{"rules": [{"type": "gdpr", "action": "mask_pii", "fields": ["email", "phone", "ssn"]}]}'::jsonb,
 true, 100, ARRAY['query', 'data_access'], '{"region": "EU"}'::jsonb, '{"mask_fields": true}'::jsonb,
 '550e8400-e29b-41d4-a716-446655440000', NOW() - INTERVAL '60 days', NOW()),
(gen_random_uuid(), 'Query Rate Limiting', 'Limits query execution rate per user', 'usage_limit',
 '{"max_queries_per_hour": 100, "max_queries_per_day": 1000}'::jsonb,
 true, 50, ARRAY['query'], '{"user_role": "analyst"}'::jsonb, '{"throttle": true, "notify": true}'::jsonb,
 '550e8400-e29b-41d4-a716-446655440000', NOW() - INTERVAL '50 days', NOW()),
(gen_random_uuid(), 'PII Data Filtering', 'Filters PII data from query results', 'result_filter',
 '{"filter_pii": true, "allowed_roles": ["admin"]}'::jsonb,
 true, 200, ARRAY['query'], '{"contains_pii": true}'::jsonb, '{"filter": true, "log": true}'::jsonb,
 '550e8400-e29b-41d4-a716-446655440000', NOW() - INTERVAL '40 days', NOW())
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- SECTION 10: USAGE METRICS AND OBSERVABILITY
-- ============================================================================

-- Insert usage metrics (last 30 days)
INSERT INTO neuronip.usage_metrics (id, user_id, resource_type, resource_id, metric_name, metric_value, unit, metadata, timestamp)
SELECT 
  gen_random_uuid(),
  CASE (n % 8)
    WHEN 0 THEN '550e8400-e29b-41d4-a716-446655440000'
    WHEN 1 THEN '550e8400-e29b-41d4-a716-446655440001'
    WHEN 2 THEN '550e8400-e29b-41d4-a716-446655440002'
    WHEN 3 THEN '550e8400-e29b-41d4-a716-446655440003'
    WHEN 4 THEN '550e8400-e29b-41d4-a716-446655440004'
    WHEN 5 THEN '550e8400-e29b-41d4-a716-446655440005'
    WHEN 6 THEN '550e8400-e29b-41d4-a716-446655440006'
    ELSE '550e8400-e29b-41d4-a716-446655440007'
  END,
  CASE (n % 5)
    WHEN 0 THEN 'query'
    WHEN 1 THEN 'agent'
    WHEN 2 THEN 'workflow'
    WHEN 3 THEN 'api_call'
    ELSE 'model_inference'
  END,
  gen_random_uuid()::text,
  CASE (n % 4)
    WHEN 0 THEN 'execution_time_ms'
    WHEN 1 THEN 'rows_processed'
    WHEN 2 THEN 'tokens_used'
    ELSE 'cost_usd'
  END,
  CASE 
    WHEN n % 4 = 0 THEN (RANDOM() * 5000 + 100)::numeric
    WHEN n % 4 = 1 THEN (RANDOM() * 10000 + 100)::numeric
    WHEN n % 4 = 2 THEN (RANDOM() * 50000 + 1000)::numeric
    ELSE (RANDOM() * 0.5 + 0.01)::numeric
  END,
  CASE (n % 4)
    WHEN 0 THEN 'ms'
    WHEN 1 THEN 'rows'
    WHEN 2 THEN 'tokens'
    ELSE 'USD'
  END,
  jsonb_build_object('source', 'api', 'endpoint', '/api/v1/query'),
  NOW() - (n || ' hours')::interval
FROM generate_series(0, 719) n; -- Last 30 days, hourly

-- Insert cost tracking
INSERT INTO neuronip.cost_tracking (id, user_id, resource_type, resource_id, cost_amount, currency, cost_category, billing_period_start, billing_period_end, created_at)
SELECT 
  gen_random_uuid(),
  CASE (n % 8)
    WHEN 0 THEN '550e8400-e29b-41d4-a716-446655440000'
    WHEN 1 THEN '550e8400-e29b-41d4-a716-446655440001'
    WHEN 2 THEN '550e8400-e29b-41d4-a716-446655440002'
    WHEN 3 THEN '550e8400-e29b-41d4-a716-446655440003'
    WHEN 4 THEN '550e8400-e29b-41d4-a716-446655440004'
    WHEN 5 THEN '550e8400-e29b-41d4-a716-446655440005'
    WHEN 6 THEN '550e8400-e29b-41d4-a716-446655440006'
    ELSE '550e8400-e29b-41d4-a716-446655440007'
  END,
  CASE (n % 4)
    WHEN 0 THEN 'query'
    WHEN 1 THEN 'agent'
    WHEN 2 THEN 'workflow'
    ELSE 'api_call'
  END,
  gen_random_uuid()::text,
  (RANDOM() * 50 + 1)::numeric,
  'USD',
  CASE (n % 6)
    WHEN 0 THEN 'compute'
    WHEN 1 THEN 'storage'
    WHEN 2 THEN 'api_calls'
    WHEN 3 THEN 'model_inference'
    WHEN 4 THEN 'data_transfer'
    ELSE 'other'
  END,
  DATE_TRUNC('month', NOW() - (n || ' days')::interval),
  DATE_TRUNC('month', NOW() - (n || ' days')::interval) + INTERVAL '1 month' - INTERVAL '1 day',
  NOW() - (n || ' days')::interval
FROM generate_series(0, 89) n; -- Last 3 months, daily

-- Insert system logs (only if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'neuronip' AND table_name = 'system_logs') THEN
        INSERT INTO neuronip.system_logs (id, log_type, level, message, context, timestamp)
SELECT 
  gen_random_uuid(),
  CASE (n % 5)
    WHEN 0 THEN 'query'
    WHEN 1 THEN 'agent'
    WHEN 2 THEN 'workflow'
    WHEN 3 THEN 'system'
    ELSE 'error'
  END,
  CASE 
    WHEN n % 100 = 0 THEN 'error'
    WHEN n % 20 = 0 THEN 'warning'
    WHEN n % 5 = 0 THEN 'info'
    ELSE 'debug'
  END,
  CASE 
    WHEN n % 100 = 0 THEN 'Query execution failed: Connection timeout'
    WHEN n % 20 = 0 THEN 'High memory usage detected: 85%'
    WHEN n % 5 = 0 THEN 'Workflow execution completed successfully'
    ELSE 'Query executed: ' || n || ' rows returned'
  END,
  jsonb_build_object('component', 'api', 'version', '1.0.0'),
  NOW() - (n || ' hours')::interval
FROM generate_series(0, 167) n; -- Last 7 days, hourly
    END IF;
END $$;

-- ============================================================================
-- SECTION 11: INGESTION JOBS
-- ============================================================================

-- Insert ingestion jobs
INSERT INTO neuronip.ingestion_jobs (id, data_source_id, job_type, status, config, progress, rows_processed, started_at, completed_at, created_at, updated_at)
SELECT 
  gen_random_uuid(),
  CASE (n % 5)
    WHEN 0 THEN '660e8400-e29b-41d4-a716-446655440000'::uuid
    WHEN 1 THEN '660e8400-e29b-41d4-a716-446655440001'::uuid
    WHEN 2 THEN '660e8400-e29b-41d4-a716-446655440002'::uuid
    WHEN 3 THEN '660e8400-e29b-41d4-a716-446655440003'::uuid
    ELSE '660e8400-e29b-41d4-a716-446655440004'::uuid
  END,
  CASE (n % 4)
    WHEN 0 THEN 'sync'
    WHEN 1 THEN 'cdc'
    WHEN 2 THEN 'etl'
    ELSE 'file_upload'
  END,
  CASE 
    WHEN n % 10 = 0 THEN 'failed'
    WHEN n % 5 = 0 THEN 'running'
    ELSE 'completed'
  END,
  jsonb_build_object('batch_size', 1000, 'parallelism', 4),
  jsonb_build_object('progress_percent', CASE WHEN n % 5 = 0 THEN 45 ELSE 100 END, 'current_table', 'sales_transactions'),
  CASE WHEN n % 5 = 0 THEN NULL ELSE (RANDOM() * 100000 + 10000)::bigint END,
  NOW() - (n || ' hours')::interval,
  CASE WHEN n % 5 = 0 THEN NULL ELSE NOW() - (n || ' hours')::interval + INTERVAL '30 minutes' END,
  NOW() - (n || ' hours')::interval,
  NOW() - (n || ' hours')::interval + INTERVAL '30 minutes'
FROM generate_series(0, 167) n; -- Last 7 days, hourly

-- ============================================================================
-- SECTION 12: KNOWLEDGE BASE AND EMBEDDINGS
-- ============================================================================

-- Insert knowledge documents (we'll skip embeddings as they require vector generation)
INSERT INTO neuronip.knowledge_documents (id, title, content, content_type, source, metadata, created_at, updated_at)
VALUES
(gen_random_uuid(), 'Getting Started Guide', 'This comprehensive guide will help you get started with NeuronIP platform. Learn how to set up your account, connect data sources, configure your first warehouse schema, and run your first queries. The platform provides powerful tools for semantic search, data warehouse Q&A, customer support automation, and compliance analytics.', 'document', 'internal', '{"author": "support-team", "version": "1.0", "category": "getting-started"}'::jsonb, NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days'),
(gen_random_uuid(), 'API Documentation', 'Complete API reference documentation for all endpoints, authentication methods, request/response formats, rate limiting, and error handling. Includes code examples in Python, JavaScript, and Go.', 'document', 'internal', '{"author": "dev-team", "version": "1.0", "category": "api"}'::jsonb, NOW() - INTERVAL '55 days', NOW() - INTERVAL '55 days'),
(gen_random_uuid(), 'Troubleshooting Guide', 'Common issues and their solutions. Learn how to troubleshoot connection problems, query errors, performance issues, authentication failures, and data synchronization problems.', 'document', 'internal', '{"author": "support-team", "version": "1.0", "category": "troubleshooting"}'::jsonb, NOW() - INTERVAL '50 days', NOW() - INTERVAL '50 days'),
(gen_random_uuid(), 'Best Practices', 'Best practices for data modeling, query optimization, and platform usage to maximize performance and efficiency. Includes recommendations for schema design, indexing strategies, and resource management.', 'document', 'internal', '{"author": "data-team", "version": "1.0", "category": "best-practices"}'::jsonb, NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days'),
(gen_random_uuid(), 'Architecture Overview', 'Detailed architecture overview explaining how NeuronIP processes data, handles queries, manages resources, and integrates with external systems. Covers the microservices architecture, data flow, and component interactions.', 'document', 'internal', '{"author": "engineering-team", "version": "1.0", "category": "architecture"}'::jsonb, NOW() - INTERVAL '40 days', NOW() - INTERVAL '40 days'),
(gen_random_uuid(), 'Security Guide', 'Security best practices including authentication, authorization, data encryption, compliance requirements, and audit logging. Learn how to secure your data and comply with regulations like GDPR and HIPAA.', 'document', 'internal', '{"author": "security-team", "version": "1.0", "category": "security"}'::jsonb, NOW() - INTERVAL '35 days', NOW() - INTERVAL '35 days'),
(gen_random_uuid(), 'Performance Tuning', 'Performance tuning guide covering query optimization, indexing strategies, resource management, caching, and monitoring. Learn how to improve query performance and reduce costs.', 'document', 'internal', '{"author": "performance-team", "version": "1.0", "category": "performance"}'::jsonb, NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days'),
(gen_random_uuid(), 'Integration Guide', 'Step-by-step integration guides for connecting external systems, APIs, and data sources. Includes examples for Salesforce, Slack, email systems, and custom integrations.', 'document', 'internal', '{"author": "integration-team", "version": "1.0", "category": "integration"}'::jsonb, NOW() - INTERVAL '25 days', NOW() - INTERVAL '25 days'),
(gen_random_uuid(), 'Data Modeling', 'Data modeling best practices for organizing your data warehouse and creating efficient schemas. Learn about star schemas, snowflake schemas, and dimensional modeling.', 'document', 'internal', '{"author": "data-team", "version": "1.0", "category": "data-modeling"}'::jsonb, NOW() - INTERVAL '20 days', NOW() - INTERVAL '20 days'),
(gen_random_uuid(), 'Query Optimization', 'Query optimization techniques to improve performance and reduce execution times. Covers SQL best practices, index usage, query planning, and execution strategies.', 'document', 'internal', '{"author": "performance-team", "version": "1.0", "category": "optimization"}'::jsonb, NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days'),
(gen_random_uuid(), 'Customer Support Workflows', 'Guide to setting up and managing customer support workflows using AI agents. Learn how to automate ticket routing, response generation, and escalation procedures.', 'document', 'internal', '{"author": "support-team", "version": "1.0", "category": "workflows"}'::jsonb, NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),
(gen_random_uuid(), 'Compliance Analytics', 'How to use compliance analytics features to monitor policy adherence, detect anomalies, and generate audit reports. Includes GDPR, HIPAA, and SOC 2 compliance examples.', 'document', 'internal', '{"author": "compliance-team", "version": "1.0", "category": "compliance"}'::jsonb, NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
(gen_random_uuid(), 'Agent Configuration', 'Complete guide to configuring AI agents for different use cases. Learn about agent types, tool selection, memory policies, and performance tuning.', 'document', 'internal', '{"author": "ai-team", "version": "1.0", "category": "agents"}'::jsonb, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
(gen_random_uuid(), 'Semantic Search Setup', 'How to set up and configure semantic search for your knowledge base. Learn about embedding models, collection management, and search optimization.', 'document', 'internal', '{"author": "search-team", "version": "1.0", "category": "semantic-search"}'::jsonb, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
(gen_random_uuid(), 'Warehouse Q&A Guide', 'Using the warehouse Q&A feature to ask natural language questions and get SQL queries, charts, and explanations. Learn how to configure schemas and optimize query generation.', 'document', 'internal', '{"author": "analytics-team", "version": "1.0", "category": "warehouse-qa"}'::jsonb, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(gen_random_uuid(), 'Deployment Guide', 'Production deployment guide covering Docker, Kubernetes, Helm charts, monitoring, and scaling. Learn how to deploy NeuronIP in production environments.', 'document', 'internal', '{"author": "ops-team", "version": "1.0", "category": "deployment"}'::jsonb, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days');

-- ============================================================================
-- SECTION 13: COMPLETION MESSAGE
-- ============================================================================

DO $$
BEGIN
  RAISE NOTICE 'Demo data insertion completed successfully!';
  RAISE NOTICE 'Total records inserted across all tables.';
  RAISE NOTICE 'The database is now populated with realistic demo data.';
END $$;
