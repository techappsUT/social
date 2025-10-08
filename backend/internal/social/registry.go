// path: backend/internal/social/registry.go
package social

import (
	"fmt"
	"sync"
)

// AdapterRegistry manages all registered social platform adapters
type AdapterRegistry struct {
	adapters map[PlatformType]SocialAdapter
	mu       sync.RWMutex
}

// NewAdapterRegistry creates a new adapter registry
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[PlatformType]SocialAdapter),
	}
}

// Register adds a new adapter to the registry
func (r *AdapterRegistry) Register(adapter SocialAdapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	platform := adapter.GetPlatformName()
	if _, exists := r.adapters[platform]; exists {
		return fmt.Errorf("adapter for platform %s already registered", platform)
	}

	r.adapters[platform] = adapter
	return nil
}

// Get retrieves an adapter by platform type
func (r *AdapterRegistry) Get(platform PlatformType) (SocialAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[platform]
	if !exists {
		return nil, fmt.Errorf("adapter for platform %s not found", platform)
	}

	return adapter, nil
}

// GetAll returns all registered adapters
func (r *AdapterRegistry) GetAll() map[PlatformType]SocialAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[PlatformType]SocialAdapter, len(r.adapters))
	for k, v := range r.adapters {
		result[k] = v
	}
	return result
}

// ListPlatforms returns all supported platform types
func (r *AdapterRegistry) ListPlatforms() []PlatformType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	platforms := make([]PlatformType, 0, len(r.adapters))
	for platform := range r.adapters {
		platforms = append(platforms, platform)
	}
	return platforms
}
