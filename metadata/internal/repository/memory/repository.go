package memory

import m "github.com/ChristopherLeo15/opentable/metadata/pkg"

type Repo struct {
	data []m.Metadata
}

func New() *Repo {
	return &Repo{data: []m.Metadata{}}
}

func (r *Repo) All() []m.Metadata {
	out := make([]m.Metadata, len(r.data))
	copy(out, r.data)
	return out
}

func (r *Repo) ByID(id int) (m.Metadata, bool) {
	for _, v := range r.data {
		if v.ID == id {
			return v, true
		}
	}
	return m.Metadata{}, false
}

func (r *Repo) Add(x m.Metadata) {
	r.data = append(r.data, x)
}