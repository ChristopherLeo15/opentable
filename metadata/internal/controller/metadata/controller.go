package metadata

import (
	"context"
	"fmt"

	m "github.com/ChristopherLeo15/opentable/metadata/model"
)

type Repository interface {
	GetAll() []m.Metadata
	GetByID(id int) (m.Metadata, error)
	Add(x m.Metadata)
}

type Controller struct {
	repo Repository
}

func New(repo Repository) *Controller {
	return &Controller{repo: repo}
}

func (c *Controller) List(ctx context.Context) []m.Metadata {
	return c.repo.GetAll()
}

func (c *Controller) GetByID(ctx context.Context, id int) (m.Metadata, error) {
	if id <= 0 {
		return m.Metadata{}, fmt.Errorf("id must be positive")
	}
	return c.repo.GetByID(id)
}

func (c *Controller) Add(ctx context.Context, x m.Metadata) (m.Metadata, error) {
	// Simple validation
	if x.Name == "" {
		return m.Metadata{}, fmt.Errorf("name is required")
	}
	if x.CuisineType == "" {
		return m.Metadata{}, fmt.Errorf("cuisine_type is required")
	}

	// Auto-assign ID if not provided
	if x.ID == 0 {
		next := 1
		for _, cur := range c.repo.GetAll() {
			if cur.ID >= next {
				next = cur.ID + 1
			}
		}
		x.ID = next
	}

	c.repo.Add(x)
	return x, nil
}