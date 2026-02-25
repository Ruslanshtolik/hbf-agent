# HBF Agent Testing Infrastructure Architecture

## Overview

This document describes a complete virtual infrastructure for testing the HBF Agent in various scenarios, including service mesh functionality, firewall rules, and high-availability configurations.

## Infrastructure Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Testing Infrastructure                           │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                    Management Network                            │  │
│  │                      10.0.0.0/24                                 │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │   Consul    │  │  Prometheus │  │   Grafana   │  │   Jenkins   │  │
│  │  10.0.0.10  │  │  10.0.0.11  │  │  10.0.0.12  │  │  10.0.0.13  │  │
│  │   (8500)    │  │   (9090)    │  │   (3000)    │  │   (8080)    │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                    Application Network                           │  │
│  │                      10.1.0.0/24                                 │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                    │
│  │  HBF Node 1 │  │  HBF Node 2 │  │  HBF Node 3 │                    │
│  │  10.1.0.101 │  │  10.1.0.102 │  │  10.1.0.103 │                    │
│  │             │  │             │  │             │                    │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │                    │
│  │ │HBF Agent│ │  │ │HBF Agent│ │  │ │HBF Agent│ │                    │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │                    │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │                    │
│  │ │Web Svc  │ │  │ │Web Svc  │ │  │ │Web Svc  │ │                    │
│  │ │  :8080  │ │  │ │  :8080  │ │  │ │  :8080  │ │                    │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │                    │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │                    │
│  │ │API Svc  │ │  │ │API Svc  │ │  │ │API Svc  │ │                    │
│  │ │  :8081  │ │  │ │  :8081  │ │  │ │  :8081  │ │                    │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │                    │
│  └─────────────┘  └─────────────┘  └─────────────┘                    │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                    Database Network                              │  │
│  │                      10.2.0.0/24                                 │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                    │
│  │  DB Node 1  │  │  DB Node 2  │  │  DB Node 3  │                    │
│  │  10.2.0.201 │  │  10.2.0.202 │  │  10.2.0.203 │                    │
│  │             │  │             │  │             │                    │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │                    │
│  │ │HBF Agent│ │  │ │HBF Agent│ │  │ │HBF Agent│ │                    │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │                    │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │                    │
│  │ │PostgreSQL│ │  │ │PostgreSQL│ │  │ │PostgreSQL│                    │
│  │ │  :5432  │ │  │ │  :5432  │ │  │ │  :5432  │ │                    │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │                    │
│  └─────────────┘  └─────────────┘  └─────────────┘                    │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                    External Network                              │  │
│  │                      10.3.0.0/24                                 │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐                                      │
│  │Load Balancer│  │  Bastion    │                                      │
│  │  10.3.0.100 │  │  10.3.0.10  │                                      │
│  │   (HAProxy) │  │   (SSH)     │                                      │
│  └─────────────┘  └─────────────┘                                      │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Management Network (10.0.0.0/24)

#### Consul Cluster (10.0.0.10)
- **Purpose**: Service discovery and configuration
- **Specs**: 2 vCPU, 4GB RAM, 20GB disk
- **Services**: Consul server, UI
- **Ports**: 8500 (HTTP), 8600 (DNS), 8300-8302 (Cluster)

#### Prometheus (10.0.0.11)
- **Purpose**: Metrics collection and alerting
- **Specs**: 2 vCPU, 4GB RAM, 50GB disk
- **Services**: Prometheus server, Alertmanager
- **Ports**: 9090 (Web UI), 9093 (Alertmanager)

#### Grafana (10.0.0.12)
- **Purpose**: Metrics visualization
- **Specs**: 1 vCPU, 2GB RAM, 20GB disk
- **Services**: Grafana server
- **Ports**: 3000 (Web UI)

#### Jenkins (10.0.0.13)
- **Purpose**: CI/CD and automated testing
- **Specs**: 2 vCPU, 4GB RAM, 50GB disk
- **Services**: Jenkins server
- **Ports**: 8080 (Web UI), 50000 (Agent)

### 2. Application Network (10.1.0.0/24)

