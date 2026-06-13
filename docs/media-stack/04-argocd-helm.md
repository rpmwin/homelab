# 04 — ArgoCD + Helm Pattern

## Stack

- **GitOps**: ArgoCD (already running in cluster).
- **Chart**: [`bjw-s/app-template`](https://bjw-s-labs.github.io/helm-charts/) — generic app chart.
- **Pattern**: one ArgoCD `Application` per media app, pointing at:
  - `chart=app-template` from bjw-s helm repo
  - per-app `values.yaml` in this repo (`k8s/apps/media/<app>/values.yaml`)

## Why bjw-s/app-template

- Generic chart — same chart for every app, only values differ. No bespoke charts to maintain.
- Active maintenance, 2025 community standard.
- Supports: Deployment/StatefulSet/DaemonSet, Service, Ingress, PVC, ConfigMap, Secret, ServiceMonitor, init containers, sidecars, persistence with hostPath, custom RBAC.
- Alternatives rejected: Truecharts (paid SCALE-focused since 2024), k8s-at-home (archived), per-app charts (abandoned).

## ApplicationSet vs per-app Application

**Recommend per-app `Application` files** for clarity. ApplicationSet is optional if you want zero-touch on adding apps.

### Per-app Application (recommended)

```yaml
# k8s/apps/media/sonarr/application.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: sonarr
  namespace: argocd
spec:
  project: default
  sources:
    - repoURL: https://bjw-s-labs.github.io/helm-charts
      chart: app-template
      targetRevision: 3.6.1   # pin — check latest stable before applying
      helm:
        releaseName: sonarr
        valueFiles:
          - $values/k8s/apps/media/sonarr/values.yaml
    - repoURL: https://github.com/rpmwin/homelab.git   # update to actual repo URL
      targetRevision: main
      ref: values
  destination:
    server: https://kubernetes.default.svc
    namespace: media
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - ServerSideApply=true
```

This is the **multi-source** pattern (`sources:` array). It lets `values.yaml` live in *your* repo while the chart comes from bjw-s — clean separation.

### Optional: ApplicationSet

If you want a single file that generates an Application for every subdir under `k8s/apps/media/`:

```yaml
# k8s/apps/media/_shared/argocd-appset.yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: media-stack
  namespace: argocd
spec:
  generators:
    - git:
        repoURL: https://github.com/rpmwin/homelab.git
        revision: main
        directories:
          - path: k8s/apps/media/*
        # exclude _shared
        # appset filter happens via template
  template:
    metadata:
      name: '{{path.basename}}'
    spec:
      project: default
      sources:
        - repoURL: https://bjw-s-labs.github.io/helm-charts
          chart: app-template
          targetRevision: 3.6.1
          helm:
            releaseName: '{{path.basename}}'
            valueFiles:
              - $values/{{path}}/values.yaml
        - repoURL: https://github.com/rpmwin/homelab.git
          targetRevision: main
          ref: values
      destination:
        server: https://kubernetes.default.svc
        namespace: media
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
        syncOptions:
          - CreateNamespace=true
```

Pick one approach; don't mix.

## Namespace

```yaml
# k8s/apps/media/_shared/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: media
  labels:
    pod-security.kubernetes.io/enforce: privileged   # only needed if Jellyfin uses /dev/dri
```

(`privileged` label is for Jellyfin's `/dev/dri` mount + render group; it does NOT mean privileged containers — it's the PSA label.)

## Values file structure (preview — full per-app values in `05-apps.md`)

```yaml
# k8s/apps/media/<app>/values.yaml — bjw-s/app-template v3
controllers:
  main:
    type: deployment
    replicas: 1
    strategy: Recreate     # required for RWO PVCs + hostPath shared paths
    containers:
      main:
        image:
          repository: lscr.io/linuxserver/<app>
          tag: latest
        env:
          PUID: "1000"
          PGID: "1000"
          TZ: Asia/Kolkata
        resources:
          requests:
            cpu: 50m
            memory: 256Mi
          limits:
            memory: 768Mi
service:
  main:
    controller: main
    ports:
      http:
        port: <port>
ingress:
  main:
    enabled: true
    className: traefik
    annotations:
      traefik.ingress.kubernetes.io/router.tls.certresolver: letsencrypt
      traefik.ingress.kubernetes.io/router.entrypoints: websecure
    hosts:
      - host: <app>.home.iamrpm.xyz
        paths:
          - path: /
            service:
              identifier: main
              port: http
    tls:
      - hosts:
          - <app>.home.iamrpm.xyz
persistence:
  config:
    type: hostPath
    hostPath: /opt/media/media-config/<app>
    globalMounts:
      - path: /config
  media:
    type: hostPath
    hostPath: /data/media
    globalMounts:
      - path: /media
  downloads:
    type: hostPath
    hostPath: /data/downloads
    globalMounts:
      - path: /downloads
```

Full per-app values are in [`05-apps.md`](05-apps.md).

## Sync ordering

ArgoCD sync waves (`argocd.argoproj.io/sync-wave` annotation) for clean bootstrap:

| Wave | Apps |
|------|------|
| -1 | `_shared` (namespace) |
| 0 | flaresolverr, mediafusion |
| 1 | prowlarr |
| 2 | qbittorrent |
| 3 | sonarr, radarr |
| 4 | bazarr, jellyfin |

Add to each Application's metadata:
```yaml
metadata:
  name: sonarr
  namespace: argocd
  annotations:
    argocd.argoproj.io/sync-wave: "3"
```
