# Quick Start Guide

This guide will help you get HBF Agent up and running quickly on a Linux host.

## Prerequisites

- Linux kernel 4.x or higher
- Root or sudo access
- iptables or nftables installed
- Go 1.21+ (for building from source)

## Installation

### Option 1: From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/hbf-agent.git
cd hbf-agent

# Build the binary
make build

# Install
sudo make install
```

### Option 2: Using Install Script

```bash
# Download and extract release
wget https://github.com/yourusername/hbf-agent/releases/latest/download/hbf-agent-linux-amd64.tar.gz
tar -xzf hbf-agent-linux-amd64.tar.gz
cd hbf-agent

# Run install script
sudo bash scripts/install.sh
```

### Option 3: Using Docker

```bash
# Build Docker image
docker build -t hbf-agent:latest .

# Run container (requires privileged mode for firewall access)
docker run -d \
  --name hbf-agent \
  --privileged \
  --network host \
  -v /etc/hbf-agent:/etc/hbf-agent \
  hbf-agent:latest
```

## Configuration

Edit the configuration file at `/etc/hbf-agent/config.yaml`:

```yaml
agent:
  node_id: "my-node"
  datacenter: "dc1"
  api_port: 9090

firewall:
  backend: "iptables"
  default_policy: "deny"

service_mesh:
  enabled: true
  discovery:
    backend: "consul"
    address: "localhost:8500"
```

## Starting the Agent

### Using systemd

```bash
# Start the service
sudo systemctl start hbf-agent

# Enable on boot
sudo systemctl enable hbf-agent

# Check status
sudo systemctl status hbf-agent

# View logs
sudo journalctl -u hbf-agent -f
```

### Manual Start

```bash
# Run in foreground
sudo hbf-agent --config /etc/hbf-agent/config.yaml

# Run in background
sudo hbf-agent --config /etc/hbf-agent/config.yaml &
```

## Basic Usage

### Check Agent Health

```bash
curl http://localhost:9090/api/v1/health
```

Expected output:
```json
{
  "status": "healthy",
  "agent": {
    "node_id": "my-node",
    "datacenter": "dc1"
  }
}
```

### Register a Service

```bash
curl -X POST http://localhost:9090/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-service",
    "address": "127.0.0.1",
    "port": 8080,
    "tags": ["web", "api"],
    "health_check": {
      "type": "http",
      "endpoint": "http://localhost:8080/health",
      "interval": "10s",
      "timeout": "5s"
    }
  }'
```

### List Services

```bash
curl http://localhost:9090/api/v1/services
```

### Add Firewall Rule

```bash
curl -X POST http://localhost:9090/api/v1/firewall/rules \
  -H "Content-Type: application/json" \
  -d '{
    "chain": "INPUT",
    "protocol": "tcp",
    "dport": "8080",
    "action": "ACCEPT",
    "comment": "Allow web service"
  }'
```

### List Firewall Rules

```bash
curl http://localhost:9090/api/v1/firewall/rules
```

### View Metrics

```bash
curl http://localhost:9091/metrics
```

## Common Tasks

### Allow SSH Access

```bash
curl -X POST http://localhost:9090/api/v1/firewall/rules \
  -H "Content-Type: application/json" \
  -d '{
    "chain": "INPUT",
    "protocol": "tcp",
    "dport": "22",
    "action": "ACCEPT",
    "comment": "Allow SSH"
  }'
```

### Block Specific IP

```bash
curl -X POST http://localhost:9090/api/v1/firewall/rules \
  -H "Content-Type: application/json" \
  -d '{
    "chain": "INPUT",
    "source": "192.168.1.100",
    "action": "DROP",
    "comment": "Block malicious IP"
  }'
```

### Register Multiple Services

```bash
# Register database service
curl -X POST http://localhost:9090/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "database",
    "address": "127.0.0.1",
    "port": 5432,
    "tags": ["db", "postgres"]
  }'

# Register cache service
curl -X POST http://localhost:9090/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cache",
    "address": "127.0.0.1",
    "port": 6379,
    "tags": ["cache", "redis"]
  }'
```

## Monitoring

### Check Prometheus Metrics

```bash
# View all metrics
curl http://localhost:9091/metrics

# Filter specific metrics
curl http://localhost:9091/metrics | grep hbf_firewall

# Check service health
curl http://localhost:9091/metrics | grep hbf_service_health
```

### View Logs

```bash
# Using journalctl (systemd)
sudo journalctl -u hbf-agent -f

# Using log file
sudo tail -f /var/log/hbf-agent/agent.log
```

## Troubleshooting

### Agent Won't Start

1. Check if iptables is installed:
   ```bash
   which iptables
   ```

2. Verify configuration:
   ```bash
   hbf-agent --config /etc/hbf-agent/config.yaml --help
   ```

3. Check permissions:
   ```bash
   ls -la /etc/hbf-agent/config.yaml
   ```

### Firewall Rules Not Applied

1. Check backend status:
   ```bash
   sudo iptables -L -n -v
   ```

2. View agent logs:
   ```bash
   sudo journalctl -u hbf-agent -n 50
   ```

3. Verify rule syntax:
   ```bash
   curl http://localhost:9090/api/v1/firewall/rules
   ```

### Service Discovery Issues

1. Check discovery backend:
   ```bash
   # For Consul
   curl http://localhost:8500/v1/agent/self
   
   # For etcd
   etcdctl endpoint health
   ```

2. Verify network connectivity:
   ```bash
   telnet localhost 8500
   ```

3. Check service mesh logs:
   ```bash
   sudo journalctl -u hbf-agent | grep "service mesh"
   ```

## Next Steps

- Read the [Architecture Documentation](ARCHITECTURE.md)
- Explore [Advanced Configuration](CONFIGURATION.md)
- Set up [Monitoring and Alerting](MONITORING.md)
- Learn about [Security Best Practices](SECURITY.md)

## Getting Help

- GitHub Issues: https://github.com/yourusername/hbf-agent/issues
- Documentation: https://docs.hbf-agent.io
- Community Slack: https://hbf-agent.slack.com