#### HBF Node 1-3 (10.1.0.101-103)
- **Purpose**: Application servers with HBF agents
- **Specs**: 2 vCPU, 4GB RAM, 30GB disk each
- **OS**: Ubuntu 22.04 LTS
- **Services**:
  - HBF Agent (ports: 9090 API, 9091 Metrics, 8080 Proxy)
  - Web Service (port 8080)
  - API Service (port 8081)
  - Node Exporter (port 9100)

### 3. Database Network (10.2.0.0/24)

#### DB Node 1-3 (10.2.0.201-203)
- **Purpose**: Database servers with HBF agents
- **Specs**: 4 vCPU, 8GB RAM, 100GB disk each
- **OS**: Ubuntu 22.04 LTS
- **Services**:
  - HBF Agent (ports: 9090 API, 9091 Metrics)
  - PostgreSQL 15 (port 5432)
  - Postgres Exporter (port 9187)

### 4. External Network (10.3.0.0/24)

#### Load Balancer (10.3.0.100)
- **Purpose**: External traffic distribution
- **Specs**: 2 vCPU, 2GB RAM, 20GB disk
- **Services**: HAProxy
- **Ports**: 80 (HTTP), 443 (HTTPS)

#### Bastion Host (10.3.0.10)
- **Purpose**: Secure SSH access
- **Specs**: 1 vCPU, 2GB RAM, 20GB disk
- **Services**: OpenSSH server
- **Ports**: 22 (SSH)

## Network Configuration

### Network Segmentation

```yaml
networks:
  management:
    subnet: 10.0.0.0/24
    gateway: 10.0.0.1
    purpose: Infrastructure services
    
  application:
    subnet: 10.1.0.0/24
    gateway: 10.1.0.1
    purpose: Application workloads
    
  database:
    subnet: 10.2.0.0/24
    gateway: 10.2.0.1
    purpose: Database services
    
  external:
    subnet: 10.3.0.0/24
    gateway: 10.3.0.1
    purpose: External access
```

### Firewall Rules Matrix

| Source Network | Destination Network | Allowed Ports | Purpose |
|----------------|---------------------|---------------|---------|
| External | Application | 80, 443 | Web traffic |
| Application | Database | 5432 | Database access |
| Application | Management | 8500, 9090 | Service discovery, metrics |
| Management | All | 22, 9090-9092 | SSH, monitoring |
| All | Management | 8500 | Consul registration |
| Bastion | All | 22 | SSH access |

## Deployment Scenarios

### Scenario 1: Basic Deployment

**Purpose**: Test basic HBF agent functionality

**Setup**:
1. Deploy 1 HBF node with web service
2. Configure basic firewall rules
3. Register service with Consul
4. Verify health checks

**Test Cases**:
- Service registration
- Firewall rule application
- Health check functionality
- Metrics collection

### Scenario 2: High Availability

**Purpose**: Test HA and failover

**Setup**:
1. Deploy 3 HBF nodes with web services
2. Configure load balancing
3. Enable health-based routing
4. Simulate node failures

**Test Cases**:
- Service discovery across nodes
- Automatic failover
- Load distribution
- Circuit breaker activation

### Scenario 3: Multi-Tier Application

**Purpose**: Test complex service mesh

**Setup**:
1. Deploy application tier (3 nodes)
2. Deploy database tier (3 nodes)
3. Configure service-to-service communication
4. Implement network segmentation

**Test Cases**:
- Cross-tier communication
- Network isolation
- Service dependencies
- Traffic routing

### Scenario 4: Security Testing

**Purpose**: Test security features

**Setup**:
1. Enable mTLS between services
2. Configure authentication
3. Implement rate limiting
4. Set up network policies

**Test Cases**:
- mTLS handshake
- Authentication enforcement
- Rate limit effectiveness
- Unauthorized access blocking

### Scenario 5: Performance Testing

**Purpose**: Test under load

**Setup**:
1. Deploy full infrastructure
2. Configure monitoring
3. Generate traffic load
4. Monitor performance metrics

**Test Cases**:
- Throughput under load
- Latency measurements
- Resource utilization
- Scalability limits

## Infrastructure as Code

### Terraform Configuration

```hcl
# See test-infrastructure/terraform/main.tf
```

