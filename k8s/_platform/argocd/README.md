# ArgoCD

## Install

Installed via Helm into the `argocd` namespace ~16 days before this file was written. Not yet tracked as a manifest in this repo.

```bash
# To verify current install:
kubectl -n argocd get deploy
kubectl -n argocd get svc argocd-server
```

## Accessing the UI

ArgoCD server is in the cluster. To access from outside:

```bash
# Port-forward (temporary)
kubectl port-forward svc/argocd-server -n argocd 8080:443
# Open: https://localhost:8080

# Or via ingress if configured (check: kubectl -n argocd get ingress)
```

## Repo credentials

The homelab repo (`git@github.com:rpmwin/homelab.git`) is registered as a repository in ArgoCD with SSH key. All Application manifests reference this repo.

To view:
```bash
kubectl -n argocd get secret -l argocd.argoproj.io/secret-type=repository
```

## Current applications

| App | Namespace | Source |
|-----|-----------|--------|
| backend-1 | default | k8s/apps/backend-1/ |
| backend-2 | default | k8s/apps/backend-2/ |
| kube-prom-stack | monitoring | kube-prometheus-stack Helm chart |
| o11y | monitoring | k8s/o11y/ |
| media apps (WIP) | media | k8s/apps/media/<app>/application.yaml |

## App structure (media stack pattern)

Each media app uses multi-source Applications:
- Source 1: bjw-s/app-template Helm chart (`https://bjw-s-labs.github.io/helm-charts`)
- Source 2: this repo as `ref: values`, providing per-app `values.yaml`

Apply a new app:
```bash
kubectl apply -f k8s/apps/media/<app>/application.yaml
```

## Helm repo

Add to ArgoCD if not already present:
```bash
argocd repo add https://bjw-s-labs.github.io/helm-charts --type helm --name bjw-s
```
