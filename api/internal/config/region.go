package config

import (
	"fmt"
)

/* RegionConfig represents region configuration */
type RegionConfig struct {
	CurrentRegion string `json:"current_region"`
	PrimaryRegion string `json:"primary_region"`
	EnableReplication bool `json:"enable_replication"`
}

/* LoadRegionConfig loads region configuration from environment */
func LoadRegionConfig() *RegionConfig {
	return &RegionConfig{
		CurrentRegion: getEnv("REGION", "us-east-1"),
		PrimaryRegion: getEnv("PRIMARY_REGION", "us-east-1"),
		EnableReplication: getEnv("ENABLE_REPLICATION", "false") == "true",
	}
}

// getEnv is already defined in config.go - no need to redeclare

// getEnv is already defined in config.go (same package)

/* IsPrimaryRegion checks if current region is primary */
func (c *RegionConfig) IsPrimaryRegion() bool {
	return c.CurrentRegion == c.PrimaryRegion
}

/* GetRegionEndpoint gets the endpoint for a region */
func GetRegionEndpoint(regionCode string) string {
	// In production, this would be loaded from configuration or service discovery
	endpoints := map[string]string{
		"us-east-1": "https://api-us-east-1.neurondb.com",
		"us-west-2": "https://api-us-west-2.neurondb.com",
		"eu-west-1": "https://api-eu-west-1.neurondb.com",
		"ap-southeast-1": "https://api-ap-southeast-1.neurondb.com",
	}

	if endpoint, ok := endpoints[regionCode]; ok {
		return endpoint
	}

	return fmt.Sprintf("https://api-%s.neurondb.com", regionCode)
}
