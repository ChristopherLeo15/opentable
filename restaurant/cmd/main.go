package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	ctrl "github.com/ChristopherLeo15/opentable/restaurant/internal/controller/restaurant"
	httpr "github.com/ChristopherLeo15/opentable/restaurant/internal/handler/http"
	gw "github.com/ChristopherLeo15/opentable/restaurant/internal/gateway/metadata/http"
)

func main() {
	var portFlag = flag.Int("port", 8082, "port to listen on")
	flag.Parse()

	port := *portFlag
	if env := os.Getenv("PORT"); env != "" {
		if p, err := strconv.Atoi(env); err == nil {
			port = p
		}
	}

	metadataGW := gw.New()
	c := ctrl.New(metadataGW)
	hdlr := httpr.New(c)
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           hdlr.Router(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Register in Consul
	serviceName := getenvDefault("SERVICE_NAME", "restaurant")
	consulAddr := getenvDefault("CONSUL_HTTP_ADDR", "http://consul:8500")
	serviceID := fmt.Sprintf("%s-%d", serviceName, port)
	if err := registerWithConsul(consulAddr, serviceID, serviceName, "restaurant", port, "/healthz"); err != nil {
		log.Printf("consul register failed: %v", err)
	}

	go func() {
		log.Printf("%s service listening on :%d", serviceName, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// graceful shutdown & de-register
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = deregisterFromConsul(consulAddr, serviceID)

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	} else {
		log.Println("server stopped gracefully")
	}
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func registerWithConsul(consul, id, name, dnsName string, port int, healthPath string) error {
	payload := map[string]any{
		"ID":      id,
		"Name":    name,
		"Address": dnsName, // docker-compose DNS name
		"Port":    port,
		"Check": map[string]any{
			"HTTP":     fmt.Sprintf("http://%s:%d%s", dnsName, port, healthPath),
			"Interval": "10s",
			"DeregisterCriticalServiceAfter": "1m",
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPut, consul+"/v1/agent/service/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("consul register: status %d", resp.StatusCode)
	}
	return nil
}

func deregisterFromConsul(consul, id string) error {
	req, _ := http.NewRequest(http.MethodPut, consul+"/v1/agent/service/deregister/"+id, nil)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}