package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	mdgw "github.com/ChristopherLeo15/opentable/restaurant/internal/gateway/metadata/http"
	ctl  "github.com/ChristopherLeo15/opentable/restaurant/internal/controller/restaurant"
	h    "github.com/ChristopherLeo15/opentable/restaurant/internal/handler/http"

	"github.com/ChristopherLeo15/opentable/pkg/discovery/consul"
	"github.com/ChristopherLeo15/opentable/pkg/discovery/memorypackage"
	discovery "github.com/ChristopherLeo15/opentable/pkg/registry"
)

func main() {
	var (
		port       = flag.Int("port", 8082, "http port")
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
	const serviceName = "restaurant"
	instanceID := discovery.GenerateInstanceID(serviceName)
	hostPort := fmt.Sprintf("%s:%d", *host, *port)
	if err := reg.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		log.Fatal(err)
	}
	defer reg.Deregister(ctx, instanceID, serviceName)

	md := mdgw.New(reg)
	c  := ctl.New(md)
	hd := h.New(c)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", hd.Health)
	mux.HandleFunc("/restaurant", func(w http.ResponseWriter, r *http.Request) {
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
	log.Printf("restaurant listening on :%d", *port)
	log.Fatal(srv.ListenAndServe())
}