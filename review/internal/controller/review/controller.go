package review

import (
	"fmt"

	m "github.com/ChristopherLeo15/opentable/review/internal/model"
)

// Interface for saving and retrieving reviews.
type Store interface {
	Create(x m.Review) m.Review
	ListByRestaurant(restaurantID int) []m.Review
}

type Controller struct {
	s Store
}

func New(s Store) *Controller { return &Controller{s: s} }

func (c *Controller) ListFor(restaurantID int) []m.Review {
	if restaurantID <= 0 {
		return nil
	}
	return c.s.ListByRestaurant(restaurantID)
}

func (c *Controller) Create(r m.Review) (m.Review, error) {
	// Simple validation
	if err := r.Validate(); err != nil {
		return m.Review{}, err
	}

	out := c.s.Create(r)
	if out.ID <= 0 {
		return m.Review{}, fmt.Errorf("failed to create review")
	}
	return out, nil
}