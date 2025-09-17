package restaurant

import (
	"context"
	"fmt"
	"sync"

	m "github.com/ChristopherLeo15/opentable/restaurant/internal/model"
	metagw "github.com/ChristopherLeo15/opentable/restaurant/internal/gateway/metadata/http"
	metamodel "github.com/ChristopherLeo15/opentable/metadata/model"
)

// Controller keeps in-memory restaurants (no repository layer) and calls metadata via gateway.
type Controller struct {
	mu    sync.RWMutex
	items []m.Restaurant

	metagw *metagw.Gateway
}

func New(gw *metagw.Gateway) *Controller {
	return &Controller{metagw: gw, items: make([]m.Restaurant, 0, 16)}
}

// Seed is optional; handy for quick manual testing
func (c *Controller) Seed() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = append(c.items,
		m.Restaurant{ID: 1, MetadataID: 1, DisplayName: "Casa Verde — Centro"},
		m.Restaurant{ID: 2, MetadataID: 2, DisplayName: "Pasta Nostra — Roma"},
		m.Restaurant{ID: 3, MetadataID: 3, DisplayName: "Tokyo Bite — GDL"},
	)
}

func (c *Controller) List(ctx context.Context) []m.Restaurant {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]m.Restaurant, len(c.items))
	copy(out, c.items)
	return out
}

func (c *Controller) GetByID(ctx context.Context, id int) (m.Restaurant, *metamodel.Metadata, error) {
	if id <= 0 {
		return m.Restaurant{}, nil, fmt.Errorf("id must be positive")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, r := range c.items {
		if r.ID == id {
			md, err := c.metagw.GetByID(ctx, r.MetadataID)
			if err != nil {
				// return restaurant even if metadata lookup fails
				return r, nil, nil
			}
			return r, &md, nil
		}
	}
	return m.Restaurant{}, nil, fmt.Errorf("restaurant not found")
}

func (c *Controller) Add(ctx context.Context, x m.Restaurant) (m.Restaurant, error) {
	if x.DisplayName == "" {
		return m.Restaurant{}, fmt.Errorf("display_name is required")
	}
	if x.MetadataID <= 0 {
		return m.Restaurant{}, fmt.Errorf("metadata_id must be positive")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if x.ID == 0 {
		next := 1
		for _, cur := range c.items {
			if cur.ID >= next {
				next = cur.ID + 1
			}
		}
		x.ID = next
	}

	c.items = append(c.items, x)
	return x, nil
}
