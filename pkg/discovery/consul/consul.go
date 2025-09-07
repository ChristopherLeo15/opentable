package consul

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	discovery "github.com/ChristopherLeo15/opentable/pkg/registry"
	consul "github.com/hashicorp/consul/api"
)

type Registry struct{ client *consul.Client }

func NewRegistry(addr string) (*Registry, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = addr
	c, err := consul.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Registry{client: c}, nil
}

func (r *Registry) Register(ctx context.Context, instanceID, serviceName, hostPort string) error {
	host, portStr, _ := strings.Cut(hostPort, ":")
	port, _ := strconv.Atoi(portStr)

	reg := &consul.AgentServiceRegistration{
		ID:      instanceID,
		Name:    serviceName,
		Port:    port,
		Address: host,
		Check: &consul.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s/health", hostPort),
			Interval: "10s",
			Timeout:  "2s",
		},
	}
	return r.client.Agent().ServiceRegister(reg)
}

func (r *Registry) Deregister(ctx context.Context, instanceID, _ string) error {
	return r.client.Agent().ServiceDeregister(instanceID)
}

func (r *Registry) ServiceAddress(ctx context.Context, serviceName string) ([]string, error) {
	svcs, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil || len(svcs) == 0 {
		return nil, discovery.ErrNotFound
	}
	out := make([]string, 0, len(svcs))
	for _, s := range svcs {
		out = append(out, fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port))
	}
	return out, nil
}

func (r *Registry) ReportHealthyState(_, _ string) error { return nil }