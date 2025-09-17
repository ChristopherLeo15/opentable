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

// Gateway discovers the metadata service via Consul (no env fallback).
type Gateway struct {
	consulAddr string
	client     *http.Client

	mu        sync.RWMutex
	cachedURL string
	expires   time.Time
}

func New() *Gateway {
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
		consulAddr: consul,
		client:     &http.Client{Transport: tr},
	}
}

// baseURL resolves metadata via Consul and always either returns a URL or panics back to caller via error on use.
func (g *Gateway) baseURL(ctx context.Context) (string, error) {
	// hit cache
	g.mu.RLock()
	if time.Now().Before(g.expires) && g.cachedURL != "" {
		u := g.cachedURL
		g.mu.RUnlock()
		return u, nil
	}
	g.mu.RUnlock()

	// ask Consul: GET /v1/health/service/metadata?passing=true
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
	if err != nil {
		return "", fmt.Errorf("consul query failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("consul query status: %d", resp.StatusCode)
	}

	var arr []svc
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&arr); err != nil {
		return "", fmt.Errorf("consul decode failed: %w", err)
	}
	if len(arr) == 0 {
		return "", fmt.Errorf("no healthy metadata instances in consul")
	}

	addr := arr[0].Service.Address
	if addr == "" {
		addr = arr[0].Node.Address
	}
	if addr == "" || arr[0].Service.Port == 0 {
		return "", fmt.Errorf("consul result missing address/port")
	}

	u := fmt.Sprintf("http://%s:%d", addr, arr[0].Service.Port)
	g.mu.Lock()
	g.cachedURL = u
	g.expires = time.Now().Add(30 * time.Second)
	g.mu.Unlock()
	return u, nil
}

// ResolveBaseURL returns the resolved URL (for /debug/metadata) or an error.
func (g *Gateway) ResolveBaseURL(ctx context.Context) (string, error) {
	return g.baseURL(ctx)
}

// Health calls metadata /healthz and returns (url, status, error).
func (g *Gateway) Health(ctx context.Context) (string, int, error) {
	if _, has := ctx.Deadline(); !has {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}
	base, err := g.baseURL(ctx)
	if err != nil {
		return "", 0, err
	}
	u := base + "/healthz"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return u, 0, err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return u, 0, err
	}
	defer resp.Body.Close()
	return u, resp.StatusCode, nil
}

func (g *Gateway) get(ctx context.Context, path string, out any) error {
	if _, has := ctx.Deadline(); !has {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}
	base, err := g.baseURL(ctx)
	if err != nil {
		return err
	}
	u := base + path

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

func (g *Gateway) GetByID(ctx context.Context, id int) (meta.Metadata, error) {
	var m meta.Metadata
	if err := g.get(ctx, fmt.Sprintf("/metadata?id=%d", id), &m); err != nil {
		return meta.Metadata{}, err
	}
	return m, nil
}