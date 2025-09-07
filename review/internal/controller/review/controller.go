package review

import "github.com/ChristopherLeo15/opentable/review/internal/model"

type Store interface {
	ListByRestaurant(int) []model.Review
	Create(model.Review) model.Review
}

type Controller struct{ s Store }

func New(s Store) *Controller { return &Controller{s} }

func (c *Controller) ListFor(id int) []model.Review { return c.s.ListByRestaurant(id) }
func (c *Controller) Create(r model.Review) model.Review { return c.s.Create(r) }