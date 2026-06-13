# 01 — Architecture

## Data flow

```
              [browser / phone / Android TV]
                    |
                    | https://jellyfin.home.iamrpm.xyz
                    v
              [Traefik ingress + LE cert]
                    |
                    v
                [Jellyfin pod]   ── reads /media (HDD) ──┐
                                                         |
       /media/movies   /media/tv   /media/music          |
            ^               ^                             |
            |               |                             |
       writes from Radarr / Sonarr (hardlink-import)     |
            ^                                             |
            |                                             |
       /downloads/complete  <── qBittorrent writes here  |
            ^                                             |
            | .torrent files via HTTP API                 |
       Sonarr / Radarr                                    |
            ^                                             |
            | release search                              |
       Prowlarr  <──  MediaFusion + FlareSolverr          |
            |                                             |
            v                                             |
       [public indexers: 1337x, TGx, RARBG mirrors]      |
       [Indian: TamilMV, TamilBlasters via MediaFusion]  |
                                                          |
       Bazarr  ── scans /media, fetches .srt ─────────────┘
```

## Component roles

| Component | Role |
|-----------|------|
| **Prowlarr** | Single source for indexers. Syncs them to Sonarr/Radarr so trackers are configured once. |
| **FlareSolverr** | Headless Chromium proxy. Bypasses Cloudflare Turnstile on protected indexers (essential for TamilMV etc.). |
| **MediaFusion** | Community scraper exposing TamilMV / TamilBlasters / 1TamilBlasters via a Jackett/Prowlarr-compatible endpoint. |
| **Sonarr** | TV mgr. Monitors shows, hands releases to qBit, renames + moves into `/media/tv`. |
| **Radarr** | Same as Sonarr but for movies → `/media/movies`. |
| **Bazarr** | Watches Sonarr/Radarr libraries, fetches subtitles. |
| **qBittorrent** | Torrent client. Writes to `/downloads/complete` and `/downloads/incomplete`. |
| **Jellyfin** | Streaming frontend. Indexes `/media`, serves UI + apps. |

## Repo layout

```
homelab/
├── apps/                              # existing — backend-1, backend-2, vtu_app
├── docs/
│   └── media-stack/                   # this folder
├── k8s/
│   ├── _platform/                     # NEW — documents existing platform install
│   │   ├── argocd/
│   │   │   └── README.md
│   │   └── traefik/
│   │       └── README.md
│   ├── apps/
│   │   ├── backend-1/                 # existing
│   │   ├── backend-2/                 # existing
│   │   └── media/                     # NEW — entire media stack
│   │       ├── _shared/
│   │       │   ├── namespace.yaml
│   │       │   └── argocd-appset.yaml   # OR per-app application.yaml files
│   │       ├── prowlarr/
│   │       │   ├── application.yaml
│   │       │   └── values.yaml
│   │       ├── flaresolverr/
│   │       │   ├── application.yaml
│   │       │   └── values.yaml
│   │       ├── mediafusion/
│   │       ├── sonarr/
│   │       ├── radarr/
│   │       ├── bazarr/
│   │       ├── qbittorrent/
│   │       └── jellyfin/
│   └── o11y/                          # existing — prometheus, grafana
└── README.md
```

## Resource budget (8GB RAM node)

| Pod | Req RAM | Limit RAM | Req CPU | Limit CPU |
|-----|---------|-----------|---------|-----------|
| Jellyfin | 512Mi | 2Gi | 250m | 2000m |
| qBittorrent | 256Mi | 1Gi | 100m | 1000m |
| Sonarr | 256Mi | 768Mi | 50m | 500m |
| Radarr | 256Mi | 768Mi | 50m | 500m |
| Prowlarr | 128Mi | 384Mi | 50m | 300m |
| Bazarr | 256Mi | 512Mi | 50m | 500m |
| FlareSolverr | 256Mi | 512Mi | 50m | 500m |
| MediaFusion | 256Mi | 512Mi | 50m | 500m |
| **Total req** | ~2.2Gi | — | ~0.65 CPU | — |

~5-6 GB RAM left for OS + k3s + existing backends + buffers. 12GB swap as backstop.
