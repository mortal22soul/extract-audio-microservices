# Video Converter Monitoring Stack

This directory contains the complete monitoring and observability stack for the Video Converter
microservices application.

## Components

### 📊 Metrics Collection

- **Prometheus**: Metrics collection and alerting
- **Grafana**: Metrics visualization and dashboards

### 📋 Logging

- **Elasticsearch**: Log storage and search
- **Logstash**: Log processing and transformation
- **Kibana**: Log visualization and analysis
- **Filebeat**: Log collection from Kubernetes pods

### 🔍 Distributed Tracing

- **Jaeger**: Distributed tracing and performance monitoring

## Quick Start

### Prerequisites

- Kubernetes cluster (1.20+)
- kubectl configured
- At least 8GB RAM and 4 CPU cores available

### Deployment

1. **Deploy the entire monitoring stack:**

   ```bash
   # From the infrastructure/k8s directory
   ./deploy-monitoring.sh
   ```

2. **Or deploy components individually:**
   ```bash
   # Deploy in order (dependencies matter)
   kubectl apply -f monitoring/elasticsearch.yaml
   kubectl apply -f monitoring/logstash.yaml
   kubectl apply -f monitoring/kibana.yaml
   kubectl apply -f monitoring/filebeat.yaml
   kubectl apply -f monitoring/prometheus.yaml
   kubectl apply -f monitoring/grafana.yaml
   kubectl apply -f monitoring/jaeger.yaml
   ```

### Access Services

Use kubectl port-forward to access the services locally:

```bash
# Grafana (Metrics Dashboard)
kubectl port-forward svc/grafana 3000:3000
# Access: http://localhost:3000 (admin/admin123)

# Prometheus (Metrics)
kubectl port-forward svc/prometheus 9090:9090
# Access: http://localhost:9090

# Kibana (Logs)
kubectl port-forward svc/kibana 5601:5601
# Access: http://localhost:5601

# Jaeger (Tracing)
kubectl port-forward svc/jaeger 16686:16686
# Access: http://localhost:16686
```

## Configuration

### Service Metrics

Each microservice should expose metrics on `/metrics` endpoint. The Prometheus configuration
automatically discovers services with the following annotations:

```yaml
annotations:
  prometheus.io/scrape: 'true'
  prometheus.io/port: '8080'
  prometheus.io/path: '/metrics'
```

### Logging

Services should log to stdout/stderr in JSON format for best results:

```json
{
  "timestamp": "2023-10-24T10:30:00Z",
  "level": "INFO",
  "service": "gateway-service",
  "message": "Request processed",
  "request_id": "req-123",
  "duration_ms": 45
}
```

### Distributed Tracing

Services should send traces to Jaeger agent:

- **Agent endpoint**: `jaeger-agent:6831` (UDP)
- **Collector endpoint**: `jaeger:14268` (HTTP)

## Dashboards

### Grafana Dashboards

The deployment includes a pre-configured dashboard for Video Converter services:

- **Service Status**: Health status of all services
- **Request Rate**: HTTP requests per second
- **Response Time**: 95th and 50th percentile response times
- **Error Rate**: HTTP 5xx error percentage
- **CPU Usage**: Container CPU utilization
- **Memory Usage**: Container memory consumption

### Custom Dashboards

To add custom dashboards:

1. Create a ConfigMap with your dashboard JSON
2. Mount it in the Grafana deployment
3. Add it to the dashboard provider configuration

## Alerts

### Prometheus Alerts

Pre-configured alerts include:

- High CPU usage (>80% for 5 minutes)
- High memory usage (>85% for 5 minutes)
- Service down (service unavailable for 1 minute)
- High error rate (>10% for 5 minutes)

### Adding Custom Alerts

Edit the `prometheus-config` ConfigMap and add rules to `alerts.yml`:

```yaml
- alert: CustomAlert
  expr: your_metric > threshold
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: 'Custom alert triggered'
    description: 'Description of the alert'
```

## Troubleshooting

### Common Issues

1. **Elasticsearch won't start**
   - Check if vm.max_map_count is set correctly
   - Ensure sufficient memory is available
   - Check storage class availability

2. **Filebeat not collecting logs**
   - Verify DaemonSet is running on all nodes
   - Check if log paths are correct
   - Ensure proper RBAC permissions

3. **Grafana shows no data**
   - Verify Prometheus is scraping targets
   - Check if services expose metrics correctly
   - Confirm datasource configuration

4. **High resource usage**
   - Adjust retention policies
   - Reduce scrape intervals
   - Limit log collection scope

### Useful Commands

```bash
# Check all monitoring pods
kubectl get pods -l 'app in (grafana,prometheus,kibana,elasticsearch,logstash,jaeger,filebeat)'

# View service logs
kubectl logs -f deployment/grafana
kubectl logs -f deployment/prometheus

# Check Prometheus targets
kubectl port-forward svc/prometheus 9090:9090
# Visit http://localhost:9090/targets

# Check Elasticsearch health
kubectl port-forward svc/elasticsearch 9200:9200
curl http://localhost:9200/_cluster/health

# Scale components
kubectl scale deployment grafana --replicas=2
kubectl scale deployment prometheus --replicas=2
```

## Resource Requirements

### Minimum Requirements

- **CPU**: 4 cores
- **Memory**: 8GB RAM
- **Storage**: 50GB

### Production Requirements

- **CPU**: 8+ cores
- **Memory**: 16+ GB RAM
- **Storage**: 200+ GB SSD

### Per Component

| Component     | CPU Request | Memory Request | Storage |
| ------------- | ----------- | -------------- | ------- |
| Elasticsearch | 500m        | 1Gi            | 10Gi    |
| Logstash      | 500m        | 1Gi            | -       |
| Kibana        | 500m        | 512Mi          | -       |
| Filebeat      | 100m        | 256Mi          | -       |
| Prometheus    | 500m        | 512Mi          | -       |
| Grafana       | 250m        | 256Mi          | -       |
| Jaeger        | 250m        | 256Mi          | -       |

## Security Considerations

1. **Network Policies**: Implement network policies to restrict traffic
2. **RBAC**: Use least-privilege RBAC for service accounts
3. **Secrets**: Store sensitive data in Kubernetes secrets
4. **TLS**: Enable TLS for inter-service communication
5. **Authentication**: Configure authentication for Grafana and Kibana

## Backup and Recovery

### Elasticsearch Backup

```bash
# Create snapshot repository
curl -X PUT "localhost:9200/_snapshot/backup_repo" -H 'Content-Type: application/json' -d'
{
  "type": "fs",
  "settings": {
    "location": "/backup"
  }
}'

# Create snapshot
curl -X PUT "localhost:9200/_snapshot/backup_repo/snapshot_1"
```

### Prometheus Backup

Prometheus data is stored in `/prometheus/` directory. Use persistent volumes for production.

## Monitoring Best Practices

1. **Metrics Naming**: Use consistent metric naming conventions
2. **Labels**: Use labels wisely to avoid high cardinality
3. **Retention**: Set appropriate retention policies
4. **Alerting**: Create actionable alerts, avoid alert fatigue
5. **Documentation**: Document custom metrics and dashboards

## Integration with CI/CD

### Helm Integration

The monitoring stack can be deployed using Helm charts for better configuration management.

### GitOps Integration

Use ArgoCD or Flux to manage monitoring stack deployments with GitOps practices.

## Support

For issues and questions:

1. Check the troubleshooting section
2. Review Kubernetes events: `kubectl get events`
3. Check component logs
4. Consult official documentation for each component
