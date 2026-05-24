# Lightweight URL Shortener

Production-like URL shortener for a small Kubernetes cluster, designed for 2 VMs and about 8 GB RAM total.

## Architecture

- `shortener-api`: REST API for creating short links.
- `redirect-service`: fast redirect endpoint with Redis cache first, PostgreSQL fallback.
- `frontend`: React UI for creating short links.
- `PostgreSQL`: source of truth.
- `Redis`: hot redirect cache.
- `NATS`: optional async redirect events.

Both services are small Go binaries, expose `/healthz`, `/readyz`, and `/metrics`, and log structured JSON.

## Local Run

Requirements: Go 1.22+, Docker, Docker Compose.

```bash
cp .env.example .env
make compose-up
make migrate
```

Open the frontend:

```bash
http://localhost:3000
```

Create a short URL:

```bash
curl -s -X POST http://localhost:8080/api/shorten \
  -H 'content-type: application/json' \
  -d '{"url":"https://example.com/very/long/url"}'
```

Follow a redirect:

```bash
curl -i http://localhost:8081/abc123
```

## Development

```bash
make test
make build
make frontend-dev
make run-api
make run-redirect
```

Useful endpoints:

- API health: `http://localhost:8080/healthz`
- API readiness: `http://localhost:8080/readyz`
- API metrics: `http://localhost:8080/metrics`
- Redirect health: `http://localhost:8081/healthz`
- Redirect metrics: `http://localhost:8081/metrics`
- Frontend: `http://localhost:5173` with `make frontend-dev`

## Docker Compose

```bash
make compose-up
make compose-logs
make compose-down
```

Compose starts PostgreSQL, Redis, optional NATS, `shortener-api`, `redirect-service`, and `frontend`.
The frontend is available at `http://localhost:3000` and proxies `/api` to `shortener-api`.

Build only the frontend image:

```bash
docker build -f frontend/Dockerfile -t url-shortener-frontend:local .
docker run --rm -p 3000:80 --add-host=host.docker.internal:host-gateway \
  -e FRONTEND_API_UPSTREAM=http://host.docker.internal:8080 \
  url-shortener-frontend:local
```

## k3s Deployment

Build and push images to your registry:

```bash
export IMAGE_REGISTRY=registry.gitlab.com/dmxmss-group/dmxmss-project
make docker-build
make docker-push
```

Install with Helm:

```bash
helm upgrade --install url-shortener deploy/helm/url-shortener \
  --namespace url-shortener --create-namespace \
  --set image.registry=$IMAGE_REGISTRY \
  --set postgresql.auth.password=change-me
```

For a tiny home cluster, keep the default limits unless you know you have spare capacity. The chart defaults to conservative CPU and memory requests.

## kubectl Examples

```bash
kubectl -n url-shortener get pods
kubectl -n url-shortener get svc
kubectl -n url-shortener logs deploy/url-shortener-shortener-api
kubectl -n url-shortener logs deploy/url-shortener-redirect-service
kubectl -n url-shortener logs deploy/url-shortener-frontend
kubectl -n url-shortener rollout status deploy/url-shortener-shortener-api
```

Port-forward services:

```bash
kubectl -n url-shortener port-forward svc/url-shortener-shortener-api 8080:80
kubectl -n url-shortener port-forward svc/url-shortener-redirect-service 8081:80
kubectl -n url-shortener port-forward svc/url-shortener-frontend 3000:80
```

## Configuration

See `.env.example` for all variables.

Important variables:

- `DATABASE_URL`: PostgreSQL DSN.
- `REDIS_ADDR`: Redis host and port.
- `PUBLIC_BASE_URL`: base URL returned by `POST /api/shorten`.
- `NATS_URL`: optional; if set, redirect events are published to `redirect.events`.
- `FRONTEND_API_UPSTREAM`: backend upstream used by the frontend nginx proxy.

## Observability

- JSON logs via Go `log/slog`.
- Prometheus metrics at `/metrics`.
- Optional dashboard: `deploy/grafana/url-shortener-dashboard.json`.

## Repository Layout

```text
api/                  shortener-api Dockerfile and service entrypoint
redirect-service/     redirect-service Dockerfile and service entrypoint
frontend/             React frontend and nginx image
internal/             shared application packages
deploy/helm/          Helm chart for Kubernetes
deploy/pure_kubectl/  plain Kubernetes manifests
scripts/              helper scripts and SQL migrations
```
