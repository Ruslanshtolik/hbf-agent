package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete agent configuration
type Config struct {
	Agent       AgentConfig       `mapstructure:"agent"`
	Firewall    FirewallConfig    `mapstructure:"firewall"`
	ServiceMesh ServiceMeshConfig `mapstructure:"service_mesh"`
	Security    SecurityConfig    `mapstructure:"security"`
	Monitoring  MonitoringConfig  `mapstructure:"monitoring"`
	Log         LogConfig         `mapstructure:"log"`
}

// AgentConfig contains agent-level configuration
type AgentConfig struct {
	NodeID     string `mapstructure:"node_id"`
	Datacenter string `mapstructure:"datacenter"`
	Region     string `mapstructure:"region"`
	BindAddr   string `mapstructure:"bind_addr"`
	APIPort    int    `mapstructure:"api_port"`
}

// FirewallConfig contains firewall configuration
type FirewallConfig struct {
	Backend       string        `mapstructure:"backend"` // iptables or nftables
	DefaultPolicy string        `mapstructure:"default_policy"`
	EnableIPv6    bool          `mapstructure:"enable_ipv6"`
	SyncInterval  time.Duration `mapstructure:"sync_interval"`
	Rules         []FirewallRule `mapstructure:"rules"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	Chain    string `mapstructure:"chain"`
	Protocol string `mapstructure:"protocol"`
	Source   string `mapstructure:"source"`
	Dest     string `mapstructure:"dest"`
	SPort    string `mapstructure:"sport"`
	DPort    string `mapstructure:"dport"`
	Action   string `mapstructure:"action"`
	Comment  string `mapstructure:"comment"`
}

// ServiceMeshConfig contains service mesh configuration
type ServiceMeshConfig struct {
	Enabled     bool              `mapstructure:"enabled"`
	BindAddress string            `mapstructure:"bind_address"`
	ProxyPort   int               `mapstructure:"proxy_port"`
	AdminPort   int               `mapstructure:"admin_port"`
	Discovery   DiscoveryConfig   `mapstructure:"discovery"`
	LoadBalance LoadBalanceConfig `mapstructure:"load_balance"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

// DiscoveryConfig contains service discovery configuration
type DiscoveryConfig struct {
	Backend  string        `mapstructure:"backend"` // consul, etcd, dns, static
	Address  string        `mapstructure:"address"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Interval time.Duration `mapstructure:"interval"`
}

// LoadBalanceConfig contains load balancing configuration
type LoadBalanceConfig struct {
	Strategy string `mapstructure:"strategy"` // round_robin, least_conn, random, weighted
}

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	Threshold         int           `mapstructure:"threshold"`
	Timeout           time.Duration `mapstructure:"timeout"`
	HalfOpenRequests  int           `mapstructure:"half_open_requests"`
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	MTLS       MTLSConfig       `mapstructure:"mtls"`
	Auth       AuthConfig       `mapstructure:"auth"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
}

// MTLSConfig contains mTLS configuration
type MTLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	CAFile   string `mapstructure:"ca_file"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Type    string   `mapstructure:"type"` // token, jwt, mtls
	Tokens  []string `mapstructure:"tokens"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled"`
	RPS     int  `mapstructure:"rps"`
	Burst   int  `mapstructure:"burst"`
}

// MonitoringConfig contains monitoring configuration
type MonitoringConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	MetricsPort    int    `mapstructure:"metrics_port"`
	MetricsPath    string `mapstructure:"metrics_path"`
	HealthPort     int    `mapstructure:"health_port"`
	HealthPath     string `mapstructure:"health_path"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// Load loads the configuration from viper
func Load() (*Config, error) {
	cfg := &Config{}
	
	// Set defaults
	setDefaults()
	
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return cfg, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Agent defaults
	viper.SetDefault("agent.node_id", "node-01")
	viper.SetDefault("agent.datacenter", "dc1")
	viper.SetDefault("agent.region", "default")
	viper.SetDefault("agent.bind_addr", "0.0.0.0")
	viper.SetDefault("agent.api_port", 9090)
	
	// Firewall defaults
	viper.SetDefault("firewall.backend", "iptables")
	viper.SetDefault("firewall.default_policy", "deny")
	viper.SetDefault("firewall.enable_ipv6", true)
	viper.SetDefault("firewall.sync_interval", "30s")
	
	// Service mesh defaults
	viper.SetDefault("service_mesh.enabled", true)
	viper.SetDefault("service_mesh.bind_address", "0.0.0.0")
	viper.SetDefault("service_mesh.proxy_port", 8080)
	viper.SetDefault("service_mesh.admin_port", 8081)
	viper.SetDefault("service_mesh.discovery.backend", "consul")
	viper.SetDefault("service_mesh.discovery.address", "localhost:8500")
	viper.SetDefault("service_mesh.discovery.timeout", "5s")
	viper.SetDefault("service_mesh.discovery.interval", "10s")
	viper.SetDefault("service_mesh.load_balance.strategy", "round_robin")
	viper.SetDefault("service_mesh.circuit_breaker.enabled", true)
	viper.SetDefault("service_mesh.circuit_breaker.threshold", 5)
	viper.SetDefault("service_mesh.circuit_breaker.timeout", "30s")
	viper.SetDefault("service_mesh.circuit_breaker.half_open_requests", 3)
	
	// Security defaults
	viper.SetDefault("security.mtls.enabled", false)
	viper.SetDefault("security.auth.enabled", false)
	viper.SetDefault("security.rate_limit.enabled", true)
	viper.SetDefault("security.rate_limit.rps", 1000)
	viper.SetDefault("security.rate_limit.burst", 2000)
	
	// Monitoring defaults
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.metrics_port", 9091)
	viper.SetDefault("monitoring.metrics_path", "/metrics")
	viper.SetDefault("monitoring.health_port", 9092)
	viper.SetDefault("monitoring.health_path", "/health")
	
	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "text")
	viper.SetDefault("log.output", "stdout")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Agent.NodeID == "" {
		return fmt.Errorf("agent.node_id is required")
	}
	
	if c.Firewall.Backend != "iptables" && c.Firewall.Backend != "nftables" {
		return fmt.Errorf("firewall.backend must be 'iptables' or 'nftables'")
	}
	
	if c.ServiceMesh.Enabled {
		if c.ServiceMesh.Discovery.Backend == "" {
			return fmt.Errorf("service_mesh.discovery.backend is required when service mesh is enabled")
		}
		
		validBackends := map[string]bool{
			"consul": true,
			"etcd":   true,
			"dns":    true,
			"static": true,
		}
		
		if !validBackends[c.ServiceMesh.Discovery.Backend] {
			return fmt.Errorf("invalid service_mesh.discovery.backend: %s", c.ServiceMesh.Discovery.Backend)
		}
	}
	
	if c.Security.MTLS.Enabled {
		if c.Security.MTLS.CertFile == "" || c.Security.MTLS.KeyFile == "" || c.Security.MTLS.CAFile == "" {
			return fmt.Errorf("mTLS requires cert_file, key_file, and ca_file")
		}
	}
	
	return nil
}
