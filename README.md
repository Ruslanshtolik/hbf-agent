# HBF Agent with Service Mesh

A high-performance Host-Based Firewall (HBF) agent with integrated service mesh capabilities for Linux hosts.

## Features

- **Host-Based Firewall**: Dynamic iptables/nftables rule management
- **Service Mesh Integration**: Service discovery, registration, and routing
- **Health Checking**: Automatic health monitoring and failover
- **Traffic Management**: Load balancing, circuit breaking, and rate limiting
- **Observability**: Metrics, logging, and distributed tracing
- **Zero-Trust Security**: mTLS, authentication, and authorization
- **Dynamic Configuration**: Hot-reload without service interruption

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     HBF Agent                           │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Firewall   │  │Service Mesh  │  │   Health     │ │
│  │   Manager    │  │   Manager    │  │   Checker    │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Service    │  │   Traffic    │  │  Metrics &   │ │
│  │  Discovery   │  │   Manager    │  │   Logging    │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
├─────────────────────────────────────────────────────────┤
│              iptables/nftables Interface                │
└─────────────────────────────────────────────────────────┘
```

## Components

### 1. Firewall Manager
- Dynamic rule management
- Support for iptables and nftables
- Rule validation and rollback
- Connection tracking

### 2. Service Mesh Manager
- Service registration and discovery
- Sidecar proxy functionality
- Traffic routing and load balancing
- Circuit breaking and retries

### 3. Health Checker
- Active and passive health checks
- Automatic service removal on failure
- Configurable check intervals
- Multiple check types (HTTP, TCP, gRPC)

### 4. Service Discovery
- Consul integration
- etcd support
- DNS-based discovery
- Static configuration

### 5. Traffic Manager
- Layer 4 and Layer 7 load balancing
- Rate limiting
- Traffic splitting
- Canary deployments

## Installation

### Prerequisites

- Linux kernel 4.x or higher
- Go 1.21+ (for building from source)
- iptables or nftables
- Root or CAP_NET_ADMIN privileges

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/hbf-agent.git
cd hbf-agent

# Build
make build

# Install
sudo make install
```

### Using Pre-built Binaries

```bash
# Download latest release
wget https://github.com/yourusername/hbf-agent/releases/latest/download/hbf-agent-linux-amd64.tar.gz

# Extract
tar -xzf hbf-agent-linux-amd64.tar.gz

# Install
sudo ./install.sh
```

## Configuration

Create a configuration file at `/etc/hbf-agent/config.yaml`:

```yaml
# See config/config.example.yaml for full configuration options
agent:
  node_id: "node-01"
  datacenter: "dc1"
  log_level: "info"

firewall:
  backend: "iptables"  # or "nftables"
  default_policy: "deny"
  
service_mesh:
  enabled: true
  bind_address: "0.0.0.0:8080"
  discovery:
    backend: "consul"
    address: "localhost:8500"
```

## Usage

### Start the Agent

```bash
# Using systemd
sudo systemctl start hbf-agent
sudo systemctl enable hbf-agent

# Manual start
sudo hbf-agent --config /etc/hbf-agent/config.yaml
```

### Register a Service

```bash
# Using CLI
hbf-agent service register \
  --name web-service \
  --port 8080 \
  --health-check http://localhost:8080/health

# Using API
curl -X POST http://localhost:9090/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-service",
    "port": 8080,
    "health_check": {
      "type": "http",
      "endpoint": "http://localhost:8080/health",
      "interval": "10s"
    }
  }'
```

### Add Firewall Rules

```bash
# Allow incoming traffic on port 80
hbf-agent firewall add-rule \
  --chain INPUT \
  --protocol tcp \
  --dport 80 \
  --action ACCEPT

# Block specific IP
hbf-agent firewall add-rule \
  --chain INPUT \
  --source 192.168.1.100 \
  --action DROP
```

## API Reference

The agent exposes a REST API on port 9090 (configurable):

- `GET /api/v1/health` - Agent health status
- `GET /api/v1/services` - List registered services
- `POST /api/v1/services` - Register a service
- `DELETE /api/v1/services/{id}` - Deregister a service
- `GET /api/v1/firewall/rules` - List firewall rules
- `POST /api/v1/firewall/rules` - Add firewall rule
- `DELETE /api/v1/firewall/rules/{id}` - Remove firewall rule
- `GET /api/v1/metrics` - Prometheus metrics

## Monitoring

### Metrics

The agent exposes Prometheus metrics at `/metrics`:

- `hbf_firewall_rules_total` - Total number of firewall rules
- `hbf_services_registered` - Number of registered services
- `hbf_service_health_status` - Service health status
- `hbf_traffic_bytes_total` - Total traffic bytes
- `hbf_connections_active` - Active connections

### Logging

Logs are written to:
- stdout/stderr (when running in foreground)
- `/var/log/hbf-agent/agent.log` (systemd)
- Syslog (configurable)

## Security

### mTLS Configuration

```yaml
security:
  mtls:
    enabled: true
    cert_file: "/etc/hbf-agent/certs/agent.crt"
    key_file: "/etc/hbf-agent/certs/agent.key"
    ca_file: "/etc/hbf-agent/certs/ca.crt"
```

### Authentication

The agent supports multiple authentication methods:
- API tokens
- mTLS client certificates
- JWT tokens

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Running Locally

```bash
make run
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## Support

- Documentation: https://docs.hbf-agent.io
- Issues: https://github.com/yourusername/hbf-agent/issues
- Slack: https://hbf-agent.slack.com
