package resources

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

/* ResourceType represents a type of resource */
type ResourceType string

const (
	ResourceTypeCPU    ResourceType = "cpu"
	ResourceTypeMemory ResourceType = "memory"
	ResourceTypeDisk   ResourceType = "disk"
	ResourceTypeNetwork ResourceType = "network"
	ResourceTypeConnections ResourceType = "connections"
)

/* ResourceUsage represents resource usage */
type ResourceUsage struct {
	Type      ResourceType
	Used      int64
	Limit     int64
	Unit      string
	Timestamp time.Time
}

/* ResourceQuota represents a resource quota */
type ResourceQuota struct {
	Type  ResourceType
	Limit int64
	Unit  string
}

/* ResourceManager manages resource quotas and usage */
type ResourceManager struct {
	quotas map[string]map[ResourceType]*ResourceQuota
	usage  map[string]map[ResourceType]*ResourceUsage
	mu     sync.RWMutex
}

/* NewResourceManager creates a new resource manager */
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		quotas: make(map[string]map[ResourceType]*ResourceQuota),
		usage:  make(map[string]map[ResourceType]*ResourceUsage),
	}
}

/* SetQuota sets a resource quota for a user/tenant */
func (rm *ResourceManager) SetQuota(userID string, quota *ResourceQuota) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.quotas[userID] == nil {
		rm.quotas[userID] = make(map[ResourceType]*ResourceQuota)
	}
	rm.quotas[userID][quota.Type] = quota
}

/* GetQuota gets a resource quota for a user/tenant */
func (rm *ResourceManager) GetQuota(userID string, resourceType ResourceType) (*ResourceQuota, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if quotas, exists := rm.quotas[userID]; exists {
		if quota, ok := quotas[resourceType]; ok {
			return quota, true
		}
	}
	return nil, false
}

/* RecordUsage records resource usage */
func (rm *ResourceManager) RecordUsage(userID string, usage *ResourceUsage) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.usage[userID] == nil {
		rm.usage[userID] = make(map[ResourceType]*ResourceUsage)
	}
	rm.usage[userID][usage.Type] = usage
}

/* CheckQuota checks if usage is within quota */
func (rm *ResourceManager) CheckQuota(userID string, resourceType ResourceType, requested int64) (bool, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	quota, exists := rm.GetQuota(userID, resourceType)
	if !exists {
		// No quota set, allow
		return true, nil
	}

	usage, exists := rm.usage[userID]
	if !exists {
		usage = make(map[ResourceType]*ResourceUsage)
	}

	currentUsage, exists := usage[resourceType]
	if !exists {
		currentUsage = &ResourceUsage{
			Type: resourceType,
			Used: 0,
		}
	}

	if currentUsage.Used+requested > quota.Limit {
		return false, fmt.Errorf("quota exceeded for %s: %d/%d", resourceType, currentUsage.Used+requested, quota.Limit)
	}

	return true, nil
}

/* GetUsage returns current resource usage */
func (rm *ResourceManager) GetUsage(userID string, resourceType ResourceType) (*ResourceUsage, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if usage, exists := rm.usage[userID]; exists {
		if resUsage, ok := usage[resourceType]; ok {
			return resUsage, true
		}
	}
	return nil, false
}

/* GetAllUsage returns all resource usage for a user */
func (rm *ResourceManager) GetAllUsage(userID string) map[ResourceType]*ResourceUsage {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	usage := make(map[ResourceType]*ResourceUsage)
	if userUsage, exists := rm.usage[userID]; exists {
		for k, v := range userUsage {
			usage[k] = v
		}
	}
	return usage
}

/* ResourceLimitMiddleware creates middleware to enforce resource limits */
func ResourceLimitMiddleware(rm *ResourceManager, resourceType ResourceType, limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from context
			userID := getUserIDFromContext(r.Context())
			if userID == "" {
				userID = "anonymous"
			}

			// Check quota
			allowed, err := rm.CheckQuota(userID, resourceType, 1)
			if !allowed {
				http.Error(w, fmt.Sprintf("Resource limit exceeded: %v", err), http.StatusTooManyRequests)
				return
			}

			// Record usage
			rm.RecordUsage(userID, &ResourceUsage{
				Type:      resourceType,
				Used:      1,
				Limit:     limit,
				Timestamp: time.Now(),
			})

			next.ServeHTTP(w, r)
		})
	}
}

func getUserIDFromContext(ctx context.Context) string {
	// Try to extract user ID from context
	// This is a placeholder - actual implementation depends on auth system
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}
