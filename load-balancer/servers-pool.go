package loadbalancer

import (
	"log"
	"net/url"
	"sync/atomic"
)

// ServersPool holds information about reachable backends
type ServersPool struct {
	servers []*Server
	current uint64
}

// AddBackend to the server pool
func (s *ServersPool) AddBackend(server *Server) {
	s.servers = append(s.servers, server)
}

// NextIndex atomically increase the counter and return an index
func (s *ServersPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.servers)))
}

// MarkBackendStatus changes a status of a backend
func (s *ServersPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.servers {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

// GetNextPeer returns next active peer to take a connection
func (s *ServersPool) GetNextPeer() *Server {
	// loop entire backends to find out an Alive backend
	next := s.NextIndex()
	l := len(s.servers) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(s.servers)     // take an index by modding
		if s.servers[idx].IsAlive() { // if we have an alive backend, use it and store if its not the original one
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.servers[idx]
		}
	}
	return nil
}

// HealthCheck pings the backends and update the status
func (sp *ServersPool) HealthCheck() {
	for _, s := range sp.servers {
		status := "up"
		alive := s.isServisAlive()
		s.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", s.URL, status)
	}
}
