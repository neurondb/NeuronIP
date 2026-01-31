package ingestion

import (
	"fmt"
	"sync"
)

/* ConnectorFramework provides plugin-based connector architecture */
type ConnectorFramework struct {
	connectors map[string]ConnectorFactory
	mu         sync.RWMutex
}

/* NewConnectorFramework creates a new connector framework */
func NewConnectorFramework() *ConnectorFramework {
	return &ConnectorFramework{
		connectors: make(map[string]ConnectorFactory),
	}
}

/* RegisterConnector registers a connector factory */
func (cf *ConnectorFramework) RegisterConnector(connectorType string, factory ConnectorFactory) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.connectors[connectorType] = factory
}

/* GetConnector creates a connector instance */
func (cf *ConnectorFramework) GetConnector(connectorType string) (Connector, error) {
	cf.mu.RLock()
	factory, exists := cf.connectors[connectorType]
	cf.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connector type %s not registered", connectorType)
	}

	return factory(connectorType), nil
}

/* ListConnectors returns all registered connector types */
func (cf *ConnectorFramework) ListConnectors() []string {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	types := make([]string, 0, len(cf.connectors))
	for t := range cf.connectors {
		types = append(types, t)
	}
	return types
}

/* ConnectorFactory creates a new connector instance */
type ConnectorFactory func(connectorType string) Connector

/* ConnectorRegistry manages connector metadata and versioning */
type ConnectorRegistry struct {
	connectors map[string]ConnectorMetadata
	mu         sync.RWMutex
}

/* NewConnectorRegistry creates a new connector registry */
func NewConnectorRegistry() *ConnectorRegistry {
	return &ConnectorRegistry{
		connectors: make(map[string]ConnectorMetadata),
	}
}

/* RegisterConnectorMetadata registers connector metadata */
func (cr *ConnectorRegistry) RegisterConnectorMetadata(connectorType string, metadata ConnectorMetadata) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.connectors[connectorType] = metadata
}

/* GetConnectorMetadata retrieves connector metadata */
func (cr *ConnectorRegistry) GetConnectorMetadata(connectorType string) (*ConnectorMetadata, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	metadata, exists := cr.connectors[connectorType]
	if !exists {
		return nil, fmt.Errorf("connector metadata for %s not found", connectorType)
	}

	return &metadata, nil
}

/* ListConnectorMetadata returns all connector metadata */
func (cr *ConnectorRegistry) ListConnectorMetadata() []ConnectorMetadata {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	metadataList := make([]ConnectorMetadata, 0, len(cr.connectors))
	for _, metadata := range cr.connectors {
		metadataList = append(metadataList, metadata)
	}
	return metadataList
}

/* ConnectorMetadata represents connector metadata */
type ConnectorMetadata struct {
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
	ConfigSchema map[string]interface{} `json:"config_schema,omitempty"`
}
