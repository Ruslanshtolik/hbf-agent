# HBF Agent Architecture

## Overview

The HBF Agent is designed as a modular, high-performance system that combines host-based firewall capabilities with service mesh functionality for Linux hosts.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         HBF Agent                               │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Agent Core                            │  │
│  │  - Lifecycle Management                                  │  │
│  │  - Configuration Management                              │  │
│  │  - Component Coordination                                │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │
│  │  Firewall   │  │   Service   │  │   Health    │           │
│  │  Manager    │  │    Mesh     │  │   Checker   │           │
│  │             │  │   Manager   │  │             │           │
│  │ - iptables  │  │ - Discovery │  │ - HTTP      │           │
│  │ - nftables  │  │ - Registry  │  │ - TCP       │           │
│  │ - Rules     │  │ - Routing   │  │ - gRPC      │           │
│  └─────────────┘  └─────────────┘  └─────────────┘           │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │
│  │   Metrics   │  │     API     │  │   Traffic   │           │
│  │   Manager   │  │   Server    │  │   Manager   │           │
│  │             │  │             │  │             │           │
│  │ - Prometheus│  │ - REST API  │  │ - Load Bal. │           │
│  │ - Logging   │  │ - Health    │  │ - Circuit   │           │
│  │ - Tracing   │  │ - Metrics   │  │   Breaker   │           │
│  └─────────────┘  └─────────────┘  └─────────────┘           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Linux Kernel                                 │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  Netfilter   │  │   Network    │  │   Conntrack  │         │
│  │  (iptables)  │  │    Stack     │  │              │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Agent Core

The agent core is responsible for:
- Initializing and coordinating all components
- Managing the agent lifecycle (start, stop, restart)
- Loading and validating configuration
- Handling graceful shutdown

**Key Files:**
- [`internal/agent/agent.go`](../internal/agent/agent.go)
- [`internal/config/config.go`](../internal/config/config.go)

### 2. Firewall Manager

Manages host-based firewall rules using iptables or nftables.

**Features:**
- Dynamic rule management
- Support for both iptables and nftables backends
- Rule validation and rollback
- Automatic rule synchronization
- Connection tracking

**Key Files:**
- [`internal/firewall/manager.go`](../internal/firewall/manager.go)

**Rule Flow:**
```
API Request → Validation → Backend (iptables/nftables) → Kernel
                ↓
            Rule Store
                ↓
          Sync Monitor
```

### 3. Service Mesh Manager

Provides service mesh capabilities including service discovery, registration, and routing.

**Features:**
- Service registration and discovery
- Multiple discovery backends (Consul, etcd, DNS, static)
- Load balancing strategies (round-robin, least-conn, random, weighted)
- Circuit breaker pattern
- Health-based routing

**Key Files:**
- [`internal/servicemesh/manager.go`](../internal/servicemesh/manager.go)
- [`internal/servicemesh/loadbalancer.go`](../internal/servicemesh/loadbalancer.go)

**Service Flow:**
```
Service Registration → Discovery Backend → Service Registry
                                              ↓
Client Request → Load Balancer → Service Selection → Health Check
                                              ↓
                                        Route to Service
```

### 4. Health Checker

Performs active health checks on registered services.

**Features:**
- Multiple check types (HTTP, TCP, gRPC)
- Configurable intervals and timeouts
- Failure threshold tracking
- Status callbacks

**Key Files:**
- [`internal/health/checker.go`](../internal/health/checker.go)

**Health Check Flow:**
```
Service Registration → Health Check Config
                            ↓
                    Periodic Check Loop
                            ↓
                    Status Update → Callback
                            ↓
                    Service Registry Update
```

### 5. Metrics Manager

Collects and exposes metrics using Prometheus.

**Metrics Categories:**
- Firewall metrics (rules, additions, deletions)
- Service mesh metrics (registrations, health, requests)
- Traffic metrics (bytes, connections)
- Health check metrics (checks, duration)
- Agent metrics (uptime, errors)

**Key Files:**
- [`internal/metrics/manager.go`](../internal/metrics/manager.go)

### 6. API Server

Provides REST API for agent management.

**Endpoints:**
- `/api/v1/health` - Agent health status
- `/api/v1/services` - Service management
- `/api/v1/firewall/rules` - Firewall rule management
- `/api/v1/metrics` - Metrics information

**Key Files:**
- [`internal/api/server.go`](../internal/api/server.go)

## Data Flow

