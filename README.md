# TempConv – Temperature conversion app (Go gRPC + Flutter, Docker, GKE)

Simple app: **backend** (Go) exposes a gRPC service with grpc-gateway providing an HTTP API at `/api/convert` for temperature conversion; **frontend** (Flutter web) calls it. No database. Containerized with Docker and orchestrated on **Google Kubernetes Engine (GKE)**. Load tested with **k6**.

---

## Project layout

```
TempConv/
├── backend/           # Go gRPC + grpc-gateway API (C↔F conversion)
│   ├── api/
│   │   └── tempconv.proto  # Protobuf service definition
│   ├── gen/           # Generated protobuf Go code
│   ├── cmd/server/
│   │   └── main.go    # Server implementation
│   ├── go.mod
│   ├── generate.sh    # Code generation script
│   ├── CODEGEN.md     # Code generation documentation
│   ├── Dockerfile
│   └── .dockerignore
├── frontend/          # Flutter web app
│   ├── lib/main.dart
│   ├── web/
│   ├── pubspec.yaml
│   ├── nginx.conf
│   └── Dockerfile
├── k8s/               # Kubernetes manifests for GKE
│   ├── namespace.yaml
│   ├── backend-deployment.yaml
│   ├── backend-service.yaml
│   ├── frontend-deployment.yaml
│   ├── frontend-service.yaml
│   └── ingress.yaml
├── loadtest/          # k6 load tests
│   └── k6-load.js
├── docker-compose.yaml
└── README.md
```

---

## Step 1 – Backend (Go with gRPC + grpc-gateway)

**What we do:** Implement a gRPC service with HTTP gateway for temperature conversion using a single unified endpoint.

- **Architecture:**
  - gRPC server on port 9090 (internal)
  - grpc-gateway HTTP proxy on port 8080 (exposed)
  - Single conversion endpoint: `POST /api/convert`
  
- **Endpoints:**
  - `POST /api/convert` – body `{"value": <number>, "from_unit": "CELSIUS|FAHRENHEIT", "to_unit": "CELSIUS|FAHRENHEIT"}` → `{"value": <converted>}`
  - `GET /health` – returns `{"status":"ok"}` for probes

**Run locally:**

```bash
cd backend
go run ./cmd/server
```

**Tests:**

```bash
cd backend
go test -v .
```

**Test with curl:**

```bash
# Convert 0°C to Fahrenheit
curl -X POST http://localhost:8080/api/convert \
  -H "Content-Type: application/json" \
  -d '{"value": 0, "from_unit": "CELSIUS", "to_unit": "FAHRENHEIT"}'

# Convert 32°F to Celsius
curl -X POST http://localhost:8080/api/convert \
  -H "Content-Type: application/json" \
  -d '{"value": 32, "from_unit": "FAHRENHEIT", "to_unit": "CELSIUS"}'
```

**Code Generation:**

The backend uses protobuf for service definitions. To regenerate the Go code from `api/tempconv.proto`:

```bash
cd backend
./generate.sh
```

See `backend/CODEGEN.md` for more details on code generation.

---

## Step 2 – Frontend (Flutter web)

**What we do:** A small Flutter web app that sends a temperature to the backend and shows the result. It uses the same origin when served behind the K8s Ingress (one host for app and API).

**Run locally (needs backend on :8080 or set API_BASE):**

```bash
cd frontend
flutter pub get
# If backend is on another origin (e.g. localhost:8080):
# flutter run -d chrome --dart-define=API_BASE=http://localhost:8080
flutter run -d chrome
```

**Build for web (static files in `build/web`):**

```bash
cd frontend
flutter build web
```

If you created the project manually, ensure web support exists:

```bash
flutter create . --platforms web
```

---

## Step 3 – Containerization (Docker, amd64 for GKE)

**What we do:** Build images for **linux/amd64** so they run on GKE nodes. Backend: multi-stage Go build with protobuf generation. Frontend: Flutter web build + nginx to serve static files.

**Build images (from repo root):**

```bash
# Backend (builds with protobuf generation)
docker build --platform linux/amd64 -t tempconv-backend:latest ./backend

# Frontend (no API_BASE = same-origin, for K8s)
docker build --platform linux/amd64 -t tempconv-frontend:latest ./frontend
```

**Run with Docker Compose (backend :8080, frontend :8081; frontend is built with API_BASE so browser can call backend):**

