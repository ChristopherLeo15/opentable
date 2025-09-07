package memorypackage

import (
	"context"
	"sync"

	discovery "github.com/ChristopherLeo15/opentable/pkg/registry"
)

type Registry struct {
	mu   sync.RWMutex
	svcs map[string][]string
}

func New() *Registry { return &Registry{svcs: map[string][]string{}} }

func (r *Registry) Register(ctx context.Context, _ , service, hostPort string) error {
	r.mu.Lock(); defer r.mu.Unlock()
	r.svcs[service] = append(r.svcs[service], hostPort)
	return nil
}
func (r *Registry) Deregister(ctx context.Context, _ , service string) error {
	r.mu.Lock(); defer r.mu.Unlock()
	delete(r.svcs, service)
	return nil
}
func (r *Registry) ServiceAddress(ctx context.Context, service string) ([]string, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	addrs := r.svcs[service]
	if len(addrs) == 0 { return nil, discovery.ErrNotFound }
	cp := make([]string, len(addrs))
	copy(cp, addrs)
	return cp, nil
}
func (r *Registry) ReportHealthyState(_, _ string) error { return nil }