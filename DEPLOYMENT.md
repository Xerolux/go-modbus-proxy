# Modbridge Deployment Guide

Complete guide for deploying Modbridge in production environments.

## Table of Contents

- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Performance Tuning](#performance-tuning)
- [Monitoring](#monitoring)
- [High Availability](#high-availability)
- [Security](#security)

## Docker Deployment

### Quick Start with Docker Compose

```bash
# Clone repository
git clone https://github.com/Xerolux/modbridge.git
cd modbridge

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f modbridge

# Stop services
docker-compose down
```

### Build Docker Image

```bash
docker build -t modbridge:v1.0.0 .
```

### Run Standalone Container

```bash
docker run -d \
  --name modbridge \
  -p 8080:8080 \
  -p 5020:5020 \
  -p 5021:5021 \
  -v $(pwd)/config.json:/app/config.json:ro \
  -v modbridge-data:/data \
  --restart unless-stopped \
  modbridge:v1.0.0
```

### Environment Variables

```bash
# Core settings
MODBRIDGE_WEB_PORT=:8080
MODBRIDGE_LOG_LEVEL=info

# Proxy configuration (alternative to config file)
MODBRIDGE_PROXY_COUNT=2

# Proxy 0
MODBRIDGE_PROXY_0_ID=plc-gateway
MODBRIDGE_PROXY_0_NAME=PLC Gateway
MODBRIDGE_PROXY_0_LISTEN=:5020
MODBRIDGE_PROXY_0_TARGET=192.168.1.10:502
MODBRIDGE_PROXY_0_POOL_SIZE=20
MODBRIDGE_PROXY_0_ENABLED=true

# Proxy 1
MODBRIDGE_PROXY_1_ID=scada-proxy
MODBRIDGE_PROXY_1_NAME=SCADA Proxy
MODBRIDGE_PROXY_1_LISTEN=:5021
MODBRIDGE_PROXY_1_TARGET=10.0.0.5:502
MODBRIDGE_PROXY_1_POOL_SIZE=15
MODBRIDGE_PROXY_1_ENABLED=true
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes 1.24+
- kubectl configured
- Namespace created

### Create Namespace

```bash
kubectl create namespace modbridge
```

### Deploy

```bash
# Apply all manifests
kubectl apply -f deployments/kubernetes/

# Check deployment status
kubectl -n modbridge get pods
kubectl -n modbridge get svc
kubectl -n modbridge get hpa

# View logs
kubectl -n modbridge logs -f deployment/modbridge
```

### Scale Manually

```bash
# Scale to 5 replicas
kubectl -n modbridge scale deployment modbridge --replicas=5

# Get current replicas
kubectl -n modbridge get deployment modbridge
```

### Access Services

```bash
# Port forward for local access
kubectl -n modbridge port-forward svc/modbridge 8080:8080

# Get LoadBalancer IP
kubectl -n modbridge get svc modbridge-loadbalancer
```

### Update Configuration

```bash
# Edit ConfigMap
kubectl -n modbridge edit configmap modbridge-config

# Restart pods to apply changes
kubectl -n modbridge rollout restart deployment modbridge

# Monitor rollout
kubectl -n modbridge rollout status deployment modbridge
```

## Performance Tuning

### System-Level Optimizations

#### Linux Kernel Parameters

```bash
# /etc/sysctl.conf

# Increase max number of open files
fs.file-max = 2097152

# Increase number of connections
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 8192

# Enable TCP fast open
net.ipv4.tcp_fastopen = 3

# TCP buffer sizes
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216

# Disable TCP slow start after idle
net.ipv4.tcp_slow_start_after_idle = 0

# Enable TCP timestamps
net.ipv4.tcp_timestamps = 1

# Increase local port range
net.ipv4.ip_local_port_range = 10000 65535

# TCP keepalive settings
net.ipv4.tcp_keepalive_time = 30
net.ipv4.tcp_keepalive_intvl = 10
net.ipv4.tcp_keepalive_probes = 3

# Apply changes
sudo sysctl -p
```

#### Process Limits

```bash
# /etc/security/limits.conf

modbridge soft nofile 1048576
modbridge hard nofile 1048576
modbridge soft nproc 65536
modbridge hard nproc 65536
```

### Application-Level Optimizations

#### Enable All Performance Features

```json
{
  "performance": {
    "enable_pprof": true,
    "enable_dns_cache": true,
    "enable_buffer_pool": true,
    "enable_keep_alive": true,
    "enable_tcp_optimize": true
  }
}
```

#### Connection Pool Tuning

```json
{
  "proxies": [
    {
      "pool_size": 100,
      "pool_min_size": 10,
      "conn_timeout": 5,
      "conn_max_idle": 300,
      "conn_keep_alive": true,
      "health_check_interval": 60
    }
  ]
}
```

## Monitoring

### Prometheus Metrics

Add scrape configuration:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'modbridge'
    static_configs:
      - targets: ['modbridge:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Grafana Dashboards

Import dashboard from `deployments/grafana/dashboards/modbridge.json`

Key metrics to monitor:
- Request rate (requests/sec)
- Error rate (%)
- Latency (P50, P95, P99)
- Active connections
- Memory usage
- CPU usage
- Goroutines

### InfluxDB Integration

Configure InfluxDB export:

```json
{
  "influxdb": {
    "enabled": true,
    "url": "http://influxdb:8086",
    "token": "your-token",
    "organization": "modbridge",
    "bucket": "modbridge",
    "batch_size": 100,
    "flush_interval": "10s"
  }
}
```

### Logging

Set log level via environment:

```bash
export MODBRIDGE_LOG_LEVEL=debug  # debug, info, warn, error
```

View structured logs:

```bash
# Docker
docker logs -f modbridge --tail 100

# Kubernetes
kubectl -n modbridge logs -f deployment/modbridge --tail 100

# Filter for errors
kubectl -n modbridge logs deployment/modbridge | grep ERROR
```

## High Availability

### Multi-Region Setup

Deploy in multiple regions with etcd for state synchronization:

```yaml
# US East
region: us-east-1
etcd_endpoints:
  - http://etcd-us-east:2379

# EU West
region: eu-west-1
etcd_endpoints:
  - http://etcd-eu-west:2379

# Replication
replication:
  enabled: true
  strategy: async
  regions:
    - us-east-1
    - eu-west-1
```

### Load Balancing

#### NGINX Configuration

```nginx
upstream modbridge {
    least_conn;
    server modbridge-1:5020 max_fails=3 fail_timeout=30s;
    server modbridge-2:5020 max_fails=3 fail_timeout=30s;
    server modbridge-3:5020 max_fails=3 fail_timeout=30s;
    keepalive 100;
}

server {
    listen 5020;
    proxy_pass modbridge;
    proxy_connect_timeout 5s;
    proxy_timeout 30s;
}
```

#### HAProxy Configuration

```haproxy
frontend modbus_front
    bind *:5020
    mode tcp
    option tcplog
    default_backend modbus_back

backend modbus_back
    mode tcp
    balance leastconn
    option tcp-check
    server modbridge-1 modbridge-1:5020 check
    server modbridge-2 modbridge-2:5020 check
    server modbridge-3 modbridge-3:5020 check
```

### Graceful Shutdown

Modbridge supports graceful shutdown with connection draining:

```bash
# Send SIGTERM for graceful shutdown
kubectl -n modbridge delete pod modbridge-xyz --grace-period=30

# Connections will drain for up to 30 seconds
```

## Security

### TLS/SSL

Enable TLS for API:

```json
{
  "server": {
    "tls_enabled": true,
    "cert_file": "/certs/server.crt",
    "key_file": "/certs/server.key"
  }
}
```

### Authentication

Configure JWT authentication:

```json
{
  "auth": {
    "enabled": true,
    "jwt_secret": "your-secret-key",
    "token_ttl": "24h"
  }
}
```

### Network Policies (Kubernetes)

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: modbridge-netpol
  namespace: modbridge
spec:
  podSelector:
    matchLabels:
      app: modbridge
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: production
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 5020
  egress:
  - to:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 502
```

### IP Filtering

Configure IP whitelist/blacklist:

```json
{
  "security": {
    "ip_whitelist": [
      "192.168.1.0/24",
      "10.0.0.0/8"
    ],
    "ip_blacklist": [
      "192.168.1.100"
    ]
  }
}
```

## Troubleshooting

### Common Issues

#### High CPU Usage

```bash
# Enable pprof
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Check for tight loops or CPU-bound operations
```

#### High Memory Usage

```bash
# Get heap profile
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Check for memory leaks
```

#### Connection Issues

```bash
# Verify network connectivity
modbusctl validate proxy --listen :5020 --target 192.168.1.10:502

# Check firewall rules
sudo iptables -L -n | grep 5020

# Test with netcat
nc -zv 192.168.1.10 502
```

#### Performance Degradation

```bash
# Check metrics
modbusctl metrics overview

# View per-proxy metrics
modbusctl metrics proxy --proxy plc-gateway

# Check system resources
modbusctl metrics system
```

## Backup and Recovery

### Backup Configuration

```bash
# Export configuration
modbusctl backup create --name "pre-upgrade-$(date +%Y%m%d)"

# List backups
modbusctl backup list

# Export backup to file
modbusctl backup export backup-20240115 -o backup.tar.gz
```

### Restore Configuration

```bash
# Restore from backup
modbusctl backup restore backup-20240115 --confirm

# Or import from file
modbusctl backup import -f backup.tar.gz
```

## Best Practices

1. **Always use load balancer** for production deployments
2. **Enable all performance features** for high-throughput scenarios
3. **Monitor metrics** continuously with Prometheus/Grafana
4. **Set resource limits** in Kubernetes to prevent resource exhaustion
5. **Use HPA** for automatic scaling based on load
6. **Configure backups** regularly
7. **Test failover scenarios** before production
8. **Use TLS** for API endpoints
9. **Implement network policies** in Kubernetes
10. **Keep logs** for at least 30 days for troubleshooting

## Support

For issues and questions:
- GitHub Issues: https://github.com/Xerolux/modbridge/issues
- Documentation: https://github.com/Xerolux/modbridge/wiki
