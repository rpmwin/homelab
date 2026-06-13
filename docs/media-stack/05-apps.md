# 05 — Per-App Values

Full `values.yaml` for each app. Replace `/opt/media` and `/data` with actual mount points.

> **Chart version**: bjw-s/app-template `3.6.1` (or latest stable). Check syntax against current chart docs before applying — v3 API differs from v1/v2.

---

## Prowlarr

**Image**: `lscr.io/linuxserver/prowlarr:latest`  •  **Port**: 9696  •  **URL**: `prowlarr.home.iamrpm.xyz`

```yaml
# k8s/apps/media/prowlarr/values.yaml
controllers:
  main:
    type: deployment
    replicas: 1
    strategy: Recreate
    containers:
      main:
        image:
          repository: lscr.io/linuxserver/prowlarr
          tag: latest
        env:
          PUID: "1000"
          PGID: "1000"
          TZ: Asia/Kolkata
        resources:
          requests: { cpu: 50m, memory: 128Mi }
          limits:   { memory: 384Mi }
service:
  main:
    controller: main
    ports:
      http: { port: 9696 }
ingress:
  main:
    enabled: true
    className: traefik
    annotations:
      traefik.ingress.kubernetes.io/router.tls.certresolver: letsencrypt
      traefik.ingress.kubernetes.io/router.entrypoints: websecure
    hosts:
      - host: prowlarr.home.iamrpm.xyz
        paths: [{ path: /, service: { identifier: main, port: http } }]
    tls: [{ hosts: [prowlarr.home.iamrpm.xyz] }]
persistence:
  config:
    type: hostPath
    hostPath: /opt/media/media-config/prowlarr
    globalMounts: [{ path: /config }]
```

---

## FlareSolverr

**Image**: `ghcr.io/flaresolverr/flaresolverr:latest`  •  **Port**: 8191  •  **No ingress** (ClusterIP only — internal)

```yaml
# k8s/apps/media/flaresolverr/values.yaml
controllers:
  main:
    type: deployment
    replicas: 1
    containers:
      main:
        image:
          repository: ghcr.io/flaresolverr/flaresolverr
          tag: latest
        env:
          LOG_LEVEL: info
          TZ: Asia/Kolkata
        resources:
          requests: { cpu: 50m, memory: 256Mi }
          limits:   { memory: 512Mi }
service:
  main:
    controller: main
    ports:
      http: { port: 8191 }
# Cluster DNS: flaresolverr.media.svc.cluster.local:8191
# Configure in Prowlarr: Settings → Indexers → Add → FlareSolverr → http://flaresolverr.media.svc.cluster.local:8191
```

---

## MediaFusion

**Image**: `mhdzumair/mediafusion:latest`  •  **Port**: 8000  •  **No ingress** (internal)

