package memory

import (
	"errors"
	"sync"

	m "github.com/ChristopherLeo15/opentable/metadata/model"
)

var (
	ErrNotFound = errors.New("metadata not found")
)

type Repo struct {
	// Mutex for safe concurrent access
	mu   sync.RWMutex
	data []m.Metadata
}

func New() *Repo {
	return &Repo{data: make([]m.Metadata, 0, 16)}
}

func (r *Repo) GetAll() []m.Metadata {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]m.Metadata, len(r.data))
	copy(out, r.data)
	return out
}

func (r *Repo) GetByID(id int) (m.Metadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, x := range r.data {
		if x.ID == id {
			return x, nil
		}
	}
	return m.Metadata{}, ErrNotFound
}

func (r *Repo) Add(x m.Metadata) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data = append(r.data, x)
}