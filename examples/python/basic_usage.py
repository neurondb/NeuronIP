"""
NeuronIP Python SDK - Basic Usage Example
"""

from neuronip import NeuronIPClient

# Initialize client
client = NeuronIPClient(
    base_url="http://localhost:8082/api/v1",
    api_key="your-api-key-here"
)

# Health check
health = client.health_check()
print(f"API Status: {health['status']}")

# Semantic search
results = client.semantic_search("machine learning algorithms", limit=10)
print(f"Found {len(results.get('results', []))} results")

# Warehouse query
query_result = client.warehouse_query(
    "Show me total sales by region",
    schema_id=None
)
print(f"Query executed: {query_result.get('query_id')}")

# Create ingestion job
job = client.create_ingestion_job(
    data_source_id="your-data-source-id",
    job_type="sync",
    config={"mode": "full", "tables": ["users", "orders"]}
)
print(f"Created job: {job['id']}")

# List ingestion jobs
jobs = client.list_ingestion_jobs(limit=10)
print(f"Found {len(jobs)} jobs")

# Get audit logs
audit_logs = client.get_audit_logs(limit=20)
print(f"Retrieved {len(audit_logs)} audit log entries")
