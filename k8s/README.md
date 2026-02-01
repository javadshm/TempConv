# Kubernetes manifests for GKE

Before applying, replace `REGISTRY` in:

- `backend-deployment.yaml` (image: REGISTRY/backend:latest)
- `frontend-deployment.yaml` (image: REGISTRY/frontend:latest)

with your full registry path, e.g.:

- `europe-west1-docker.pkg.dev/YOUR_PROJECT_ID/tempconv`

**One-time substitution (PowerShell):**

```powershell
$REGISTRY = "europe-west1-docker.pkg.dev/YOUR_PROJECT_ID/tempconv"
(Get-Content backend-deployment.yaml) -replace 'REGISTRY', $REGISTRY | Set-Content backend-deployment.yaml
(Get-Content frontend-deployment.yaml) -replace 'REGISTRY', $REGISTRY | Set-Content frontend-deployment.yaml
```

**Apply order:**

```bash
kubectl apply -f namespace.yaml
kubectl apply -f backend-deployment.yaml -f backend-service.yaml
kubectl apply -f frontend-deployment.yaml -f frontend-service.yaml
kubectl apply -f ingress.yaml
```
