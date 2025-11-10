package loadbalancer

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	Attempts int = iota
	Retry
)

type LoadBalancer struct {
	serversPool *ServersPool
}

func CreateLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		serversPool: &ServersPool{},
	}
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func (lb LoadBalancer) Handle(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := lb.serversPool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		log.Printf("Proxy connection to server %s", peer.URL)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func (lb *LoadBalancer) HealthCheck() {
	for _, s := range lb.serversPool.servers {
		status := "up"
		alive := s.isServisAlive()
		s.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", s.URL, status)
	}
}

func (lb *LoadBalancer) AddServer(serverUrl *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(serverUrl)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
		retries := GetRetryFromContext(request)
		if retries < 3 {
			select {
			case <-time.After(10 * time.Millisecond):
				ctx := context.WithValue(request.Context(), Retry, retries+1)
				proxy.ServeHTTP(writer, request.WithContext(ctx))
			}
			return
		}

		// after 3 retries, mark this backend as down
		lb.serversPool.MarkBackendStatus(serverUrl, false)

		// if the same request routing for few attempts with different backends, increase the count
		attempts := GetAttemptsFromContext(request)
		log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
		ctx := context.WithValue(request.Context(), Attempts, attempts+1)
		lb.Handle(writer, request.WithContext(ctx))
	}

	lb.serversPool.AddBackend(
		&Server{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		},
	)
	log.Printf("Configured server: %s\n", serverUrl)

}
