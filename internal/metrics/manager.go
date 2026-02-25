package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/hbf-agent/internal/config"
)

// Manager manages metrics collection and exposure
type Manager struct {
	config   config.MonitoringConfig
	log      *logrus.Logger
	registry *prometheus.Registry
	server   *http.Server
	metrics  *Metrics
	mu       sync.RWMutex
	running  bool
}

// Metrics contains all Prometheus metrics
type Metrics struct {
	// Firewall metrics
	FirewallRulesTotal    prometheus.Gauge
	FirewallRuleAdditions prometheus.Counter
	FirewallRuleDeletions prometheus.Counter
	
	// Service mesh metrics
	ServicesRegistered    prometheus.Gauge
	ServiceHealthStatus   *prometheus.GaugeVec
	ServiceRequests       *prometheus.CounterVec
	ServiceRequestDuration *prometheus.HistogramVec
	
	// Traffic metrics
	TrafficBytesTotal     *prometheus.CounterVec
	ConnectionsActive     prometheus.Gauge
	ConnectionsTotal      prometheus.Counter
	
	// Health check metrics
	HealthChecksTotal     *prometheus.CounterVec
	HealthCheckDuration   *prometheus.HistogramVec
	
	// Agent metrics
	AgentUptime           prometheus.Counter
	AgentErrors           *prometheus.CounterVec
}

// NewManager creates a new metrics manager
func NewManager(cfg config.MonitoringConfig, log *logrus.Logger) *Manager {
	registry := prometheus.NewRegistry()
	
	metrics := &Metrics{
		// Firewall metrics
		FirewallRulesTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "hbf_firewall_rules_total",
			Help: "Total number of firewall rules",
		}),
		FirewallRuleAdditions: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "hbf_firewall_rule_additions_total",
			Help: "Total number of firewall rule additions",
		}),
		FirewallRuleDeletions: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "hbf_firewall_rule_deletions_total",
			Help: "Total number of firewall rule deletions",
		}),
		
		// Service mesh metrics
		ServicesRegistered: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "hbf_services_registered",
			Help: "Number of registered services",
		}),
		ServiceHealthStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hbf_service_health_status",
				Help: "Service health status (1=healthy, 0=unhealthy)",
			},
			[]string{"service_name", "service_id"},
		),
		ServiceRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "hbf_service_requests_total",
				Help: "Total number of service requests",
			},
			[]string{"service_name", "method", "status"},
		),
		ServiceRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "hbf_service_request_duration_seconds",
				Help:    "Service request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service_name", "method"},
		),
		
		// Traffic metrics
		TrafficBytesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "hbf_traffic_bytes_total",
				Help: "Total traffic bytes",
			},
			[]string{"direction"}, // inbound, outbound
		),
		ConnectionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "hbf_connections_active",
			Help: "Number of active connections",
		}),
		ConnectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "hbf_connections_total",
			Help: "Total number of connections",
		}),
		
		// Health check metrics
		HealthChecksTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "hbf_health_checks_total",
				Help: "Total number of health checks",
			},
			[]string{"check_id", "status"},
		),
		HealthCheckDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "hbf_health_check_duration_seconds",
				Help:    "Health check duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"check_id"},
		),
		
		// Agent metrics
		AgentUptime: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "hbf_agent_uptime_seconds",
			Help: "Agent uptime in seconds",
		}),
		AgentErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "hbf_agent_errors_total",
				Help: "Total number of agent errors",
			},
			[]string{"component", "error_type"},
		),
	}
	
	// Register all metrics
	registry.MustRegister(
		metrics.FirewallRulesTotal,
		metrics.FirewallRuleAdditions,
		metrics.FirewallRuleDeletions,
		metrics.ServicesRegistered,
		metrics.ServiceHealthStatus,
		metrics.ServiceRequests,
		metrics.ServiceRequestDuration,
		metrics.TrafficBytesTotal,
		metrics.ConnectionsActive,
		metrics.ConnectionsTotal,
		metrics.HealthChecksTotal,
		metrics.HealthCheckDuration,
		metrics.AgentUptime,
		metrics.AgentErrors,
	)
	
	return &Manager{
		config:   cfg,
		log:      log,
		registry: registry,
		metrics:  metrics,
	}
}

