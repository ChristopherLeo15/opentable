package memory

import (
	"errors"
	"sync"

	m "github.com/ChristopherLeo15/opentable/review/internal/model"
)

var ErrNotFound = errors.New("review not found")

// Repo: tiny in-memory store, safe for concurrent access.
type Repo struct {
	mu   sync.RWMutex
	data []m.Review
}

func New() *Repo {
	return &Repo{data: make([]m.Review, 0, 32)}
}

func (r *Repo) nextID() int {
	max := 0
	for _, v := range r.data {
		if v.ID > max {
			max = v.ID
		}
	}
	return max + 1
}

func (r *Repo) Create(x m.Review) m.Review {
	r.mu.Lock()
	defer r.mu.Unlock()
	if x.ID == 0 {
		x.ID = r.nextID()
	}
	r.data = append(r.data, x)
	return x
}

func (r *Repo) ListByRestaurant(restaurantID int) []m.Review {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]m.Review, 0, 8)
	for _, v := range r.data {
		if v.RestaurantID == restaurantID {
			out = append(out, v)
		}
	}
	return out
}