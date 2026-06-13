# 02 — Prerequisites (Run BEFORE writing manifests)

> **Status (2026-06-13)**: Node already probed. Results documented below. DNS ✓. Dirs ✓. Swap still needs +8GB. Manifests written. Pending: push to git + apply Application CRDs to argocd ns.

## Probed node specs (devbox)

| | Value |
|-|-------|
| Hostname | `devbox` |
| CPU | Intel i3-2120 @ 3.30GHz — **Sandy Bridge Gen 2** |
| iGPU | Intel 2nd Gen Integrated Graphics |
| QSV decision | **DISABLED** — Sandy Bridge has no encode support. Direct-play only. |
| RAM | 7.6GB |
| Swap | 6GB partition (`/dev/sda3`) — **needs +8GB** |
| SSD (238GB, ROTA=0) | mounted at **`/`** — use `/opt/media/` for configs |
| HDD (465GB, ROTA=1) | mounted at **`/data`** — use `/data/media/` and `/data/downloads/` |
| OS | Ubuntu 24.04.4 LTS |
| k3s | v1.35.5+k3s1 |
| ArgoCD | ✓ running in `argocd` ns (16 days old) |
| StorageClass | `local-path` (default, rancher.io/local-path) |
| Node dirs | ✓ created via Job (2026-06-13) |
| DNS wildcard | ✓ `*.home.iamrpm.xyz → 16.112.38.35` confirmed |

---

## 1. Probe the node (DONE — skip)

SSH into the Optiplex and capture:

```bash
# CPU / generation — used to decide on Intel QuickSync (QSV) for Jellyfin
lscpu | grep -E 'Model name|Architecture'

# iGPU presence
lspci | grep -iE 'vga|display|3d'

# Memory + swap
free -h

# Disks + mount points — note which mount is SSD vs HDD
lsblk -o NAME,SIZE,TYPE,MOUNTPOINT,ROTA,FSTYPE
# ROTA=0 → SSD, ROTA=1 → HDD

# k3s version + node info
kubectl get node -o wide

# Is ArgoCD running?
kubectl get ns argocd
kubectl -n argocd get deploy
```

**Decision points captured from output:**

- [ ] **Optiplex generation** — record CPU model (e.g. `i5-7500T`). Gen 6+ (`i*-6xxx` or newer) → enable QSV path in Jellyfin. Older → direct-play only.
- [ ] **SSD mount point** — e.g. `/` or `/mnt/ssd`. Use for configs + caches.
- [ ] **HDD mount point** — e.g. `/mnt/hdd` or `/mnt/data`. Use for media + completed downloads.
- [ ] **ArgoCD repo URL + sync mode** — `kubectl -n argocd get configmap argocd-cm -o yaml | grep -A5 repositories`.

## 2. Wildcard DNS

The plan adds 6 new subdomains. Don't add them one-by-one — set a wildcard once.

At your DNS provider (Cloudflare / wherever `iamrpm.xyz` is hosted):

```
*.home.iamrpm.xyz   A   <node-public-IP>   (proxy: OFF for Traefik LE HTTP-01)
```

If Cloudflare proxying is ON, Let's Encrypt HTTP-01 challenge will still work via Cloudflare, but TLS termination becomes CF's cert, not LE. For homelab simplicity, leave proxy OFF (DNS-only).

Verify:
```bash
dig +short test.home.iamrpm.xyz
# should return your node's public IP
```

## 3. Bump swap 6GB → 14GB (on node) ← STILL NEEDED

8GB RAM is tight with Jellyfin + FlareSolverr (Chromium) + the existing backends.

```bash
# On the node, as root:
fallocate -l 6G /swapfile2
chmod 600 /swapfile2
mkswap /swapfile2
swapon /swapfile2

# Persist
echo '/swapfile2 none swap sw 0 0' >> /etc/fstab

# Verify
free -h
swapon --show
```

Adjust kernel pressure preferences if needed:
```bash
sysctl vm.swappiness=10
echo 'vm.swappiness=10' >> /etc/sysctl.conf
```

## 4. Create node directories (DONE — skip)

Created 2026-06-13 via `media-setup-dirs` Job. Real paths confirmed: `/opt/media/config/*`, `/opt/media/cache/jellyfin`, `/opt/media/downloads-incomplete`, `/data/media/{tv,movies,music}`, `/data/downloads/complete/{tv,movies,music}`. All owned by `1000:1000`.

Original commands for reference:

```bash
# Configs + caches on SSD
mkdir -p /opt/media/media-config/{sonarr,radarr,prowlarr,bazarr,qbittorrent,jellyfin,flaresolverr,mediafusion}
mkdir -p /opt/media/jellyfin-cache
mkdir -p /opt/media/downloads-incomplete

# Media + completed downloads on HDD (same filesystem → hardlinks work)
mkdir -p /data/media/{tv,movies,music}
mkdir -p /data/downloads/complete/{tv,movies,music}

# Ownership — LSIO/hotio default PUID=1000, PGID=1000
chown -R 1000:1000 /opt/media/media-config /opt/media/jellyfin-cache /opt/media/downloads-incomplete
chown -R 1000:1000 /data/media /data/downloads

# Permissive but not world-writable
chmod -R 775 /opt/media/media-config /data/media /data/downloads
```

## 5. Confirm Traefik LE certresolver name (DONE — skip)

Check what certresolver the existing ingresses use:

```bash
grep -r certresolver /Users/iamrpm/Developer/projects/homelab/k8s/apps
# Expected: traefik.ingress.kubernetes.io/router.tls.certresolver: letsencrypt
```

Record the value (usually `letsencrypt`) — every new media ingress uses the same name.

## 6. (Optional) Intel GPU device plugin

Only if Optiplex is Gen 6+ AND you want HW-accelerated transcoding:

```bash
# Deploy Intel device plugin operator (one-time, cluster-wide)
kubectl apply -k 'https://github.com/intel/intel-device-plugins-for-kubernetes/deployments/gpu_plugin/overlays/nfd_labeled_nodes?ref=v0.30.0'

# Verify
kubectl get pods -n intel-device-plugin
kubectl describe node <node> | grep gpu.intel.com
# Expect: gpu.intel.com/i915: 1 (or higher with sharedDevNum)
```

If you skip this, Jellyfin will still work — just no HW transcoding. Direct-play covers most cases.

## Done checklist

- [ ] CPU model recorded, QSV path decided.
- [ ] SSD + HDD mount points recorded.
- [ ] Wildcard DNS `*.home.iamrpm.xyz` resolves to node IP.
- [ ] Swap is 12GB+, `swapon --show` confirms.
- [ ] All node directories exist with correct ownership.
- [ ] Certresolver name recorded.
- [ ] (Optional) Intel device plugin running.

Proceed to [`03-storage.md`](03-storage.md).
