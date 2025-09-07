package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	meta "github.com/ChristopherLeo15/opentable/metadata/pkg"
	discovery "github.com/ChristopherLeo15/opentable/pkg/registry"
)

type Gateway struct{ reg discovery.Registry }

func New(reg discovery.Registry) *Gateway { return &Gateway{reg} }

func (g *Gateway) get(ctx context.Context, path string, out any) error {
	addrs, err := g.reg.ServiceAddress(ctx, "metadata")
	if err != nil || len(addrs) == 0 {
		return fmt.Errorf("metadata not found: %w", err)
	}
	resp, err := http.Get("http://" + addrs[0] + path)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("metadata %s -> %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (g *Gateway) GetByID(ctx context.Context, id int) (meta.Metadata, error) {
	var m meta.Metadata
	if err := g.get(ctx, fmt.Sprintf("/metadata?id=%d", id), &m); err != nil {
		return meta.Metadata{}, err
	}
	return m, nil
}