package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	ctrl "github.com/ChristopherLeo15/opentable/metadata/internal/controller/metadata"
	httph "github.com/ChristopherLeo15/opentable/metadata/internal/handler/http"
	repo "github.com/ChristopherLeo15/opentable/metadata/internal/repository/memory"
)

func main() {
	// Allow PORT via flag or env
	var portFlag = flag.Int("port", 8081, "port to listen on")
	flag.Parse()

	port := *portFlag
	if env := os.Getenv("PORT"); env != "" {
		if p, err := strconv.Atoi(env); err == nil {
			port = p
		}
	}

	r := repo.New()
	c := ctrl.New(r)
	h := httph.New(c)

	// HTTP server with timeouts
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           h.Router(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("metadata service listening on :%d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown (Ctrl + C)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	} else {
		log.Println("server stopped gracefully")
	}
}