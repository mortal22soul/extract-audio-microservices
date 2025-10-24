# Troubleshooting and Maintenance Guide

This guide provides comprehensive troubleshooting procedures and maintenance tasks for the Video
Converter microservices platform.

## Table of Contents

1. [Common Issues](#common-issues)
2. [Service-Specific Troubleshooting](#service-specific-troubleshooting)
3. [Database Issues](#database-issues)
4. [Performance Troubleshooting](#performance-troubleshooting)
5. [Network and Connectivity](#network-and-connectivity)
6. [Monitoring and Alerting](#monitoring-and-alerting)
7. [Maintenance Tasks](#maintenance-tasks)
8. [Emergency Procedures](#emergency-procedures)

## Common Issues

### 1. Pod Startup Failures

#### Symptoms

- Pods stuck in `Pending`, `CrashLoopBackOff`, or `ImagePullBackOff` state
- Services not responding to health checks

#### Diagnosis

```bash
# Check pod status
kubectl get pods -o wide

# Describe pod for detailed events
kubectl describe pod <pod-name>

# Check recent logs
kubectl logs <pod-name> --tail=100

# Check previous container logs if pod restarted
kubectl logs <pod-name> --previous
```

#### Common Causes and Solutions

**ImagePullBackOff**:

```bash
# Check image name and tag
kubectl describe pod <pod-name> | grep Image

# Verify image exists in registry
docker pull <image-name>

# Check image pull secrets
kubectl get secrets
kubectl describe secret <image-pull-secret>
```

**Resource Constraints**:

```bash
# Check node resources
kubectl top nodes
kubectl describe node <node-name>

# Check resource requests/limits
kubectl describe pod <pod-name> | grep -A 5 "Requests\|Limits"

# Adjust resource limits in deployment
kubectl patch deployment <deployment-name> -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container-name>","resources":{"limits":{"memory":"2Gi","cpu":"1000m"}}}]}}}}'
```

**Configuration Issues**:

```bash
# Check ConfigMaps
kubectl get configmaps
kubectl describe configmap <configmap-name>

# Check Secrets
kubectl get secrets
kubectl describe secret <secret-name>

# Verify environment variables
kubectl exec -it <pod-name> -- env | grep -i <variable-name>
```

### 2. Service Discovery Issues

#### Symptoms

- Services cannot communicate with each other
- gRPC connection failures
- DNS resolution errors

#### Diagnosis

```bash
# Check service endpoints
kubectl get endpoints

# Test DNS resolution
kubectl run debug --image=busybox -it --rm -- nslookup <service-name>

# Test service connectivity
kubectl run debug --image=busybox -it --rm -- telnet <service-name> <port>

# Check service configuration
kubectl describe service <service-name>
```

#### Solutions

**DNS Issues**:

```bash
# Check CoreDNS pods
kubectl get pods -n kube-system -l k8s-app=kube-dns

# Check CoreDNS configuration
kubectl get configmap coredns -n kube-system -o yaml

# Restart CoreDNS if needed
kubectl rollout restart deployment coredns -n kube-system
```

**Network Policies**:

```bash
# Check network policies
kubectl get networkpolicies

# Test without network policies (temporarily)
kubectl delete networkpolicy <policy-name>

# Verify pod labels match policy selectors
kubectl get pods --show-labels
```

### 3. Authentication and Authorization Issues

#### Symptoms

- JWT token validation failures
- gRPC authentication errors
- User login failures

#### Diagnosis

```bash
# Check Auth Service logs
kubectl logs -l app=auth-service --tail=100

# Test token validation endpoint
kubectl exec -it <gateway-pod> -- curl -H "Authorization: Bearer <token>" http://auth-service:50051/health

# Check JWT secret configuration
kubectl get secret app-secrets -o yaml | base64 -d
```

#### Solutions

**JWT Secret Mismatch**:

```bash
# Update JWT secret across all services
kubectl create secret generic app-secrets \
  --from-literal=jwt-secret=<new-secret> \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart services to pick up new secret
kubectl rollout restart deployment auth-service
kubectl rollout restart deployment gateway-service
```

**Database Connection Issues**:

```bash
# Test database connectivity from Auth Service
kubectl exec -it <auth-pod> -- psql -h postgresql -U postgres -d video_converter

# Check database credentials
kubectl get secret database-secrets -o yaml
```

## Service-Specific Troubleshooting

### Gateway Service

#### Common Issues

**High Memory Usage**:

```bash
# Check memory usage
kubectl top pod -l app=gateway-service

# Check for memory leaks in logs
kubectl logs -l app=gateway-service | grep -i "memory\|leak\|gc"

# Increase memory limits
kubectl patch deployment gateway-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"gateway","resources":{"limits":{"memory":"2Gi"}}}]}}}}'
```

**File Upload Failures**:

```bash
# Check upload size limits
kubectl logs -l app=gateway-service | grep -i "upload\|size"

# Verify MongoDB GridFS connection
kubectl exec -it <gateway-pod> -- mongo mongodb://mongodb:27017/video_converter

# Check disk space on nodes
kubectl get nodes -o wide
kubectl describe node <node-name> | grep -i "disk\|storage"
```

### Auth Service

#### Common Issues

**Database Connection Pool Exhaustion**:

```bash
# Check connection pool metrics
kubectl exec -it <auth-pod> -- curl localhost:8080/metrics | grep db_connections

# Check PostgreSQL connections
kubectl exec -it postgresql-0 -- psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# Adjust connection pool settings
kubectl patch deployment auth-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"auth","env":[{"name":"DB_MAX_CONNECTIONS","value":"50"}]}]}}}}'
```

**Slow Query Performance**:

```bash
# Enable query logging in PostgreSQL
kubectl exec -it postgresql-0 -- psql -U postgres -c "ALTER SYSTEM SET log_statement = 'all';"
kubectl exec -it postgresql-0 -- psql -U postgres -c "SELECT pg_reload_conf();"

# Check slow queries
kubectl exec -it postgresql-0 -- psql -U postgres -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

# Add missing indexes
kubectl exec -it postgresql-0 -- psql -U postgres video_converter -c "CREATE INDEX CONCURRENTLY idx_users_email ON users(email);"
```

### Converter Service

#### Common Issues

**FFmpeg Conversion Failures**:

```bash
# Check FFmpeg logs
kubectl logs -l app=converter-service | grep -i ffmpeg

# Test FFmpeg installation
kubectl exec -it <converter-pod> -- ffmpeg -version

# Check available codecs
kubectl exec -it <converter-pod> -- ffmpeg -codecs | grep mp3
```

**RabbitMQ Connection Issues**:

```bash
# Check RabbitMQ status
kubectl exec -it rabbitmq-0 -- rabbitmqctl status

# Check queue status
kubectl exec -it rabbitmq-0 -- rabbitmqctl list_queues

# Check connection from converter
kubectl exec -it <converter-pod> -- amqp-tools list connections
```

**High CPU Usage During Conversion**:

```bash
# Check CPU usage
kubectl top pod -l app=converter-service

# Limit concurrent conversions
kubectl patch deployment converter-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"converter","env":[{"name":"MAX_CONCURRENT_JOBS","value":"2"}]}]}}}}'

# Add more converter replicas
kubectl scale deployment converter-service --replicas=5
```

### Analytics Service

#### Common Issues

**ML Model Loading Failures**:

```bash
# Check model loading logs
kubectl logs -l app=analytics-service | grep -i "model\|load"

# Check available disk space
kubectl exec -it <analytics-pod> -- df -h

# Verify model files
kubectl exec -it <analytics-pod> -- ls -la /app/models/
```

**Python Memory Issues**:

```bash
# Check Python memory usage
kubectl exec -it <analytics-pod> -- python -c "import psutil; print(f'Memory: {psutil.virtual_memory().percent}%')"

# Check for memory leaks
kubectl logs -l app=analytics-service | grep -i "memory\|oom"

# Increase memory limits
kubectl patch deployment analytics-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"analytics","resources":{"limits":{"memory":"4Gi"}}}]}}}}'
```

### Realtime Service

#### Common Issues

**WebSocket Connection Drops**:

```bash
# Check WebSocket connections
kubectl logs -l app=realtime-service | grep -i "websocket\|disconnect"

# Check Redis pub/sub
kubectl exec -it redis-0 -- redis-cli monitor

# Test WebSocket endpoint
kubectl port-forward svc/realtime-service 3001:3001
# Use WebSocket client to test ws://localhost:3001
```

**High Connection Count**:

```bash
# Check active connections
kubectl exec -it <realtime-pod> -- curl localhost:3001/metrics | grep websocket_connections

# Implement connection limits
kubectl patch deployment realtime-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"realtime","env":[{"name":"MAX_CONNECTIONS","value":"1000"}]}]}}}}'
```

## Database Issues

### PostgreSQL

#### Connection Issues

```bash
# Check PostgreSQL status
kubectl exec -it postgresql-0 -- pg_isready

# Check active connections
kubectl exec -it postgresql-0 -- psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# Check connection limits
kubectl exec -it postgresql-0 -- psql -U postgres -c "SHOW max_connections;"

# Kill long-running queries
kubectl exec -it postgresql-0 -- psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'active' AND query_start < now() - interval '5 minutes';"
```

#### Performance Issues

```bash
# Check database size
kubectl exec -it postgresql-0 -- psql -U postgres -c "SELECT pg_size_pretty(pg_database_size('video_converter'));"

# Check table sizes
kubectl exec -it postgresql-0 -- psql -U postgres video_converter -c "SELECT schemaname,tablename,pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size FROM pg_tables ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"

# Analyze query performance
kubectl exec -it postgresql-0 -- psql -U postgres video_converter -c "EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';"

# Run VACUUM and ANALYZE
kubectl exec -it postgresql-0 -- psql -U postgres video_converter -c "VACUUM ANALYZE;"
```

### MongoDB

#### Connection Issues

```bash
# Check MongoDB status
kubectl exec -it mongodb-0 -- mongo --eval "db.adminCommand('ismaster')"

# Check connections
kubectl exec -it mongodb-0 -- mongo --eval "db.serverStatus().connections"

# Check replica set status (if applicable)
kubectl exec -it mongodb-0 -- mongo --eval "rs.status()"
```

#### GridFS Issues

```bash
# Check GridFS collections
kubectl exec -it mongodb-0 -- mongo video_converter --eval "db.fs.files.count()"
kubectl exec -it mongodb-0 -- mongo video_converter --eval "db.fs.chunks.count()"

# Check for orphaned chunks
kubectl exec -it mongodb-0 -- mongo video_converter --eval "
var orphanedChunks = [];
db.fs.chunks.find().forEach(function(chunk) {
  if (!db.fs.files.findOne({_id: chunk.files_id})) {
    orphanedChunks.push(chunk._id);
  }
});
print('Orphaned chunks: ' + orphanedChunks.length);
"

# Clean up orphaned chunks
kubectl exec -it mongodb-0 -- mongo video_converter --eval "
db.fs.chunks.remove({files_id: {$nin: db.fs.files.distinct('_id')}});
"
```

### Redis

#### Memory Issues

```bash
# Check Redis memory usage
kubectl exec -it redis-0 -- redis-cli info memory

# Check key distribution
kubectl exec -it redis-0 -- redis-cli --bigkeys

# Set memory policy
kubectl exec -it redis-0 -- redis-cli config set maxmemory-policy allkeys-lru

# Clear specific keys
kubectl exec -it redis-0 -- redis-cli flushdb
```

#### Pub/Sub Issues

```bash
# Check pub/sub channels
kubectl exec -it redis-0 -- redis-cli pubsub channels

# Monitor pub/sub activity
kubectl exec -it redis-0 -- redis-cli monitor

# Check client connections
kubectl exec -it redis-0 -- redis-cli client list
```

## Performance Troubleshooting

### High CPU Usage

#### Diagnosis

```bash
# Check CPU usage by pod
kubectl top pods --sort-by=cpu

# Check CPU usage by node
kubectl top nodes

# Check CPU throttling
kubectl describe pod <pod-name> | grep -i throttl
```

#### Solutions

```bash
# Increase CPU limits
kubectl patch deployment <deployment-name> -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container-name>","resources":{"limits":{"cpu":"2000m"}}}]}}}}'

# Scale horizontally
kubectl scale deployment <deployment-name> --replicas=5

# Enable HPA
kubectl autoscale deployment <deployment-name> --cpu-percent=70 --min=3 --max=10
```

### High Memory Usage

#### Diagnosis

```bash
# Check memory usage
kubectl top pods --sort-by=memory

# Check for OOMKilled pods
kubectl get pods | grep OOMKilled

# Check memory limits
kubectl describe pod <pod-name> | grep -A 5 "Limits"
```

#### Solutions

```bash
# Increase memory limits
kubectl patch deployment <deployment-name> -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container-name>","resources":{"limits":{"memory":"4Gi"}}}]}}}}'

# Check for memory leaks
kubectl logs <pod-name> | grep -i "memory\|leak\|gc"

# Restart pods to clear memory
kubectl rollout restart deployment <deployment-name>
```

### Slow Response Times

#### Diagnosis

```bash
# Check application metrics
kubectl port-forward svc/prometheus-server 9090:80
# Query: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Check database performance
kubectl exec -it postgresql-0 -- psql -U postgres -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

# Check network latency
kubectl exec -it <pod-name> -- ping <target-service>
```

#### Solutions

```bash
# Add database indexes
kubectl exec -it postgresql-0 -- psql -U postgres video_converter -c "CREATE INDEX CONCURRENTLY idx_videos_user_id ON videos(user_id);"

# Implement caching
kubectl patch deployment gateway-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"gateway","env":[{"name":"CACHE_TTL","value":"300"}]}]}}}}'

# Optimize queries
# Review and optimize slow queries identified in diagnosis
```

## Network and Connectivity

### DNS Resolution Issues

#### Diagnosis

```bash
# Test DNS resolution
kubectl run debug --image=busybox -it --rm -- nslookup kubernetes.default

# Check CoreDNS logs
kubectl logs -n kube-system -l k8s-app=kube-dns

# Check DNS configuration
kubectl get configmap coredns -n kube-system -o yaml
```

#### Solutions

```bash
# Restart CoreDNS
kubectl rollout restart deployment coredns -n kube-system

# Update DNS configuration
kubectl edit configmap coredns -n kube-system

# Check node DNS configuration
kubectl get nodes -o wide
```

### Network Policies

#### Diagnosis

```bash
# Check network policies
kubectl get networkpolicies

# Test connectivity without policies
kubectl delete networkpolicy <policy-name>
kubectl run debug --image=busybox -it --rm -- telnet <service-name> <port>
```

#### Solutions

```bash
# Update network policy to allow traffic
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-gateway-to-auth
spec:
  podSelector:
    matchLabels:
      app: auth-service
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: gateway-service
    ports:
    - protocol: TCP
      port: 50051
EOF
```

### Load Balancer Issues

#### Diagnosis

```bash
# Check service type and external IP
kubectl get services

# Check ingress status
kubectl get ingress

# Check load balancer logs (cloud provider specific)
# AWS: Check ALB/NLB logs in CloudWatch
# GCP: Check Load Balancer logs in Cloud Logging
```

## Monitoring and Alerting

### Prometheus Issues

#### Diagnosis

```bash
# Check Prometheus status
kubectl get pods -l app=prometheus

# Check Prometheus configuration
kubectl get configmap prometheus-config -o yaml

# Check targets
kubectl port-forward svc/prometheus-server 9090:80
# Navigate to http://localhost:9090/targets
```

#### Solutions

```bash
# Restart Prometheus
kubectl rollout restart deployment prometheus-server

# Update scrape configuration
kubectl edit configmap prometheus-config

# Check service discovery
kubectl get servicemonitor
```

### Grafana Issues

#### Diagnosis

```bash
# Check Grafana status
kubectl get pods -l app=grafana

# Check Grafana logs
kubectl logs -l app=grafana

# Access Grafana
kubectl port-forward svc/grafana 3000:80
```

#### Solutions

```bash
# Reset Grafana admin password
kubectl exec -it <grafana-pod> -- grafana-cli admin reset-admin-password newpassword

# Import dashboards
kubectl create configmap grafana-dashboards --from-file=dashboards/
```

## Maintenance Tasks

### Regular Maintenance

#### Daily Tasks

```bash
# Check cluster health
kubectl get nodes
kubectl get pods --all-namespaces | grep -v Running

# Check resource usage
kubectl top nodes
kubectl top pods --all-namespaces --sort-by=memory

# Check recent alerts
kubectl logs -l app=alertmanager --since=24h
```

#### Weekly Tasks

```bash
# Update container images
make update-images

# Run database maintenance
kubectl exec -it postgresql-0 -- psql -U postgres video_converter -c "VACUUM ANALYZE;"
kubectl exec -it mongodb-0 -- mongo video_converter --eval "db.runCommand({compact: 'videos'})"

# Check backup status
kubectl get cronjobs
kubectl get jobs --sort-by=.status.startTime
```

#### Monthly Tasks

```bash
# Review and rotate logs
kubectl logs --since=720h <pod-name> > archived-logs.txt

# Update Kubernetes cluster
# Follow cloud provider specific procedures

# Review and update resource limits
kubectl describe nodes | grep -A 5 "Allocated resources"
```

### Database Maintenance

#### PostgreSQL Maintenance

```bash
# Check database statistics
kubectl exec -it postgresql-0 -- psql -U postgres video_converter -c "
SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del, n_dead_tup
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;
"

# Reindex tables
kubectl exec -it postgresql-0 -- psql -U postgres video_converter -c "REINDEX DATABASE video_converter;"

# Update table statistics
kubectl exec -it postgresql-0 -- psql -U postgres video_converter -c "ANALYZE;"
```

#### MongoDB Maintenance

```bash
# Check collection statistics
kubectl exec -it mongodb-0 -- mongo video_converter --eval "db.stats()"

# Compact collections
kubectl exec -it mongodb-0 -- mongo video_converter --eval "db.runCommand({compact: 'videos'})"

# Check index usage
kubectl exec -it mongodb-0 -- mongo video_converter --eval "db.videos.getIndexes()"
```

### Log Rotation and Cleanup

#### Application Logs

```bash
# Set up log rotation
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: logrotate-config
data:
  logrotate.conf: |
    /var/log/app/*.log {
      daily
      rotate 7
      compress
      delaycompress
      missingok
      notifempty
      create 644 app app
    }
EOF
```

#### System Cleanup

```bash
# Clean up completed jobs
kubectl delete jobs --field-selector status.successful=1

# Clean up old replica sets
kubectl delete replicaset --all

# Clean up unused images on nodes
kubectl get nodes -o name | xargs -I {} kubectl debug {} -it --image=busybox -- docker system prune -f
```

## Emergency Procedures

### Service Outage Response

#### Immediate Actions

1. **Assess Impact**:

   ```bash
   # Check service status
   kubectl get pods --all-namespaces
   kubectl get services

   # Check recent events
   kubectl get events --sort-by=.metadata.creationTimestamp
   ```

2. **Identify Root Cause**:

   ```bash
   # Check logs for errors
   kubectl logs -l app=<affected-service> --tail=100

   # Check resource usage
   kubectl top nodes
   kubectl top pods
   ```

3. **Immediate Mitigation**:

   ```bash
   # Scale up healthy services
   kubectl scale deployment <service> --replicas=5

   # Restart failed services
   kubectl rollout restart deployment <service>

   # Rollback if recent deployment caused issue
   kubectl rollout undo deployment <service>
   ```

### Data Recovery

#### Database Recovery

```bash
# Stop application services
kubectl scale deployment --all --replicas=0

# Restore from backup
kubectl exec -it postgresql-0 -- psql -U postgres video_converter < /backup/latest.sql

# Verify data integrity
kubectl exec -it postgresql-0 -- psql -U postgres video_converter -c "SELECT count(*) FROM users;"

# Restart services
kubectl scale deployment --all --replicas=3
```

#### File Recovery

```bash
# Check GridFS integrity
kubectl exec -it mongodb-0 -- mongo video_converter --eval "
db.fs.files.find().forEach(function(file) {
  var chunks = db.fs.chunks.count({files_id: file._id});
  var expectedChunks = Math.ceil(file.length / file.chunkSize);
  if (chunks !== expectedChunks) {
    print('Corrupted file: ' + file._id + ' (' + file.filename + ')');
  }
});
"

# Restore files from backup
kubectl exec -it mongodb-0 -- mongorestore --uri="mongodb://localhost:27017/video_converter" /backup/gridfs/
```

### Security Incident Response

#### Immediate Actions

```bash
# Rotate all secrets
kubectl create secret generic app-secrets \
  --from-literal=jwt-secret=$(openssl rand -base64 32) \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart all services to pick up new secrets
kubectl rollout restart deployment --all

# Check for suspicious activity
kubectl logs --all-namespaces | grep -i "unauthorized\|failed\|error"
```

#### Investigation

```bash
# Check access logs
kubectl logs -l app=gateway-service | grep -E "POST|PUT|DELETE"

# Check authentication logs
kubectl logs -l app=auth-service | grep -i "login\|token"

# Export logs for analysis
kubectl logs --all-namespaces --since=24h > incident-logs.txt
```

### Contact Information

For escalation and support:

- **On-call Engineer**: +1-XXX-XXX-XXXX
- **DevOps Team**: devops@company.com
- **Security Team**: security@company.com
- **Incident Management**: incidents@company.com

### Runbooks

Detailed runbooks for specific scenarios:

- [Database Failover Runbook](runbooks/database-failover.md)
- [Security Incident Runbook](runbooks/security-incident.md)
- [Performance Degradation Runbook](runbooks/performance-degradation.md)
- [Disaster Recovery Runbook](runbooks/disaster-recovery.md)
