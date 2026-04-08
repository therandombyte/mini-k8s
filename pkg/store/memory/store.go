// one in-memory map: store object by resource, namespace and name
// protect with a mutex and a resourceVersion counter so list response can carry a version

package memory

import (
	"context"
	"sync"

	"github.com/therandombyte/mini-k8s/pkg/store"
)

// Implementation detail, hence not exported
// a 3 level map to store "pod:default:nginx":"*v1.Pod"(pointer to object)
// rv is resourceversion counter
type memoryStore struct {
	mu sync.RWMutex
	objects map[string]map[string]map[string]any
	rv int64
}

// constructor that returns a store backed by memorystore
// to be invoked by API server
func New() store.Store {
	return &memoryStore{
		objects: make(map[string]map[string]map[string]any),
	}
}



func (s *memoryStore) Create(ctx context.Context, resource, namespace, name string, obj any) error {
	return nil
}
func (s *memoryStore) Get(ctx context.Context, resource, namespace, name string) (any, bool, error) {
	return nil,false,nil
}
func (s *memoryStore) List(ctx context.Context, resource, namespace string) ([]any, int64, error) {
	return nil,0,nil
}
func (s *memoryStore) Update(ctx context.Context, resource, namespace, name string, obj any) error {
	return nil
}
func (s *memoryStore) Delete(ctx context.Context, resource, namespace, name string) error {
	return nil
}
