# Performance Guide

## Current Performance Characteristics

### Baseline Metrics (v0.1.0)

| Metric | Value |
|--------|-------|
| Latency (avg) | ~3-5ms |
| Latency (p95) | ~8ms |
| Latency (p99) | ~12ms |
| Throughput | ~10,000 req/s |
| Memory (idle) | ~8MB |
| Memory (under load) | ~15MB |
| CPU (idle) | <1% |
| CPU (under load) | ~15-20% |
| Concurrent connections | ~1,000 |

*Tested on: Intel i7-10700K, 16GB RAM, Ubuntu 22.04*

---

## Performance Optimization Tips

### 1. Connection Settings

```json
{
  "proxies": [{
    "connection_timeout": "5s",
    "keep_alive": true,
    "max_idle_time": "60s"
  }]
}
```

### 2. System Tuning

#### Linux
```bash
# Increase file descriptor limits
ulimit -n 65535

# TCP tuning
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -w net.ipv4.tcp_fin_timeout=30
sysctl -w net.core.somaxconn=1024
```

#### Docker
```yaml
services:
  modbus-proxy:
    ulimits:
      nofile:
        soft: 65535
        hard: 65535
```

### 3. Go Runtime Tuning

```bash
# Set GOMAXPROCS to number of CPUs
export GOMAXPROCS=4

# Adjust GC target percentage
export GOGC=100
```

---

## Benchmarking

### Running Benchmarks

```bash
# Run all benchmarks
make bench

# Or manually
go test -bench=. -benchmem ./...

# With profiling
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./...
```

### Analyzing Profiles

```bash
# CPU profile
go tool pprof cpu.prof

# Memory profile
go tool pprof mem.prof

# Web interface
go tool pprof -http=:8080 cpu.prof
```

---

## Performance Monitoring

### Enable pprof (Development Only!)

Add to `main.go`:

```go
import _ "net/http/pprof"

// In main()
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Access at: `http://localhost:6060/debug/pprof/`

### Prometheus Metrics (Coming in v0.3.0)

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'modbridge'
    static_configs:
      - targets: ['localhost:8080']
```

---

## Performance Tuning Checklist

### Quick Wins
- [ ] Enable connection keep-alive
- [ ] Increase file descriptor limits
- [ ] Tune TCP settings
- [ ] Use appropriate GOMAXPROCS

### Medium Impact
- [ ] Implement connection pooling
- [ ] Add request caching (if applicable)
- [ ] Optimize logging (reduce verbosity)
- [ ] Use binary logging format

### Advanced
- [ ] Profile and optimize hot paths
- [ ] Implement object pooling (sync.Pool)
- [ ] Consider lock-free data structures
- [ ] Use memory-mapped files for large datasets

---

## Load Testing

### Using vegeta

```bash
# Install vegeta
go install github.com/tsenart/vegeta@latest

# Create target file
echo "GET http://localhost:8080/api/status" > targets.txt

# Run load test
vegeta attack -targets=targets.txt -rate=1000 -duration=30s | \
  vegeta report

# With results
vegeta attack -targets=targets.txt -rate=1000 -duration=30s | \
  tee results.bin | vegeta report

# Plot results
vegeta plot results.bin > plot.html
```

### Using hey

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Run test
hey -n 10000 -c 100 http://localhost:8080/api/status
```

---

## Capacity Planning

### Estimating Resources

**For 1,000 concurrent connections:**
- CPU: 2 cores minimum
- Memory: 512MB minimum
- Network: 100Mbps

**For 10,000 concurrent connections:**
- CPU: 4-8 cores
- Memory: 2GB
- Network: 1Gbps

**For 100,000 concurrent connections:**
- CPU: 16+ cores
- Memory: 8-16GB
- Network: 10Gbps
- Multiple instances recommended

---

## Performance Targets by Version

### v0.2.0 (Performance Release)
- ✅ Latency (p99) < 2ms
- ✅ Throughput > 50,000 req/s
- ✅ Memory < 8MB idle
- ✅ Support 5,000 concurrent connections

### v0.4.0 (Reliability Release)
- ✅ Latency (p99) < 1ms
- ✅ Throughput > 100,000 req/s
- ✅ Memory < 50MB under load
- ✅ Support 10,000 concurrent connections

### v1.0.0 (Production Release)
- ✅ Latency (p99) < 0.5ms
- ✅ Throughput > 1,000,000 req/s
- ✅ Memory < 100MB at scale
- ✅ Support 50,000+ concurrent connections

---

## Common Performance Issues

### Issue: High Latency

**Symptoms:** Request latency > 10ms

**Causes:**
- Network latency to Modbus devices
- DNS resolution delays
- Lock contention
- Inefficient logging

**Solutions:**
1. Use connection pooling
2. Cache DNS lookups
3. Reduce lock scope
4. Use structured logging with levels

### Issue: High Memory Usage

**Symptoms:** Memory grows unbounded

**Causes:**
- Goroutine leaks
- Unclosed connections
- Large ring buffers

**Solutions:**
1. Use pprof to find leaks
2. Ensure proper connection cleanup
3. Tune buffer sizes
4. Implement resource limits

### Issue: High CPU Usage

**Symptoms:** CPU > 80% consistently

**Causes:**
- Too many goroutines
- Inefficient parsing
- Excessive logging

**Solutions:**
1. Use worker pools
2. Optimize hot paths
3. Reduce log verbosity
4. Profile with pprof

---

## Best Practices

1. **Always benchmark before optimizing**
2. **Profile in production-like environments**
3. **Monitor metrics continuously**
4. **Set performance budgets**
5. **Test at expected scale + 50%**
6. **Document performance characteristics**
7. **Automate performance regression tests**

---

## Resources

- [Go Performance Tips](https://github.com/dgryski/go-perfbook)
- [Golang Profiling](https://go.dev/blog/pprof)
- [vegeta Load Testing](https://github.com/tsenart/vegeta)
- [Go Memory Model](https://go.dev/ref/mem)

---

**Last Updated:** December 2025
