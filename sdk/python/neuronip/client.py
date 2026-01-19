"""
NeuronIP Python Client
"""

import requests
from typing import Optional, Dict, Any, List


class NeuronIPClient:
    """Main client for NeuronIP API"""
    
    def __init__(self, base_url: str, api_key: Optional[str] = None):
        """
        Initialize NeuronIP client
        
        Args:
            base_url: Base URL of the NeuronIP API
            api_key: API key for authentication
        """
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self.session = requests.Session()
        
        if api_key:
            self.session.headers.update({
                'Authorization': f'Bearer {api_key}',
                'Content-Type': 'application/json'
            })
    
    def _request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        """Make HTTP request"""
        url = f"{self.base_url}{endpoint}"
        response = self.session.request(method, url, **kwargs)
        response.raise_for_status()
        return response.json()
    
    def health_check(self) -> Dict[str, Any]:
        """Check API health"""
        return self._request('GET', '/health')
    
    def semantic_search(self, query: str, limit: int = 10) -> Dict[str, Any]:
        """Perform semantic search"""
        return self._request('POST', '/semantic/search', json={
            'query': query,
            'limit': limit
        })
    
    def warehouse_query(self, query: str, schema_id: Optional[str] = None) -> Dict[str, Any]:
        """Execute warehouse query"""
        payload = {'query': query}
        if schema_id:
            payload['schema_id'] = schema_id
        return self._request('POST', '/warehouse/query', json=payload)
    
    def create_ingestion_job(self, data_source_id: str, job_type: str, config: Dict[str, Any]) -> Dict[str, Any]:
        """Create an ingestion job"""
        return self._request('POST', '/ingestion/jobs', json={
            'data_source_id': data_source_id,
            'job_type': job_type,
            'config': config
        })
    
    def list_ingestion_jobs(self, data_source_id: Optional[str] = None, limit: int = 100) -> List[Dict[str, Any]]:
        """List ingestion jobs"""
        params = {'limit': limit}
        if data_source_id:
            params['data_source_id'] = data_source_id
        return self._request('GET', '/ingestion/jobs', params=params)
    
    def get_metric(self, metric_id: str) -> Dict[str, Any]:
        """Get a metric"""
        return self._request('GET', f'/metrics/{metric_id}')
    
    def create_metric(self, metric: Dict[str, Any]) -> Dict[str, Any]:
        """Create a metric"""
        return self._request('POST', '/metrics', json=metric)
    
    def get_audit_logs(self, filters: Optional[Dict[str, Any]] = None, limit: int = 100) -> List[Dict[str, Any]]:
        """Get audit logs"""
        params = {'limit': limit}
        if filters:
            params.update(filters)
        return self._request('GET', '/audit/events', params=params)
