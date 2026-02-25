package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/hbf-agent/internal/config"
	"github.com/yourusername/hbf-agent/internal/firewall"
	"github.com/yourusername/hbf-agent/internal/health"
	"github.com/yourusername/hbf-agent/internal/metrics"
	"github.com/yourusername/hbf-agent/internal/servicemesh"
	"github.com/yourusername/hbf-agent/internal/api"
)

// Agent represents the main HBF agent
type Agent struct {
	config      *config.Config
	log         *logrus.Logger
	
	firewall    *firewall.Manager
	serviceMesh *servicemesh.Manager
	healthCheck *health.Checker
	metrics     *metrics.Manager
	apiServer   *api.Server
	
	mu          sync.RWMutex
	running     bool
	stopChan    chan struct{}
}

// New creates a new agent instance
func New(cfg *config.Config, log *logrus.Logger) (*Agent, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	
	if log == nil {
		log = logrus.New()
	}
	
	agent := &Agent{
		config:   cfg,
		log:      log,
		stopChan: make(chan struct{}),
	}
	
	// Initialize firewall manager
	fwManager, err := firewall.NewManager(cfg.Firewall, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create firewall manager: %w", err)
	}
	agent.firewall = fwManager
	
	// Initialize service mesh manager if enabled
	if cfg.ServiceMesh.Enabled {
		smManager, err := servicemesh.NewManager(cfg.ServiceMesh, log)
		if err != nil {
			return nil, fmt.Errorf("failed to create service mesh manager: %w", err)
		}
		agent.serviceMesh = smManager
	}
	
	// Initialize health checker
	healthChecker := health.NewChecker(log)
	agent.healthCheck = healthChecker
	
	// Initialize metrics manager
	metricsManager := metrics.NewManager(cfg.Monitoring, log)
	agent.metrics = metricsManager
	
	// Initialize API server
	apiServer, err := api.NewServer(cfg, agent.firewall, agent.serviceMesh, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create API server: %w", err)
	}
	agent.apiServer = apiServer
	
	return agent, nil
}

// Start starts the agent and all its components
func (a *Agent) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("agent is already running")
	}
	a.running = true
	a.mu.Unlock()
	
	a.log.Info("Starting agent components...")
	
	// Start firewall manager
	if err := a.firewall.Start(ctx); err != nil {
		return fmt.Errorf("failed to start firewall manager: %w", err)
	}
	a.log.Info("Firewall manager started")
	
	// Start service mesh manager if enabled
	if a.serviceMesh != nil {
		if err := a.serviceMesh.Start(ctx); err != nil {
			return fmt.Errorf("failed to start service mesh manager: %w", err)
		}
		a.log.Info("Service mesh manager started")
	}
	
	// Start health checker
	if err := a.healthCheck.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health checker: %w", err)
	}
	a.log.Info("Health checker started")
	
	// Start metrics manager
	if err := a.metrics.Start(ctx); err != nil {
		return fmt.Errorf("failed to start metrics manager: %w", err)
	}
	a.log.Info("Metrics manager started")
	
	// Start API server
	go func() {
		if err := a.apiServer.Start(); err != nil {
			a.log.Errorf("API server error: %v", err)
		}
	}()
	a.log.Info("API server started")
	
	// Wait for context cancellation
	<-ctx.Done()
	
	return nil
}

// Stop stops the agent and all its components
func (a *Agent) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return fmt.Errorf("agent is not running")
	}
	a.running = false
	a.mu.Unlock()
	
	a.log.Info("Stopping agent components...")
	
	var errors []error
	
	// Stop API server
	if err := a.apiServer.Stop(); err != nil {
		errors = append(errors, fmt.Errorf("failed to stop API server: %w", err))
	}
	
	// Stop metrics manager
	if err := a.metrics.Stop(); err != nil {
		errors = append(errors, fmt.Errorf("failed to stop metrics manager: %w", err))
	}
	
	// Stop health checker
	if err := a.healthCheck.Stop(); err != nil {
		errors = append(errors, fmt.Errorf("failed to stop health checker: %w", err))
	}
	
	// Stop service mesh manager
	if a.serviceMesh != nil {
		if err := a.serviceMesh.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop service mesh manager: %w", err))
		}
	}
	
	// Stop firewall manager
	if err := a.firewall.Stop(); err != nil {
		errors = append(errors, fmt.Errorf("failed to stop firewall manager: %w", err))
	}
	
	close(a.stopChan)
	
	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}
	
	return nil
}

// IsRunning returns whether the agent is currently running
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// GetFirewallManager returns the firewall manager
func (a *Agent) GetFirewallManager() *firewall.Manager {
	return a.firewall
}

// GetServiceMeshManager returns the service mesh manager
func (a *Agent) GetServiceMeshManager() *servicemesh.Manager {
	return a.serviceMesh
}

// GetHealthChecker returns the health checker
func (a *Agent) GetHealthChecker() *health.Checker {
	return a.healthCheck
}

// GetMetricsManager returns the metrics manager
func (a *Agent) GetMetricsManager() *metrics.Manager {
	return a.metrics
}
