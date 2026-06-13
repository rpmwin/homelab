# 03 — Storage

## Layout

| Disk | Mount (example) | Holds |
|------|-----------------|-------|
| SSD 256GB | `/opt/media` | k3s data, all `/config` PVCs (sqlite DBs), Jellyfin cache, qBit incomplete |
| HDD 500GB | `/data` | Media library, qBit completed downloads |

Replace `/opt/media` and `/data` with the actual mount points discovered in [`02-prerequisites.md`](02-prerequisites.md).

## Why this split

- **sqlite IO**: Sonarr/Radarr/Prowlarr/Jellyfin all use sqlite. Lots of random reads/writes for library scans + metadata. SSD makes these snappy. HDD makes them grind.
- **Jellyfin cache**: thumbnails, transcoded segments, subtitle conversions. Must be fast. SSD.
- **Media files**: large sequential reads during streaming. HDD is fine — even a 5400rpm HDD does ~80 MB/s sequential, while a 1080p H.264 file plays at ~3-5 MB/s.
- **Downloads (incomplete)**: torrent clients do random writes across many chunks. SSD avoids HDD thrashing while downloading.
- **Downloads (complete) + media library on same FS**: hardlinks. Sonarr/Radarr can "import" via hardlink → no double-disk-use, instant import.

## Directory tree on node

```
/opt/media/
├── media-config/
│   ├── sonarr/
│   ├── radarr/
│   ├── prowlarr/
│   ├── bazarr/
│   ├── qbittorrent/
│   ├── jellyfin/
│   ├── flaresolverr/
│   └── mediafusion/
├── jellyfin-cache/
└── downloads-incomplete/

/data/
├── media/
│   ├── tv/
│   ├── movies/
│   └── music/
└── downloads/
    └── complete/
        ├── tv/
        ├── movies/
        └── music/
```

## Mount conventions used in pod values

| App | Pod path | Node hostPath | Notes |
|-----|----------|---------------|-------|
| all *arr + qBit + jellyfin | `/config` | `/opt/media/media-config/<app>` | per-app config + sqlite |
| sonarr / radarr / qbit | `/downloads` | `/data/downloads` | full downloads tree (incl. incomplete subdir) |
| qbit | `/downloads/incomplete` | `/opt/media/downloads-incomplete` | optional override — faster active dl |
| sonarr / radarr / bazarr / jellyfin | `/media` | `/data/media` | media library (RO for jellyfin, RW for *arr) |
| jellyfin | `/cache` | `/opt/media/jellyfin-cache` | transcode + thumbnail cache |

## PUID / PGID / TZ

- `PUID=1000`, `PGID=1000` — matches the `chown` in prereqs.
- `TZ=Asia/Kolkata` — for log timestamps + scheduled tasks.

## Why hostPath instead of PVC

hostPath chosen over `local-path-provisioner`-backed PVCs because:

1. **Explicit path control** — we need specific dirs on specific disks (SSD vs HDD split). PVCs let the provisioner pick.
2. **Predictable hardlinks** — both `complete` and `media` must be the same filesystem. hostPath makes this provable.
3. **Single-node cluster** — no scheduling benefit from PVCs.
4. **Survives PVC deletion** — accidental `kubectl delete pvc` would not touch the host dir.

Caveat: hostPath means the manifest is **node-specific**. If you ever add a node, switch to NFS or Longhorn.

## Hardlink check (do this after qBit + Sonarr are up)

```bash
# On node:
ls -li /data/downloads/complete/tv/<some-show>/Episode.mkv
ls -li /data/media/tv/<some-show>/<season>/Episode.mkv
# Both should have the same inode number → hardlink, not copy
```