> **Note**: MediaFusion has dependencies (MongoDB + Redis). Easiest path: use the [official docker-compose](https://github.com/mhdzumair/MediaFusion/blob/main/docker-compose.yml) as a reference and translate it into bjw-s values with multiple containers (main + mongo + redis sidecars) OR run a 3-resource bundle (separate Deployments). For first cut, **defer MediaFusion and use only Prowlarr defaults + 1337x + TGx**, then add MediaFusion in a follow-up.

Skeleton (incomplete — flesh out during exec):

```yaml
# k8s/apps/media/mediafusion/values.yaml
controllers:
  main:
    type: deployment
    replicas: 1
    containers:
      main:
        image:
          repository: mhdzumair/mediafusion
          tag: latest
        env:
          MONGO_URI: mongodb://mediafusion-mongo:27017
          REDIS_URL: redis://mediafusion-redis:6379
          # ... see MediaFusion README for full env
        resources:
          requests: { cpu: 50m, memory: 256Mi }
          limits:   { memory: 512Mi }
service:
  main:
    controller: main
    ports:
      http: { port: 8000 }
# Plus separate Deployments/Services for mongo + redis,
# OR use a Helm chart that bundles them.
```

---

## Sonarr

**Image**: `lscr.io/linuxserver/sonarr:latest`  •  **Port**: 8989  •  **URL**: `sonarr.home.iamrpm.xyz`

```yaml
# k8s/apps/media/sonarr/values.yaml
controllers:
  main:
    type: deployment
    replicas: 1
    strategy: Recreate
    containers:
      main:
        image:
          repository: lscr.io/linuxserver/sonarr
          tag: latest
        env:
          PUID: "1000"
          PGID: "1000"
          TZ: Asia/Kolkata
        resources:
          requests: { cpu: 50m, memory: 256Mi }
          limits:   { memory: 768Mi }
service:
  main:
    controller: main
    ports:
      http: { port: 8989 }
ingress:
  main:
    enabled: true
    className: traefik
    annotations:
      traefik.ingress.kubernetes.io/router.tls.certresolver: letsencrypt
      traefik.ingress.kubernetes.io/router.entrypoints: websecure
    hosts:
      - host: sonarr.home.iamrpm.xyz
        paths: [{ path: /, service: { identifier: main, port: http } }]
    tls: [{ hosts: [sonarr.home.iamrpm.xyz] }]
persistence:
  config:
    type: hostPath
    hostPath: /opt/media/media-config/sonarr
    globalMounts: [{ path: /config }]
  media:
    type: hostPath
    hostPath: /data/media
    globalMounts: [{ path: /media }]
  downloads:
    type: hostPath
    hostPath: /data/downloads
    globalMounts: [{ path: /downloads }]
```

---

## Radarr

Identical to Sonarr — substitute `sonarr` → `radarr`, port `8989` → `7878`.

---

## Bazarr

**Image**: `lscr.io/linuxserver/bazarr:latest`  •  **Port**: 6767  •  **URL**: `bazarr.home.iamrpm.xyz`

```yaml
# k8s/apps/media/bazarr/values.yaml
controllers:
  main:
    type: deployment
    replicas: 1
    strategy: Recreate
    containers:
      main:
        image:
          repository: lscr.io/linuxserver/bazarr
          tag: latest
        env:
          PUID: "1000"
          PGID: "1000"
          TZ: Asia/Kolkata
        resources:
          requests: { cpu: 50m, memory: 256Mi }
          limits:   { memory: 512Mi }
service:
  main:
    controller: main
    ports:
      http: { port: 6767 }
ingress:
  main:
    enabled: true
    className: traefik
    annotations:
      traefik.ingress.kubernetes.io/router.tls.certresolver: letsencrypt
      traefik.ingress.kubernetes.io/router.entrypoints: websecure
    hosts:
      - host: bazarr.home.iamrpm.xyz
        paths: [{ path: /, service: { identifier: main, port: http } }]
    tls: [{ hosts: [bazarr.home.iamrpm.xyz] }]
persistence:
  config:
    type: hostPath
    hostPath: /opt/media/media-config/bazarr
    globalMounts: [{ path: /config }]
  media:
    type: hostPath
    hostPath: /data/media
    globalMounts: [{ path: /media }]
```

Note: Bazarr only needs `/media` (read+write for subtitle files), not `/downloads`.

---

## qBittorrent

**Image**: `ghcr.io/hotio/qbittorrent:release-5.0.4` (pin to a release tag)  •  **WebUI port**: 8080  •  **Torrent port**: 6881  •  **URL**: `qbit.home.iamrpm.xyz`

```yaml
# k8s/apps/media/qbittorrent/values.yaml
controllers:
  main:
    type: deployment
    replicas: 1
    strategy: Recreate
    containers:
      main:
        image:
          repository: ghcr.io/hotio/qbittorrent
          tag: release-5.0.4
        env:
          PUID: "1000"
          PGID: "1000"
          TZ: Asia/Kolkata
          WEBUI_PORTS: "8080/tcp,8080/udp"
        resources:
          requests: { cpu: 100m, memory: 256Mi }
          limits:   { memory: 1Gi }
service:
  main:
    controller: main
    ports:
      http:   { port: 8080 }
      bt-tcp: { port: 6881, protocol: TCP }
      bt-udp: { port: 6881, protocol: UDP }
  # Expose 6881 via LoadBalancer/NodePort if you need inbound peer connections
  # from outside your NAT — otherwise outbound-only is fine.
ingress:
  main:
    enabled: true
    className: traefik
    annotations:
      traefik.ingress.kubernetes.io/router.tls.certresolver: letsencrypt
      traefik.ingress.kubernetes.io/router.entrypoints: websecure
    hosts:
      - host: qbit.home.iamrpm.xyz
        paths: [{ path: /, service: { identifier: main, port: http } }]
    tls: [{ hosts: [qbit.home.iamrpm.xyz] }]
persistence:
  config:
    type: hostPath
    hostPath: /opt/media/media-config/qbittorrent
    globalMounts: [{ path: /config }]
  downloads:
    type: hostPath
    hostPath: /data/downloads
    globalMounts: [{ path: /downloads }]
  incomplete:
    type: hostPath
    hostPath: /opt/media/downloads-incomplete
    globalMounts: [{ path: /downloads/incomplete }]
```

**qBit UI config (after first login)**:

- Default creds: hotio image generates a temp password — check pod logs:
  `kubectl -n media logs deploy/qbittorrent | grep -i password`
- Settings → Web UI → change password.
- Settings → Downloads:
  - Default save path: `/downloads/complete`
  - Keep incomplete torrents in: `/downloads/incomplete`
  - Append `.!qB` extension: ON
  - Run external program on completion: leave empty (Sonarr/Radarr handle imports via API)
- Settings → BitTorrent → Enable DHT, PeX, LSD.
- Categories (matters for *arr import paths):
  - `tv` → save path `/downloads/complete/tv`
  - `movies` → save path `/downloads/complete/movies`
  - `music` → save path `/downloads/complete/music`

---

## Jellyfin

**Image**: `lscr.io/linuxserver/jellyfin:latest`  •  **Port**: 8096  •  **URL**: `jellyfin.home.iamrpm.xyz`

```yaml
# k8s/apps/media/jellyfin/values.yaml
controllers:
  main:
    type: deployment
    replicas: 1
    strategy: Recreate
    containers:
      main:
        image:
          repository: lscr.io/linuxserver/jellyfin
          tag: latest
        env:
          PUID: "1000"
          PGID: "1000"
          TZ: Asia/Kolkata
          JELLYFIN_PublishedServerUrl: https://jellyfin.home.iamrpm.xyz
        resources:
          requests:
            cpu: 250m
            memory: 512Mi
          limits:
            memory: 2Gi
            # Uncomment if Intel device plugin is installed AND QSV is desired:
            # gpu.intel.com/i915: 1
        # Uncomment for QSV:
        # securityContext:
        #   supplementalGroups: [105]   # 'render' group GID — verify on node: getent group render
service:
  main:
    controller: main
    ports:
      http: { port: 8096 }
ingress:
  main:
    enabled: true
    className: traefik
    annotations:
      traefik.ingress.kubernetes.io/router.tls.certresolver: letsencrypt
      traefik.ingress.kubernetes.io/router.entrypoints: websecure
    hosts:
      - host: jellyfin.home.iamrpm.xyz
        paths: [{ path: /, service: { identifier: main, port: http } }]
    tls: [{ hosts: [jellyfin.home.iamrpm.xyz] }]
persistence:
  config:
    type: hostPath
    hostPath: /opt/media/media-config/jellyfin
    globalMounts: [{ path: /config }]
  cache:
    type: hostPath
    hostPath: /opt/media/jellyfin-cache
    globalMounts: [{ path: /cache }]
  media:
    type: hostPath
    hostPath: /data/media
    globalMounts:
      - path: /media
        readOnly: true   # Jellyfin only reads; *arr writes
  # Uncomment for QSV — mount the iGPU device node:
  # dri:
  #   type: hostPath
  #   hostPath: /dev/dri
  #   globalMounts: [{ path: /dev/dri }]
```

**Jellyfin first-time setup**:

1. Open `https://jellyfin.home.iamrpm.xyz` → onboarding wizard → create admin user.
2. Add libraries:
   - TV Shows → `/media/tv` → metadata source TheTVDB
   - Movies → `/media/movies` → metadata source TheMovieDB
3. Dashboard → Playback (if QSV):
   - Hardware acceleration: Intel QuickSync (QSV)
   - Enable HW decoding for: H.264, HEVC, VP9 (skip AV1 on Gen 6-10)
   - Enable HW encoding: ON
   - Tone mapping: ON (only useful for HDR → SDR; OK to leave on)
   - Transcoding thread count: 2
4. Dashboard → Networking:
   - Published server URL: `https://jellyfin.home.iamrpm.xyz`
   - Known proxies: add the cluster's pod CIDR (so X-Forwarded-For works)
