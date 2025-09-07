package restaurant

import (
    "context"
    "errors"

    model "github.com/ChristopherLeo15/opentable/restaurant/internal/model"
    meta  "github.com/ChristopherLeo15/opentable/metadata/pkg"
)

type Metadata interface {
    GetByID(ctx context.Context, id int) (meta.Metadata, error)
}

type Controller struct {
    md Metadata
    db map[int]model.Restaurant
    id int
}

func New(md Metadata) *Controller {
    return &Controller{
        md: md,
        db: map[int]model.Restaurant{},
        id: 1,
    }
}

func (c *Controller) List() []model.Restaurant {
    out := make([]model.Restaurant, 0, len(c.db))
    for _, v := range c.db {
        out = append(out, v)
    }
    return out
}

func (c *Controller) Create(ctx context.Context, metadataID int, displayName string) (model.Restaurant, error) {
    if displayName == "" {
        return model.Restaurant{}, errors.New("display_name required")
    }
    // Validate existence in metadata service (we ignore the returned fields)
    if _, err := c.md.GetByID(ctx, metadataID); err != nil {
        return model.Restaurant{}, errors.New("unknown metadata_id")
    }
    r := model.Restaurant{ID: c.id, MetadataID: metadataID, DisplayName: displayName}
    c.db[c.id] = r
    c.id++
    return r, nil
}