// Start starts the metrics manager
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("metrics manager is already running")
	}
	m.running = true
	m.mu.Unlock()
	
	if !m.config.Enabled {
		m.log.Info("Metrics collection is disabled")
		return nil
	}
	
	m.log.Info("Starting metrics manager...")
	
	// Create HTTP server for metrics
	mux := http.NewServeMux()
	mux.Handle(m.config.MetricsPath, promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
	
	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.config.MetricsPort),
		Handler: mux,
	}
	
	// Start server in goroutine
	go func() {
		m.log.Infof("Metrics server listening on :%d%s", m.config.MetricsPort, m.config.MetricsPath)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.log.Errorf("Metrics server error: %v", err)
		}
	}()
	
	return nil
}

// Stop stops the metrics manager
func (m *Manager) Stop() error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return fmt.Errorf("metrics manager is not running")
	}
	m.running = false
	m.mu.Unlock()
	
	if m.server != nil {
		if err := m.server.Close(); err != nil {
			return fmt.Errorf("failed to stop metrics server: %w", err)
		}
	}
	
	m.log.Info("Metrics manager stopped")
	return nil
}

// GetMetrics returns the metrics instance
func (m *Manager) GetMetrics() *Metrics {
	return m.metrics
}

// RecordFirewallRuleAdd records a firewall rule addition
func (m *Manager) RecordFirewallRuleAdd() {
	m.metrics.FirewallRuleAdditions.Inc()
	m.metrics.FirewallRulesTotal.Inc()
}

// RecordFirewallRuleDelete records a firewall rule deletion
func (m *Manager) RecordFirewallRuleDelete() {
	m.metrics.FirewallRuleDeletions.Inc()
	m.metrics.FirewallRulesTotal.Dec()
}

// SetFirewallRulesTotal sets the total number of firewall rules
func (m *Manager) SetFirewallRulesTotal(count float64) {
	m.metrics.FirewallRulesTotal.Set(count)
}

// SetServicesRegistered sets the number of registered services
func (m *Manager) SetServicesRegistered(count float64) {
	m.metrics.ServicesRegistered.Set(count)
}

// SetServiceHealthStatus sets the health status of a service
func (m *Manager) SetServiceHealthStatus(serviceName, serviceID string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.metrics.ServiceHealthStatus.WithLabelValues(serviceName, serviceID).Set(value)
}

// RecordServiceRequest records a service request
func (m *Manager) RecordServiceRequest(serviceName, method, status string, duration float64) {
	m.metrics.ServiceRequests.WithLabelValues(serviceName, method, status).Inc()
	m.metrics.ServiceRequestDuration.WithLabelValues(serviceName, method).Observe(duration)
}

// RecordTrafficBytes records traffic bytes
func (m *Manager) RecordTrafficBytes(direction string, bytes float64) {
	m.metrics.TrafficBytesTotal.WithLabelValues(direction).Add(bytes)
}

// SetConnectionsActive sets the number of active connections
func (m *Manager) SetConnectionsActive(count float64) {
	m.metrics.ConnectionsActive.Set(count)
}

// RecordConnection records a new connection
func (m *Manager) RecordConnection() {
	m.metrics.ConnectionsTotal.Inc()
}

// RecordHealthCheck records a health check
func (m *Manager) RecordHealthCheck(checkID, status string, duration float64) {
	m.metrics.HealthChecksTotal.WithLabelValues(checkID, status).Inc()
	m.metrics.HealthCheckDuration.WithLabelValues(checkID).Observe(duration)
}

// RecordError records an error
func (m *Manager) RecordError(component, errorType string) {
	m.metrics.AgentErrors.WithLabelValues(component, errorType).Inc()
}

// IncrementUptime increments the agent uptime
func (m *Manager) IncrementUptime() {
	m.metrics.AgentUptime.Inc()
}
