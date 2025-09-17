package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	meta "github.com/ChristopherLeo15/opentable/metadata/model"
)

// Gateway discovers the metadata service via Consul (with a METADATA_URL fallback).
type Gateway struct {
	fallbackBase string // from METADATA_URL
	consulAddr   string // from CONSUL_HTTP_ADDR
	client       *http.Client

	mu        sync.RWMutex
	cachedURL string
	expires   time.Time
}

func New() *Gateway {
	base := os.Getenv("METADATA_URL")
	if base == "" {
		base = "http://metadata:8081"
	}
	consul := os.Getenv("CONSUL_HTTP_ADDR")
	if consul == "" {
		consul = "http://consul:8500"
	}
	tr := &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		MaxIdleConns:          10,
	}
	return &Gateway{
		fallbackBase: base,
		consulAddr:   consul,
		client:       &http.Client{Transport: tr},
	}
}

// baseURL returns a usable base URL for the metadata service.
// It always returns a string (either a cached Consul value or the fallback).
func (g *Gateway) baseURL(ctx context.Context) string {
	// Use cached value if still valid
	g.mu.RLock()
	if time.Now().Before(g.expires) && g.cachedURL != "" {
		u := g.cachedURL
		g.mu.RUnlock()
		return u
	}
	g.mu.RUnlock()

	// Query Consul: GET /v1/health/service/metadata?passing=true
	type svc struct {
		Service struct {
			Address string
			Port    int
		}
		Node struct {
			Address string
		}
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, g.consulAddr+"/v1/health/service/metadata?passing=true", nil)
	resp, err := g.client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var arr []svc
		if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&arr); err == nil && len(arr) > 0 {
			addr := arr[0].Service.Address
			if addr == "" {
				addr = arr[0].Node.Address
			}
			if addr != "" && arr[0].Service.Port != 0 {
				u := fmt.Sprintf("http://%s:%d", addr, arr[0].Service.Port)
				g.mu.Lock()
				g.cachedURL = u
				g.expires = time.Now().Add(30 * time.Second)
				g.mu.Unlock()
				return u
			}
		}
		// if JSON decode fails or array empty, fall through to fallback
	}

	// Fallback to env/base
	return g.fallbackBase
}

// get performs a GET to path on the resolved baseURL and decodes JSON into out.
// It ALWAYS returns an error (nil on success).
func (g *Gateway) get(ctx context.Context, path string, out any) error {
	// Add a short per-call timeout if caller didn't
	if _, has := ctx.Deadline(); !has {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	u := g.baseURL(ctx) + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("metadata %s -> %d", path, resp.StatusCode)
	}

	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
}

// GetByID fetches one metadata record by id. It ALWAYS returns (value, error).
func (g *Gateway) GetByID(ctx context.Context, id int) (meta.Metadata, error) {
	var m meta.Metadata
	if err := g.get(ctx, fmt.Sprintf("/metadata?id=%d", id), &m); err != nil {
		return meta.Metadata{}, err
	}
	return m, nil
}
