// one in-memory map: store object by resource, namespace and name
// protect with a mutex and a resourceVersion counter so list response can carry a version

package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

// Implementation detail, hence not exported
// data is a map whose each value is also a map
// Nested map: resource -> (namespace/name -> obj)
// Splitting the map by resource makes it easy to implement API collections like GET /pods
// Storing ns/name as single string prevents exta nested level map
// listRV is resourceversion counter
type Store struct {
	mu sync.RWMutex
	data map[string]map[string]any
	listRV atomic.Int64
}

// NOTE: Every object needs a struct, a constuctor and helper functions
// What: constructor that returns a store backed by memorystore
// Why: to be invoked by API server
// Initializing the data map keeps the other code free to check for nil
func New() *Store {
	return &Store{
		data: map[string]map[string]any{},
	}
}

// key should be unique, so ns/name is the choice
func key(namespace, name string) string {
	return namespace + "/" + name
}

// rough analogue of etcd's index used for resourceversion in k8s
func (s *Store) nextListRV() int64 {
	return s.listRV.Add(1)
}
 
// handlers will call Create for Pods, Nodes, Deployments
// return error will converted to http error codes by handlers
func (s *Store) Create(ctx context.Context, resource, namespace, name string, obj any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
 
	// ensure map exists, check for existing obj, assign obj, ser object RV if it has metadata

	// lazy initialization
	if s.data[resource] == nil {
		s.data[resource] = map[string]any{}
	}

	k := key(namespace, name)
	if _,exists := s.data[resource][k];exists {
		return fmt.Errorf("%s %q already exists", resource, k)
	}

	s.data[resource][k] = obj
	s.nextListRV()
	return nil
}

// backs the GET /api/v1/pods/{name} or /nodes/{name}
func (s *Store) Get(ctx context.Context, resource, namespace, name string) (any, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	//lookup and return obj
	obj, ok := s.data[resource][key(namespace,name)]
	return obj, ok, nil
}

// returned rv is used to update PodList.Metadata.ResourceVersion for future watch mechanism
func (s *Store) List(ctx context.Context, resource, namespace string) ([]any, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// collect all objects under resource+namespace, return rv

	items := make([]any, 0, len(s.data[resource]))
	for key, obj := range s.data[resource] {
		if namespace == "" || strings.HasPrefix(key, namespace+"/") {
			items = append(items, obj)
		}
	}

	return items, s.listRV.Load(), nil
}

// for pod status updates
func (s *Store) Update(ctx context.Context, resource, namespace, name string, obj any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// ensure obj exists, assign new rv, replace

	if s.data[resource] == nil {
		s.data[resource] = map[string]any{}
	}

	s.data[resource][key(namespace, name)] = obj
	s.nextListRV()
	return nil
}

// delete from store
func (s *Store) Delete(ctx context.Context, resource, namespace, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// delete if present

	if s.data[resource] != nil {
		delete(s.data[resource], key(namespace, name))
	}

	s.nextListRV()
	return nil
}
