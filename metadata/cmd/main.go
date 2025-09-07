package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	ctrl "github.com/ChristopherLeo15/opentable/metadata/internal/controller/metadata"
	httph "github.com/ChristopherLeo15/opentable/metadata/internal/handler/http"
	repo "github.com/ChristopherLeo15/opentable/metadata/internal/repository/memory"

	"github.com/ChristopherLeo15/opentable/pkg/discovery/consul"
	"github.com/ChristopherLeo15/opentable/pkg/discovery/memorypackage"
	discovery "github.com/ChristopherLeo15/opentable/pkg/registry"
)

func main() {
	var (
		port       = flag.Int("port", 8081, "http port")
		host       = flag.String("host", "127.0.0.1", "bind host")
		consulAddr = flag.String("consul", "", "consul addr (e.g. 127.0.0.1:8500)")
	)
	flag.Parse()

	var reg discovery.Registry
	if *consulAddr == "" {
		reg = memorypackage.New()
	} else {
		r, err := consul.NewRegistry(*consulAddr)
		if err != nil { log.Fatal(err) }
		reg = r
	}

	ctx := context.Background()
	const serviceName = "metadata"
	instanceID := discovery.GenerateInstanceID(serviceName)
	hostPort := fmt.Sprintf("%s:%d", *host, *port)

	if err := reg.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		log.Fatal(err)
	}
	defer reg.Deregister(ctx, instanceID, serviceName)

	rp := repo.New()
	c := ctrl.New(rp)
	h := httph.New(c)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/metadata", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if r.URL.Query().Get("id") != "" { h.Get(w, r); return }
			h.List(w, r); return
		}
		if r.Method == http.MethodPost { h.Create(w, r); return }
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Printf("metadata listening on :%d", *port)
	log.Fatal(srv.ListenAndServe())
}