### Service Registration Flow

```
1. Client → POST /api/v1/services
2. API Server → Service Mesh Manager
3. Service Mesh Manager → Discovery Backend
4. Discovery Backend → Service Registry
5. Health Checker → Start Health Checks
6. Metrics Manager → Update Service Count
```

### Firewall Rule Addition Flow

```
1. Client → POST /api/v1/firewall/rules
2. API Server → Firewall Manager
3. Firewall Manager → Validate Rule
4. Firewall Manager → Backend (iptables/nftables)
5. Backend → Kernel Netfilter
6. Firewall Manager → Store Rule
7. Metrics Manager → Update Rule Count
```

### Service Request Flow

```
1. Client → Request Service
2. Service Mesh Manager → Discover Service Instances
3. Load Balancer → Select Instance
4. Health Checker → Verify Health
5. Traffic Manager → Route Request
6. Metrics Manager → Record Request
```

## Configuration

Configuration is managed through YAML files with the following structure:

```yaml
agent:          # Agent-level settings
firewall:       # Firewall configuration
service_mesh:   # Service mesh configuration
security:       # Security settings (mTLS, auth, rate limiting)
monitoring:     # Monitoring and metrics
log:            # Logging configuration
```

See [`config/config.example.yaml`](../config/config.example.yaml) for details.

## Security Model

### Capabilities

The agent requires the following Linux capabilities:
- `CAP_NET_ADMIN` - For firewall rule management
- `CAP_NET_RAW` - For raw socket access

### mTLS Support

When enabled, the agent supports mutual TLS for:
- Service-to-service communication
- API authentication
- Discovery backend connections

### Authentication

Multiple authentication methods:
- API tokens
- JWT tokens
- mTLS client certificates

### Rate Limiting

Built-in rate limiting to prevent abuse:
- Configurable requests per second
- Burst capacity
- Per-endpoint limits

## Deployment Models

### 1. Standalone Mode

Single agent per host managing local firewall and services.

```
┌─────────────┐
│    Host     │
│             │
│  HBF Agent  │
│     ↓       │
│  iptables   │
└─────────────┘
```

### 2. Cluster Mode

Multiple agents coordinated through service discovery.

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   Host 1    │  │   Host 2    │  │   Host 3    │
│             │  │             │  │             │
│  HBF Agent  │  │  HBF Agent  │  │  HBF Agent  │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │
       └────────────────┼────────────────┘
                        │
                   ┌────▼────┐
                   │ Consul/ │
                   │  etcd   │
                   └─────────┘
```

### 3. Container Mode

Agent running in containers with host network access.

```
┌─────────────────────────────┐
│          Host               │
│                             │
│  ┌───────────────────────┐  │
│  │   Docker Container    │  │
│  │                       │  │
│  │     HBF Agent         │  │
│  │         ↓             │  │
│  │   Host Network        │  │
│  └───────────────────────┘  │
│             ↓               │
│        iptables             │
└─────────────────────────────┘
```

## Performance Considerations

### Memory Usage

- Base memory: ~50MB
- Per service: ~1KB
- Per firewall rule: ~500B
- Metrics storage: ~10MB

### CPU Usage

- Idle: <1%
- Active (100 services): ~5%
- Rule sync: <2%

### Network

- API server: Minimal overhead
- Service discovery: Periodic polling (configurable)
- Health checks: Configurable intervals

## Scalability

The agent is designed to handle:
- 10,000+ firewall rules
- 1,000+ registered services
- 100+ concurrent health checks
- 10,000+ requests per second (API)

## Monitoring and Observability

### Metrics

Prometheus metrics exposed at `/metrics`:
- Counter metrics for events
- Gauge metrics for current state
- Histogram metrics for durations

### Logging

Structured logging with configurable levels:
- DEBUG: Detailed operation logs
- INFO: Normal operation logs
- WARN: Warning conditions
- ERROR: Error conditions

### Health Checks

Multiple health endpoints:
- Agent health: `/api/v1/health`
- Metrics health: `/metrics`
- Component health: Internal monitoring

## Future Enhancements

1. **eBPF Support**: Use eBPF for more efficient packet filtering
2. **gRPC API**: Add gRPC API alongside REST
3. **Distributed Tracing**: OpenTelemetry integration
4. **Policy Engine**: Advanced policy management
5. **Multi-tenancy**: Support for multiple tenants
6. **WebAssembly**: Plugin system using WASM
