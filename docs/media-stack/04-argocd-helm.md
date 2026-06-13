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

## Pattern: ApplicationSet (chosen)

**One ApplicationSet generates all media app Applications.** Single resource to apply; new apps auto-register when a new subdir is added to the repo.

File: `k8s/_platform/argocd/media-applicationset.yaml`

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: media-stack
  namespace: argocd
spec:
  goTemplate: false           # use fasttemplate syntax {{path.basename}}
  generators:
    - git:
        repoURL: git@github.com:rpmwin/homelab.git
        revision: main
        directories:
          - path: k8s/apps/media/*
          - path: k8s/apps/media/_shared
            exclude: true     # skip the _shared dir
  template:
    metadata:
      name: '{{path.basename}}'       # → prowlarr, sonarr, radarr, ...
      namespace: argocd
      labels:
        app.kubernetes.io/part-of: media-stack
    spec:
      project: default
      sources:
        - repoURL: https://bjw-s-labs.github.io/helm-charts
          chart: app-template
          targetRevision: 3.6.1
          helm:
            releaseName: '{{path.basename}}'
            valueFiles:
              - $values/{{path}}/values.yaml     # k8s/apps/media/<app>/values.yaml
        - repoURL: git@github.com:rpmwin/homelab.git
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

### How it works

1. ApplicationSet controller polls the repo (~3 min).
2. The **git directories** generator enumerates subdirs matching `k8s/apps/media/*` (excluding `_shared`). For each found dir, generates one template parameter set with `{{path}}` and `{{path.basename}}`.
3. The template is rendered once per dir → produces 7 Application CRs.
4. ArgoCD reconciles each Application as if it were applied manually — pulls bjw-s/app-template + `values.yaml` from this repo, helm-templates, applies to `media` ns.

### Adding / removing apps

- **Add**: `mkdir k8s/apps/media/lidarr && echo "..." > k8s/apps/media/lidarr/values.yaml && git push` → ApplicationSet auto-creates `lidarr` Application within ~3 min.
- **Remove**: `git rm -r k8s/apps/media/<app> && git push` → ApplicationSet removes the Application; since `prune: true`, all workloads are deleted from cluster too.

### Why no sync waves?

The apps in this stack don't have **runtime** dependencies — only **configuration** dependencies wired via UI later (e.g., Sonarr UI points at Prowlarr URL). All 7 pods can start concurrently. Some Liveness probe failures during initial boot are normal and self-resolve.

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