```bash
docker compose up --build
# Open http://localhost:8081
```

**Note:** The backend Dockerfile automatically generates protobuf code during the build process, so no pre-generation is needed.

---

## Step 4 – Kubernetes on GKE

**What we do:** Deploy the app to GKE using the manifests in `k8s/`. One Ingress serves the Flutter app at `/` and proxies `/api` to the Go backend, so the frontend can use the same origin (no CORS/API_BASE needed in K8s).

**Prerequisites:** `gcloud`, `kubectl`, `gke-gcloud-auth-plugin`; GKE cluster created and `kubectl` configured.

**4.1 Create cluster (if needed):**

```bash
export PROJECT_ID=your-gcp-project-id
export REGION=europe-west1
export CLUSTER_NAME=tempconv-cluster

gcloud container clusters create $CLUSTER_NAME \
  --project=$PROJECT_ID \
  --region=$REGION \
  --num-nodes=2 \
  --machine-type=e2-small \
  --enable-autoscaling --min-nodes=1 --max-nodes=4

gcloud container clusters get-credentials $CLUSTER_NAME --region=$REGION --project=$PROJECT_ID
```

**4.2 Build for amd64 and push to Google Artifact Registry (or GCR):**

```bash
export REGION=europe-west1
export REGISTRY=$REGION-docker.pkg.dev/$PROJECT_ID/tempconv

# Create repo (once)
gcloud artifacts repositories create tempconv --repository-format=docker --location=$REGION --project=$PROJECT_ID

# Build and push (from repo root)
# Note: Backend build includes automatic protobuf code generation
docker build --platform linux/amd64 -t $REGISTRY/backend:latest ./backend
docker build --platform linux/amd64 -t $REGISTRY/frontend:latest ./frontend
docker push $REGISTRY/backend:latest
docker push $REGISTRY/frontend:latest
```

**4.3 Substitute `REGISTRY` in manifests and apply:**

Replace `REGISTRY` in `k8s/backend-deployment.yaml` and `k8s/frontend-deployment.yaml` with your registry (e.g. `europe-west1-docker.pkg.dev/PROJECT_ID/tempconv`), then:

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/backend-deployment.yaml
kubectl apply -f k8s/backend-service.yaml
kubectl apply -f k8s/frontend-deployment.yaml
kubectl apply -f k8s/frontend-service.yaml
kubectl apply -f k8s/ingress.yaml
```

**4.4 Wait for Ingress IP:**

```bash
kubectl get ingress -n tempconv -w
# Use the ADDRESS when it appears.
```

Then open `http://<ADDRESS>` in the browser.

---

## Step 5 – Load testing (k6)

**What we do:** Simulate many “frontends” (virtual users) calling the backend to verify it under load.

**Local (backend on :8080):**

```bash
cd loadtest
k6 run k6-load.js
```

**Against GKE (use the Ingress IP; requests go to the same host, so `/api` hits the backend):**

```bash
BASE_URL=http://<INGRESS_IP> k6 run loadtest/k6-load.js
# Or with more load:
k6 run --vus 50 --duration 120s loadtest/k6-load.js
BASE_URL=http://<INGRESS_IP> k6 run --vus 100 --duration 120s loadtest/k6-load.js
```

k6 reports success rate, latency (e.g. p95), and RPS. Adjust `replicas` for the backend Deployment if you need more capacity.

---

## Summary

| Step | What | Command / location |
|------|------|--------------------|
| 1 | Go backend (gRPC + HTTP gateway) | `backend/`, `go run ./cmd/server` / `go test .` |
| 2 | Flutter web frontend | `frontend/`, `flutter run -d chrome` |
| 3 | Docker (amd64) | `docker build --platform linux/amd64` for both; `docker compose up` |
| 4 | GKE | Push to Artifact Registry, `kubectl apply -f k8s/` |
| 5 | Load test | `k6 run loadtest/k6-load.js` (optionally with `BASE_URL`) |

**Backend Architecture:**
- gRPC service on port 9090 (internal communication)
- grpc-gateway HTTP proxy on port 8080 (external HTTP/JSON API)
- Single conversion endpoint: `/api/convert` with flexible unit specification
- Protobuf-based service definitions with automatic code generation

GKE nodes are **amd64**; all images in this guide are built with `--platform linux/amd64` so they run correctly on GKE.
