package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type Backend struct {
	URL   *url.URL
	Alive bool
	Proxy *httputil.ReverseProxy

	mu sync.RWMutex
}

func NewBackend(rawURL string) (*Backend, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	proxy.Transport = transport

	return &Backend{
		URL:   target,
		Alive: true,
		Proxy: proxy,
	}, nil
}

func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Alive = alive
}

func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.Alive
}

type LoadBalancer struct {
	backends []*Backend

	current uint64

	mu sync.RWMutex
}

func NewLoadBalancer(backends []*Backend) *LoadBalancer {
	return &LoadBalancer{
		backends: backends,
	}
}

func (lb *LoadBalancer) NextBackend() *Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	total := len(lb.backends)

	if total == 0 {
		return nil
	}

	start := atomic.AddUint64(
		&lb.current,
		1,
	)

	for i := 0; i < total; i++ {

		index := int(
			(start + uint64(i)) %
				uint64(total),
		)

		backend := lb.backends[index]

		if backend.IsAlive() {
			return backend
		}
	}

	return nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.NextBackend()

	if backend == nil {

		http.Error(
			w,
			"All backend servers are unavailable",
			http.StatusServiceUnavailable,
		)

		return
	}

	log.Printf(
		"Request %s %s -> %s",
		r.Method,
		r.URL.Path,
		backend.URL.String(),
	)

	backend.Proxy.ServeHTTP(w, r)
}

func checkBackend(backend *Backend) {

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(
		backend.URL.String() + "/health",
	)

	if err != nil {

		if backend.IsAlive() {
			log.Printf(
				"Backend DOWN: %s",
				backend.URL.String(),
			)
		}

		backend.SetAlive(false)

		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		if backend.IsAlive() {
			log.Printf(
				"Backend DOWN: %s status=%d",
				backend.URL.String(),
				resp.StatusCode,
			)
		}

		backend.SetAlive(false)

		return
	}

	if !backend.IsAlive() {

		log.Printf(
			"Backend RECOVERED: %s",
			backend.URL.String(),
		)
	}

	backend.SetAlive(true)
}

func startHealthChecker(ctx context.Context, backends []*Backend) {
	ticker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()

	for {

		select {

		case <-ticker.C:

			for _, backend := range backends {
				go checkBackend(backend)
			}

		case <-ctx.Done():
			return
		}
	}
}

func main() {
	backendURLs := []string{
		"http://localhost:9001",
		"http://localhost:9002",
		"http://localhost:9003",
	}
	var backends []*Backend
	for _, rawURL := range backendURLs {
		backend, err := NewBackend(rawURL)
		if err != nil {
			log.Fatalf(
				"Failed to create backend: %v",
				err,
			)
		}
		backends = append(backends, backend)
	}
	lb := NewLoadBalancer(backends)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           lb,
		ReadHeaderTimeout: 5 * time.Second,
	}
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()
	go startHealthChecker(ctx, backends)
	go func() {
		log.Println(
			"Load balancer started on :8080",
		)
		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("Load balancer error: %v", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println(
		"Shutting down load balancer...",
	)
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Load balancer stopped")
}
