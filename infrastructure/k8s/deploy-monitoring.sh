#!/bin/bash

# Video Converter Monitoring Stack Deployment Script
set -e

echo "🚀 Deploying Video Converter Monitoring Stack..."

# Function to wait for deployment to be ready
wait_for_deployment() {
    local deployment=$1
    local namespace=${2:-default}
    echo "⏳ Waiting for $deployment to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/$deployment -n $namespace
    echo "✅ $deployment is ready"
}

# Function to wait for statefulset to be ready
wait_for_statefulset() {
    local statefulset=$1
    local namespace=${2:-default}
    echo "⏳ Waiting for $statefulset to be ready..."
    kubectl wait --for=condition=ready --timeout=300s pod -l app=$statefulset -n $namespace
    echo "✅ $statefulset is ready"
}

# Create namespace if it doesn't exist
echo "📁 Creating monitoring namespace..."
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -

# Deploy Elasticsearch first (required by Logstash and Kibana)
echo "🔍 Deploying Elasticsearch..."
kubectl apply -f monitoring/elasticsearch.yaml
wait_for_statefulset elasticsearch

# Deploy Logstash (depends on Elasticsearch)
echo "📊 Deploying Logstash..."
kubectl apply -f monitoring/logstash.yaml
wait_for_deployment logstash

# Deploy Kibana (depends on Elasticsearch)
echo "📈 Deploying Kibana..."
kubectl apply -f monitoring/kibana.yaml
wait_for_deployment kibana

# Deploy Filebeat (depends on Logstash)
echo "📋 Deploying Filebeat..."
kubectl apply -f monitoring/filebeat.yaml
echo "✅ Filebeat DaemonSet deployed"

# Deploy Prometheus
echo "📊 Deploying Prometheus..."
kubectl apply -f monitoring/prometheus.yaml
wait_for_deployment prometheus

# Deploy Grafana (depends on Prometheus)
echo "📈 Deploying Grafana..."
kubectl apply -f monitoring/grafana.yaml
wait_for_deployment grafana

# Deploy Jaeger for distributed tracing
echo "🔍 Deploying Jaeger..."
kubectl apply -f monitoring/jaeger.yaml
wait_for_deployment jaeger

echo ""
echo "🎉 Monitoring stack deployment completed!"
echo ""
echo "📊 Access URLs (use kubectl port-forward):"
echo "  Grafana:      kubectl port-forward svc/grafana 3000:3000"
echo "  Prometheus:   kubectl port-forward svc/prometheus 9090:9090"
echo "  Kibana:       kubectl port-forward svc/kibana 5601:5601"
echo "  Jaeger:       kubectl port-forward svc/jaeger 16686:16686"
echo ""
echo "🔐 Default Credentials:"
echo "  Grafana: admin / admin123"
echo ""
echo "📋 Useful Commands:"
echo "  Check all pods: kubectl get pods -l 'app in (grafana,prometheus,kibana,elasticsearch,logstash,jaeger)'"
echo "  View logs: kubectl logs -f deployment/<service-name>"
echo "  Delete stack: kubectl delete -f monitoring/"
echo ""
echo "⚠️  Note: Make sure your services are configured to send metrics and logs to these endpoints"