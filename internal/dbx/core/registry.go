package core

import (
	"sort"
	"sync"
)

// Registry registers and resolves database adapters by type.
type Registry interface {
	// Register stores a valid adapter or returns a typed registration error.
	Register(adapter Adapter) error

	// Get returns a registered adapter or a typed unsupported-database error.
	Get(dbType DBType) (Adapter, error)

	// List returns registered adapter metadata in deterministic type order.
	List() []AdapterInfo
}

// AdapterRegistry is an in-memory registry with per-instance isolation.
type AdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[DBType]Adapter
}

// NewRegistry returns an empty isolated adapter registry.
func NewRegistry() *AdapterRegistry {
	return &AdapterRegistry{adapters: map[DBType]Adapter{}}
}

// Register stores a valid adapter or returns a typed registration error.
func (r *AdapterRegistry) Register(adapter Adapter) error {
	if adapter == nil {
		return InvalidAdapterError("nil adapter")
	}
	dbType := adapter.Type()
	if dbType == "" {
		return InvalidAdapterError("empty database type")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.adapters[dbType]; exists {
		return DuplicateAdapterError(dbType)
	}
	r.adapters[dbType] = adapter
	return nil
}

// Get returns a registered adapter or a typed unsupported-database error.
func (r *AdapterRegistry) Get(dbType DBType) (Adapter, error) {
	r.mu.RLock()
	adapter, ok := r.adapters[dbType]
	r.mu.RUnlock()
	if !ok {
		return nil, UnsupportedDatabaseError(dbType)
	}
	return adapter, nil
}

// List returns registered adapter metadata in deterministic type order.
func (r *AdapterRegistry) List() []AdapterInfo {
	r.mu.RLock()
	infos := make([]AdapterInfo, 0, len(r.adapters))
	for dbType, adapter := range r.adapters {
		infos = append(infos, AdapterInfo{Type: dbType, DisplayName: adapter.DisplayName()})
	}
	r.mu.RUnlock()

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Type < infos[j].Type
	})
	return infos
}