### Vagrant Configuration

```ruby
# See test-infrastructure/vagrant/Vagrantfile
```

### Docker Compose

```yaml
# See test-infrastructure/docker-compose/docker-compose.yml
```

## Monitoring Setup

### Prometheus Targets

```yaml
scrape_configs:
  - job_name: 'hbf-agents'
    static_configs:
      - targets:
        - '10.1.0.101:9091'
        - '10.1.0.102:9091'
        - '10.1.0.103:9091'
        - '10.2.0.201:9091'
        - '10.2.0.202:9091'
        - '10.2.0.203:9091'
    
  - job_name: 'consul'
    static_configs:
      - targets: ['10.0.0.10:8500']
    
  - job_name: 'node-exporters'
    static_configs:
      - targets:
        - '10.1.0.101:9100'
        - '10.1.0.102:9100'
        - '10.1.0.103:9100'
```

### Grafana Dashboards

1. **HBF Agent Overview**
   - Active services
   - Firewall rules count
   - Health check status
   - API request rate

2. **Service Mesh Metrics**
   - Service discovery events
   - Load balancer distribution
   - Circuit breaker status
   - Request latency

3. **Network Traffic**
   - Inbound/outbound traffic
   - Connection count
   - Blocked connections
   - Traffic by protocol

4. **System Resources**
   - CPU usage
   - Memory usage
   - Disk I/O
   - Network bandwidth

## Testing Automation

### Test Suite Structure

```
test-infrastructure/
├── terraform/           # Infrastructure provisioning
├── ansible/            # Configuration management
├── tests/
│   ├── unit/          # Unit tests
│   ├── integration/   # Integration tests
│   ├── e2e/           # End-to-end tests
│   └── performance/   # Performance tests
└── scripts/
    ├── setup.sh       # Environment setup
    ├── deploy.sh      # Deployment script
    ├── test.sh        # Test execution
    └── cleanup.sh     # Cleanup script
```

### Automated Test Pipeline

```yaml
stages:
  - provision
  - configure
  - deploy
  - test
  - cleanup

provision:
  script:
    - terraform apply -auto-approve
    
configure:
  script:
    - ansible-playbook -i inventory configure.yml
    
deploy:
  script:
    - ./scripts/deploy.sh
    
test:
  script:
    - ./scripts/test.sh
    
cleanup:
  script:
    - terraform destroy -auto-approve
```

## Resource Requirements

### Minimum Requirements

- **Total vCPUs**: 30
- **Total RAM**: 60GB
- **Total Disk**: 500GB
- **Network**: 1Gbps

### Recommended Requirements

- **Total vCPUs**: 48
- **Total RAM**: 96GB
- **Total Disk**: 1TB
- **Network**: 10Gbps

## Deployment Options

### Option 1: Local VMs (VirtualBox/VMware)

```bash
cd test-infrastructure/vagrant
vagrant up
```

### Option 2: Cloud (AWS/GCP/Azure)

```bash
cd test-infrastructure/terraform
terraform init
terraform apply
```

### Option 3: Docker Compose

```bash
cd test-infrastructure/docker-compose
docker-compose up -d
```

### Option 4: Kubernetes

```bash
cd test-infrastructure/kubernetes
kubectl apply -f manifests/
```

## Maintenance

### Backup Strategy

- **Configuration**: Daily backups
- **Metrics**: 30-day retention
- **Logs**: 7-day retention
- **Snapshots**: Weekly VM snapshots

### Update Procedure

1. Update HBF agent on one node
2. Verify functionality
3. Roll out to remaining nodes
4. Monitor for issues
5. Rollback if necessary

## Troubleshooting

### Common Issues

1. **Service Discovery Failures**
   - Check Consul connectivity
   - Verify network routes
   - Review agent logs

2. **Firewall Rule Issues**
   - Verify iptables/nftables
   - Check rule syntax
   - Review kernel logs

3. **Performance Problems**
   - Monitor resource usage
   - Check network latency
   - Review metrics

## Next Steps

1. Provision infrastructure
2. Deploy HBF agents
3. Configure monitoring
4. Run test scenarios
5. Analyze results
6. Optimize configuration
