# Build Status — Media Stack

Snapshot as of 2026-06-13 ~20:30 IST. Update this file as the build progresses.

## Infrastructure ✓

| Item | Status | Notes |
|------|--------|-------|
| k3s cluster | ✓ running | single node `devbox`, Ubuntu 24.04, v1.35.5+k3s1 |
| ArgoCD | ✓ running | 16+ days, namespace `argocd` |
| ApplicationSet `media-stack` | ✓ Synced | auto-generates 7 Applications from `k8s/apps/media/*/` |
| Node directories (`/opt/media`, `/data/...`) | ✓ created | chown 1000:1000 |
| Swap (14GB total) | ✓ active | 6GB partition + 8GB swapfile |
| Tailscale DNS | ✓ disabled | `tailscale set --accept-dns=false` |
| CoreDNS upstream override | ✓ applied | forwards to 1.1.1.1/8.8.8.8/9.9.9.9 |

## Pods ✓

All 7 media pods Running:

| Pod | Image | Notes |
|-----|-------|-------|
| prowlarr | `lscr.io/linuxserver/prowlarr:latest` | UI up |
| sonarr | `lscr.io/linuxserver/sonarr:latest` | UI up |
| radarr | `lscr.io/linuxserver/radarr:latest` | UI up |
| bazarr | `lscr.io/linuxserver/bazarr:latest` | UI up |
| qbittorrent | `ghcr.io/hotio/qbittorrent:release-5.0.4` | UI up |
| jellyfin | `lscr.io/linuxserver/jellyfin:latest` | UI up |
| flaresolverr (Byparr) | `ghcr.io/thephaseless/byparr:latest` | Cloudflare bypass working |

## HTTPS ✓

All 6 subdomains serving valid Let's Encrypt certs:

| Host | Cert issuer |
|------|-------------|
| prowlarr.home.iamrpm.xyz | Let's Encrypt YR1 |
| sonarr.home.iamrpm.xyz | Let's Encrypt YR2 |
| radarr.home.iamrpm.xyz | Let's Encrypt YR1 |
| bazarr.home.iamrpm.xyz | Let's Encrypt YR1 |
| qbit.home.iamrpm.xyz | Let's Encrypt YR1 |
| jellyfin.home.iamrpm.xyz | Let's Encrypt YR1 |

## External dependencies (off-cluster)

| Thing | State |
|-------|-------|
| Wildcard DNS `*.home.iamrpm.xyz → 16.112.38.35` | ✓ at Namecheap parking DNS |
| EC2 nginx port 443 SNI passthrough → `10.0.0.2:443` | ✓ pre-existing |
| EC2 nginx port 80 `/.well-known/acme-challenge/` → `10.0.0.2:80` | ✓ added 2026-06-13 |
| Wireguard tunnel EC2 ↔ home cluster | ✓ pre-existing |

## App configuration progress

| App | UI auth set | Wired to deps | Test data added | Notes |
|-----|-------------|---------------|-----------------|-------|
| Prowlarr | partial | partial | — | Indexer-proxy "byparr" tag created. **Add CF-protected indexers with `flaresolverr` tag.** |
| qBittorrent | — | — | — | Get temp pwd from logs, set permanent, create `tv`/`movies` categories |
| Sonarr | — | — | — | Wait until Prowlarr indexers are in |
| Radarr | — | — | — | Mirror Sonarr |
| Bazarr | — | — | — | Wait until Sonarr/Radarr have API keys |
| Jellyfin | — | — | — | Last — needs library content before useful |

## Commits (chronological, post-rollout debug)

| Commit | Subject |
|--------|---------|
| `a1bfa8f` | Initial media stack (manifests + docs) |
| `e58ef25` | Switch to ApplicationSet pattern |
| `610bd6d` | Repo URL: SSH → HTTPS (auth fix) |
| `a5131bc` | (Failed attempt) Add spec.tls — reverted later |
| `dd84f94` | FlareSolverr → Byparr |
| `691e2a5` | Remove spec.tls (Traefik annotations alone work) |
| `8a9b41e` | Byparr: bypass CoreDNS via dnsPolicy:None + Cloudflare DNS |
| `ae7fb6a` | Byparr: disable AAAA lookups (IPv4-only cluster) |
| `528254a` | Docs: troubleshooting + in-app setup + IPv6 fix note |

## Known constraints / decisions logged

| Decision | Reason |
|----------|--------|
| **No QSV transcoding** in Jellyfin | i3-2120 Sandy Bridge (Gen 2) — too old. Direct-play only. |
| **No VPN** for qBit | User accepts home IP exposure. IN ISP risk = low. See `08-legal-notes.md`. |
| **No spec.tls** in ingresses | Traefik annotations alone are sufficient and don't break HTTP-01. |
| **hostPath, not PVC** | Single-node + need explicit SSD vs HDD placement. |
| **Byparr** over FlareSolverr | Cloudflare Turnstile defeats FlareSolverr in 2025. |
| **MediaFusion deferred** | First get English content working end-to-end. Add Indian content scraper in a follow-up. |
| **No Real-Debrid (yet)** | Optional QoL upgrade — defer until base stack works. See research doc. |
| **No monitoring** (yet) | Defer exportarr sidecars + ServiceMonitors until stack stable. See `01-architecture.md` resource budget. |

## What's next

1. **Finish Prowlarr config**: add Tier-1 indexers, tag CF-protected ones with `flaresolverr`. (`10-app-setup.md` step 2a)
2. **qBittorrent**: temp pwd → permanent pwd → categories. (step 2b)
3. **Sonarr**: wire to Prowlarr + qBit + media management settings + first test series. (step 2c)
4. **Radarr**: mirror Sonarr. (step 2d)
5. **Bazarr**: wire to Sonarr + Radarr + sub providers. (step 2e)
6. **Jellyfin**: onboarding + add `/media/tv` + `/media/movies` libraries. (step 2f)
7. **End-to-end smoke test** with a CC-licensed test movie. (step 3)
8. **Hardlink verification** on node. (step 4)

## Deferred / backlog

- **MediaFusion deployment** for Tamil/Telugu/Hindi indexers (TamilMV, TamilBlasters)
- **Real-Debrid + Decypharr + Zilean** for cached streams (₹250/mo via IN reseller)
- **Lidarr** for music management
- **Exportarr sidecars** + Grafana dashboards (monitoring)
- **Renovate** for image tag auto-bumps
- **Backup** of `/opt/media/config/` via restic or velero
- **Authelia/Authentik SSO** in front of *arr UIs
- **Switch DNS to Cloudflare** + DNS-01 challenge (eliminates EC2 nginx dependency for cert issuance)
