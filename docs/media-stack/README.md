# Media Stack — Reference

Self-hosted *arr + Jellyfin media stack on this k3s homelab. Streaming-first.

## Files in this folder

| File | Purpose |
|------|---------|
| [`README.md`](README.md) | You are here. High-level reference. |
| [`01-architecture.md`](01-architecture.md) | Components, data flow, repo layout. |
| [`02-prerequisites.md`](02-prerequisites.md) | Node probe, DNS, swap bump, dirs. |
| [`03-storage.md`](03-storage.md) | SSD vs HDD layout + hardlinks. |
| [`04-argocd-helm.md`](04-argocd-helm.md) | ArgoCD + bjw-s/app-template pattern. |
| [`05-apps.md`](05-apps.md) | Per-app values (image, ports, env, mounts, ingress). |
| [`06-rollout.md`](06-rollout.md) | Step-by-step execution order. |
| [`07-verification.md`](07-verification.md) | End-to-end test checklist. |
| [`08-legal-notes.md`](08-legal-notes.md) | India torrent legal context. |
| [`09-troubleshooting.md`](09-troubleshooting.md) | Real issues hit during rollout + fixes (DNS, LE certs, Cloudflare). |
| [`10-app-setup.md`](10-app-setup.md) | UI configuration walkthrough for every app (Prowlarr, qBit, Sonarr, etc.). |
| [`research-indexers-2026.md`](research-indexers-2026.md) | Curated indexer/proxy research for 2025-2026. |

## Target architecture (TL;DR)

- **Cluster**: existing k3s, single Optiplex node, 8GB RAM, 256GB SSD + 500GB HDD.
- **Ingress**: Traefik (k3s default) with `letsencrypt` certresolver, `*.home.iamrpm.xyz` wildcard DNS.
- **GitOps**: ArgoCD (already running). Each media app = one `Application` using **bjw-s/app-template** Helm chart.
- **Namespace**: `media`.
- **No VPN** (Gluetun dropped) — see `08-legal-notes.md` for rationale + IN-specific notice info.
- **Downloads**: qBittorrent only (no usenet).
- **Indexers**: Prowlarr + FlareSolverr + MediaFusion (for TamilMV/TamilBlasters).
- **Streaming**: Jellyfin. Direct-play first; Intel QuickSync HW accel if Optiplex iGPU is Gen 6+.

## Apps in scope

| App | Role | Image | UI ingress |
|-----|------|-------|------------|
| Prowlarr | indexer mgr | `lscr.io/linuxserver/prowlarr` | `prowlarr.home.iamrpm.xyz` |
| FlareSolverr | Cloudflare bypass for IN indexers | `ghcr.io/flaresolverr/flaresolverr` | (internal, ClusterIP only) |
| MediaFusion | TamilMV/Blasters scraper → Prowlarr-compat | `mhdzumair/mediafusion` | (internal, ClusterIP only) |
| Sonarr | TV mgr | `lscr.io/linuxserver/sonarr` | `sonarr.home.iamrpm.xyz` |
| Radarr | movie mgr | `lscr.io/linuxserver/radarr` | `radarr.home.iamrpm.xyz` |
| Bazarr | subtitle mgr | `lscr.io/linuxserver/bazarr` | `bazarr.home.iamrpm.xyz` |
| qBittorrent | torrent client | `ghcr.io/hotio/qbittorrent` | `qbit.home.iamrpm.xyz` |
| Jellyfin | streaming frontend | `lscr.io/linuxserver/jellyfin` | `jellyfin.home.iamrpm.xyz` |

Lidarr (music) deferred — add later via same template.

## Why these images

- **LSIO (`lscr.io/linuxserver/*`)** for *arrs + Jellyfin: consistent `PUID/PGID/TZ` env vars, ecosystem alignment, every tutorial assumes it.
- **hotio/qbittorrent**: LSIO's qBit lags on libtorrent fixes; hotio rebases faster. Community standard for qBit in 2025.
- **bjw-s/app-template** chart: 2025 de-facto standard (Truecharts went paid, k8s-at-home archived).

## Key constraints

- Single node, 8GB RAM → tight. Resource limits enforced. Swap bumped from 6GB → 12GB.
- Hardlinks require completed-dl + final media on **same filesystem** → both go on HDD.
- RWO PVCs are fine on single node (RWO = one node, not one pod).
- No multi-node, no NFS (Jellyfin DB on network storage is a known footgun).
