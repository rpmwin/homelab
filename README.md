# homelab

Personal homelab running on k3s. Two backend services with full observability stack.

## Structure

```
homelab/
├── apps/
│   ├── backend-1/   # Node.js + TypeScript + Express
│   └── backend-2/   # Go + net/http
└── k8s/
    ├── apps/
    │   ├── backend-1/   # Deployment, Service, Ingress, HPA, ConfigMap
    │   └── backend-2/   # Deployment, Service, Ingress, ConfigMap
    └── o11y/            # Grafana ingress
```

## Stack

- **Runtime**: k3s
- **Ingress**: Traefik
- **Observability**: Prometheus + Grafana (kube-prometheus-stack via ArgoCD)
- **Logging**: Loki + Promtail
- **GitOps**: ArgoCD
- **Registry**: Docker Hub

## Services

| Service | Language | Port | URL |
|---------|----------|------|-----|
| backend-1 | Node/TS | 3000 | https://back1.home.iamrpm.xyz |
| backend-2 | Go | 8080 | https://back2.home.iamrpm.xyz |

Both services expose `/metrics` for Prometheus scraping.

## Docs

- [`docs/media-stack/`](docs/media-stack/README.md) — *arr + Jellyfin media stack plan (not yet deployed). Step-by-step rollout reference for executing the build.
