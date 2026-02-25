package servicemesh

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// RoundRobinLoadBalancer implements round-robin load balancing
type RoundRobinLoadBalancer struct {
	counter uint64
	log     *logrus.Logger
}

// LeastConnLoadBalancer implements least-connection load balancing
type LeastConnLoadBalancer struct {
	connections map[string]int64
	mu          sync.RWMutex
	log         *logrus.Logger
}

// RandomLoadBalancer implements random load balancing
type RandomLoadBalancer struct {
	log *logrus.Logger
}

// WeightedLoadBalancer implements weighted load balancing
type WeightedLoadBalancer struct {
	log *logrus.Logger
}

// NewLoadBalancer creates a new load balancer based on strategy
func NewLoadBalancer(strategy string, log *logrus.Logger) LoadBalancer {
	switch strategy {
	case "round_robin":
		return &RoundRobinLoadBalancer{log: log}
	case "least_conn":
		return &LeastConnLoadBalancer{
			connections: make(map[string]int64),
			log:         log,
		}
	case "random":
		return &RandomLoadBalancer{log: log}
	case "weighted":
		return &WeightedLoadBalancer{log: log}
	default:
		return &RoundRobinLoadBalancer{log: log}
	}
}

// RoundRobinLoadBalancer implementation

func (lb *RoundRobinLoadBalancer) Select(services []*Service) (*Service, error) {
	if len(services) == 0 {
		return nil, fmt.Errorf("no services available")
	}
	
	index := atomic.AddUint64(&lb.counter, 1) % uint64(len(services))
	return services[index], nil
}

func (lb *RoundRobinLoadBalancer) UpdateStrategy(strategy string) error {
	return fmt.Errorf("cannot change strategy on existing load balancer")
}

// LeastConnLoadBalancer implementation

func (lb *LeastConnLoadBalancer) Select(services []*Service) (*Service, error) {
	if len(services) == 0 {
		return nil, fmt.Errorf("no services available")
	}
	
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	var selected *Service
	minConn := int64(-1)
	
	for _, service := range services {
		conn := lb.connections[service.ID]
		if minConn == -1 || conn < minConn {
			minConn = conn
			selected = service
		}
	}
	
	if selected != nil {
		lb.mu.RUnlock()
		lb.mu.Lock()
		lb.connections[selected.ID]++
		lb.mu.Unlock()
		lb.mu.RLock()
	}
	
	return selected, nil
}

func (lb *LeastConnLoadBalancer) UpdateStrategy(strategy string) error {
	return fmt.Errorf("cannot change strategy on existing load balancer")
}

func (lb *LeastConnLoadBalancer) ReleaseConnection(serviceID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	if count, exists := lb.connections[serviceID]; exists && count > 0 {
		lb.connections[serviceID]--
	}
}

// RandomLoadBalancer implementation

func (lb *RandomLoadBalancer) Select(services []*Service) (*Service, error) {
	if len(services) == 0 {
		return nil, fmt.Errorf("no services available")
	}
	
	index := rand.Intn(len(services))
	return services[index], nil
}

func (lb *RandomLoadBalancer) UpdateStrategy(strategy string) error {
	return fmt.Errorf("cannot change strategy on existing load balancer")
}

// WeightedLoadBalancer implementation

func (lb *WeightedLoadBalancer) Select(services []*Service) (*Service, error) {
	if len(services) == 0 {
		return nil, fmt.Errorf("no services available")
	}
	
	// For now, treat all services with equal weight
	// In production, you'd read weights from service metadata
	index := rand.Intn(len(services))
	return services[index], nil
}

func (lb *WeightedLoadBalancer) UpdateStrategy(strategy string) error {
	return fmt.Errorf("cannot change strategy on existing load balancer")
}
