package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ChristopherLeo15/opentable/pkg/discovery/consul"
	"github.com/ChristopherLeo15/opentable/pkg/discovery/memorypackage"
	discovery "github.com/ChristopherLeo15/opentable/pkg/registry"

	ctl "github.com/ChristopherLeo15/opentable/review/internal/controller/review"
	h   "github.com/ChristopherLeo15/opentable/review/internal/handler/http"
	repo "github.com/ChristopherLeo15/opentable/review/internal/repository/memory"
)

func main() {
	var (
		port       = flag.Int("port", 8083, "http port")
		host       = flag.String("host", "127.0.0.1", "bind host")
		consulAddr = flag.String("consul", "", "consul addr")
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
	const serviceName = "review"
	instanceID := discovery.GenerateInstanceID(serviceName)
	hostPort := fmt.Sprintf("%s:%d", *host, *port)
	if err := reg.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		log.Fatal(err)
	}
	defer reg.Deregister(ctx, instanceID, serviceName)

	rp := repo.New()
	c := ctl.New(rp)
	hd := h.New(c)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", hd.Health)
	mux.HandleFunc("/review", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet  { hd.List(w, r); return }
		if r.Method == http.MethodPost { hd.Create(w, r); return }
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Printf("review listening on :%d", *port)
	log.Fatal(srv.ListenAndServe())
}