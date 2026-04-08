// decouple API server from store implementation, so it can be changed later
// store will be in-memory for now

package store

import "context"

type Store interface {
	Create(ctx context.Context, resource, namespace, name string, obj any) error
	Get(ctx context.Context, resource, namespace, name string) (any, bool, error)
	List(ctx context.Context, resource, namespace string) ([]any, int64, error)
	Update(ctx context.Context, resource, namespace, name string, obj any) error
	Delete(ctx context.Context, resource, namespace, name string) error
}
