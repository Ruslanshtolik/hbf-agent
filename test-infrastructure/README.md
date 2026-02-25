# HBF Agent Test Infrastructure

This directory contains complete infrastructure-as-code for testing the HBF Agent in various environments.

## Overview

The test infrastructure provides multiple deployment options:
- **Docker Compose**: Quick local testing with containers
- **Vagrant**: Full VM-based testing with VirtualBox
- **Terraform**: Cloud deployment on AWS
- **Kubernetes**: Container orchestration testing

## Quick Start

### Option 1: Docker Compose (Fastest)

```bash
cd docker-compose
docker-compose up -d
```

Access services:
- Consul UI: http://localhost:8500
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000
- HBF Node 1 API: http://localhost:9090

### Option 2: Vagrant (Full VMs)

```bash
cd vagrant
vagrant up
```

This will create:
- 1 Consul server
- 1 Prometheus server
- 1 Grafana server
- 3 HBF application nodes
- 3 HBF database nodes
- 1 HAProxy load balancer

### Option 3: Terraform (AWS Cloud)

```bash
cd terraform
terraform init
terraform plan -var="key_name=your-ssh-key"
terraform apply -var="key_name=your-ssh-key"
```

## Architecture

```
Management Network (10.0.0.0/24)
├── Consul (10.0.0.10)
├── Prometheus (10.0.0.11)
└── Grafana (10.0.0.12)

Application Network (10.1.0.0/24)
├── HBF Node 1 (10.1.0.101)
├── HBF Node 2 (10.1.0.102)
└── HBF Node 3 (10.1.0.103)

Database Network (10.2.0.0/24)
├── DB Node 1 (10.2.0.201)
├── DB Node 2 (10.2.0.202)
└── DB Node 3 (10.2.0.203)

External Network (10.3.0.0/24)
└── Load Balancer (10.3.0.100)
```

## Testing Scenarios

### 1. Basic Functionality Test

```bash
# Register a service
curl -X POST http://localhost:9090/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-service",
    "port": 8080,
    "health_check": {
      "type": "http",
      "endpoint": "http://localhost:8080/health",
      "interval": "10s"
    }
  }'

# Add firewall rule
curl -X POST http://localhost:9090/api/v1/firewall/rules \
  -H "Content-Type: application/json" \
  -d '{
    "chain": "INPUT",
    "protocol": "tcp",
    "dport": "8080",
    "action": "ACCEPT"
  }'

# Check metrics
curl http://localhost:9091/metrics
```

### 2. High Availability Test

```bash
# Check service discovery across nodes
for i in {1..3}; do
  curl http://localhost:$((9090 + (i-1)*100))/api/v1/services
done

# Simulate node failure
docker-compose stop hbf-node-1

# Verify failover
curl http://localhost:9090/api/v1/services
```

### 3. Load Balancing Test

```bash
# Generate traffic
for i in {1..100}; do
  curl http://localhost:80/
done

# Check distribution
curl http://localhost:8404/haproxy?stats
```

### 4. Security Test

```bash
# Test firewall rules
# From test-client container
docker exec test-client sh -c "
  # Should succeed
  curl http://10.1.0.101:8080
  
  # Should fail (blocked by firewall)
  curl http://10.2.0.201:5432
"
```

### 5. Performance Test

```bash
# Install Apache Bench
apt-get install apache2-utils

# Run load test
ab -n 10000 -c 100 http://localhost:80/

# Monitor metrics
watch -n 1 'curl -s http://localhost:9091/metrics | grep hbf_'
```

## Monitoring

### Prometheus Queries

```promql
# Service health
hbf_service_health_status

# Firewall rules count
hbf_firewall_rules_total

# Request rate
rate(hbf_service_requests_total[5m])

# Request latency
histogram_quantile(0.95, rate(hbf_service_request_duration_seconds_bucket[5m]))
```

### Grafana Dashboards

Import the pre-configured dashboards:
1. HBF Agent Overview
2. Service Mesh Metrics
3. Network Traffic
4. System Resources

## Troubleshooting

### Docker Compose Issues

```bash
# View logs
docker-compose logs -f hbf-node-1

# Restart service
docker-compose restart hbf-node-1

# Rebuild
docker-compose build --no-cache
docker-compose up -d
```

### Vagrant Issues

```bash
# Check VM status
vagrant status

# SSH into VM
vagrant ssh hbf-node-1

# Reload VM
vagrant reload hbf-node-1

# Destroy and recreate
vagrant destroy -f
vagrant up
```

### Terraform Issues

```bash
# Check state
terraform show

# Refresh state
terraform refresh

# Destroy resources
terraform destroy -auto-approve
```

## Cleanup

### Docker Compose

```bash
docker-compose down -v
```

### Vagrant

```bash
vagrant destroy -f
```

### Terraform

```bash
terraform destroy -var="key_name=your-ssh-key"
```

## Advanced Configuration

### Custom Network Topology

Edit `docker-compose.yml` or `Vagrantfile` to modify:
- IP addresses
- Network segments
- Resource allocation
- Service configuration

### Adding More Nodes

**Docker Compose:**
```yaml
hbf-node-4:
  build: ../../
  networks:
    application:
      ipv4_address: 10.1.0.104
```

**Vagrant:**
```ruby
config.vm.define "hbf-node-4" do |node|
  node.vm.network "private_network", ip: "10.1.0.104"
end
```

### Custom Firewall Rules

Edit node configuration files in `configs/` directory.

## Performance Tuning

### Resource Allocation

**Docker:**
```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 4G
```

**Vagrant:**
```ruby
vb.memory = "8192"
vb.cpus = 4
```

### Network Optimization

- Enable jumbo frames
- Adjust MTU settings
- Configure TCP tuning parameters

## Integration Tests

Run the full test suite:

```bash
cd tests
./run-all-tests.sh
```

Individual test suites:
```bash
./test-service-discovery.sh
./test-firewall-rules.sh
./test-load-balancing.sh
./test-health-checks.sh
./test-failover.sh
```

## CI/CD Integration

### Jenkins Pipeline

```groovy
pipeline {
  stages {
    stage('Deploy') {
      steps {
        sh 'cd test-infrastructure/docker-compose && docker-compose up -d'
      }
    }
    stage('Test') {
      steps {
        sh 'cd tests && ./run-all-tests.sh'
      }
    }
    stage('Cleanup') {
      steps {
        sh 'cd test-infrastructure/docker-compose && docker-compose down -v'
      }
    }
  }
}
```

### GitHub Actions

```yaml
name: Test Infrastructure
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Deploy
        run: cd test-infrastructure/docker-compose && docker-compose up -d
      - name: Test
        run: cd tests && ./run-all-tests.sh
      - name: Cleanup
        run: cd test-infrastructure/docker-compose && docker-compose down -v
```

## Documentation

- [Testing Architecture](TESTING_ARCHITECTURE.md) - Detailed architecture
- [Test Scenarios](scenarios/) - Specific test scenarios
- [Troubleshooting Guide](TROUBLESHOOTING.md) - Common issues

## Support

For issues or questions:
- GitHub Issues: https://github.com/yourusername/hbf-agent/issues
- Documentation: https://docs.hbf-agent.io
