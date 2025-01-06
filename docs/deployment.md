# Deployment Documentation

This document describes the deployment process for the Fowergram API.

## Prerequisites

- Google Cloud Platform (GCP) account
- Google Kubernetes Engine (GKE) cluster
- Docker
- kubectl configured for your GKE cluster
- Access to Container Registry

## GitHub Secrets Setup

Before deploying, you need to set up the following secrets in your GitHub repository:

1. Go to your GitHub repository
2. Navigate to Settings > Secrets and variables > Actions
3. Add the following secrets:

| Secret Name | Description |
|------------|-------------|
| GCP_PROJECT_ID | Your Google Cloud Project ID |
| GCP_SA_KEY | Base64-encoded service account key JSON |
| GKE_CLUSTER | Your GKE cluster name |
| GKE_ZONE | Your GKE cluster zone |

To get the service account key:
```bash
# Create service account
gcloud iam service-accounts create github-actions --display-name="GitHub Actions"

# Add necessary roles
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/container.developer"

# Create and download key
gcloud iam service-accounts keys create key.json \
    --iam-account=github-actions@$PROJECT_ID.iam.gserviceaccount.com

# Encode key for GitHub secrets
cat key.json | base64
```

## Environment Setup

1. Set environment variables:
```bash
export PROJECT_ID=your-project-id
export GKE_CLUSTER=your-cluster-name
export REGION=your-region
export IMAGE=gcr.io/$PROJECT_ID/fowergram-api
```

2. Configure kubectl:
```bash
gcloud container clusters get-credentials $GKE_CLUSTER --region $REGION --project $PROJECT_ID
```

## Deployment Process

1. Build and push Docker image:
```bash
docker build -t $IMAGE .
docker push $IMAGE
```

2. Apply Kubernetes configurations:
```bash
# Apply ConfigMap
kubectl apply -f k8s/base/configmap.yaml

# Apply Secrets
kubectl apply -f k8s/base/secrets.yaml

# Apply Deployment
kubectl apply -f k8s/base/deployment.yaml

# Apply Service
kubectl apply -f k8s/base/service.yaml
```

3. Verify deployment:
```bash
# Check deployment status
kubectl rollout status deployment/fowergram-api

# Check pods
kubectl get pods -l app=fowergram-api

# Check service
kubectl get service fowergram-api
```

## Deployment Configuration

### Resource Limits

The deployment is configured with the following resource limits:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi
```

### Health Checks

The deployment includes both readiness and liveness probes:

```yaml
readinessProbe:
  httpGet:
    path: /ping
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

livenessProbe:
  httpGet:
    path: /ping
    port: 8080
  initialDelaySeconds: 15
  periodSeconds: 20
```

### Rolling Updates

The deployment uses a rolling update strategy:

```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1
    maxUnavailable: 0
```

## Monitoring

1. Check application logs:
```bash
kubectl logs -l app=fowergram-api
```

2. Monitor pod status:
```bash
kubectl get pods -l app=fowergram-api -w
```

3. Check service endpoints:
```bash
kubectl get endpoints fowergram-api
```

## Troubleshooting

1. Pod not starting:
   - Check pod events: `kubectl describe pod <pod-name>`
   - Check logs: `kubectl logs <pod-name>`
   - Verify resource limits
   - Check image pull status

2. Database connection issues:
   - Verify secrets are correctly set
   - Check Cloud SQL proxy status
   - Test connection using temporary pod

3. Service not accessible:
   - Check service configuration
   - Verify endpoints
   - Check firewall rules

## Rollback

To rollback to a previous version:

```bash
# Get deployment history
kubectl rollout history deployment/fowergram-api

# Rollback to previous version
kubectl rollout undo deployment/fowergram-api

# Rollback to specific revision
kubectl rollout undo deployment/fowergram-api --to-revision=<revision-number>
```

## Security Considerations

1. Secrets Management
   - Use Kubernetes Secrets for sensitive data
   - Rotate credentials regularly
   - Use Secret Manager for production

2. Network Security
   - Configure Network Policies
   - Use TLS for all external traffic
   - Implement proper RBAC

3. Container Security
   - Use minimal base images
   - Scan for vulnerabilities
   - Run as non-root user

## Best Practices

1. Always tag images with specific versions
2. Use resource limits and requests
3. Implement proper health checks
4. Use rolling updates for zero-downtime deployments
5. Monitor application metrics and logs
6. Keep secrets and configurations separate
7. Document all deployment procedures 