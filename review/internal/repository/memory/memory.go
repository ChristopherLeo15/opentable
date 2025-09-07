package memory

import "github.com/ChristopherLeo15/opentable/review/internal/model"

type Repo struct {
	db map[int]model.Review
	id int
}

func New() *Repo { return &Repo{db: map[int]model.Review{}, id: 1} }

func (r *Repo) ListByRestaurant(id int) []model.Review {
	out := []model.Review{}
	for _, v := range r.db {
		if v.RestaurantID == id {
			out = append(out, v)
		}
	}
	return out
}
func (r *Repo) Create(rv model.Review) model.Review {
	rv.ID = r.id
	r.id++
	r.db[rv.ID] = rv
	return rv
}