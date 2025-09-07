package metadata

import (
	m "github.com/ChristopherLeo15/opentable/metadata/pkg"
	"github.com/ChristopherLeo15/opentable/metadata/internal/repository/memory"
)

type Controller struct{ repo *memory.Repo }

func New(repo *memory.Repo) *Controller { return &Controller{repo: repo} }

func (c *Controller) List() []m.Metadata           { return c.repo.All() }
func (c *Controller) Get(id int) (m.Metadata, bool) { return c.repo.ByID(id) }
func (c *Controller) Add(x m.Metadata)             { c.repo.Add(x) }