# 10 — In-App Setup Guide (UI configuration after pods are running)

Each app needs UI configuration after the pod is up. Order matters — later apps depend on earlier ones.

URLs (replace if your domain differs):

| App | URL | Default user | Default pwd |
|-----|-----|--------------|-------------|
| Prowlarr | `https://prowlarr.home.iamrpm.xyz` | (none) | (none — wizard creates one) |
| qBittorrent | `https://qbit.home.iamrpm.xyz` | `admin` | auto-generated temp pwd (from logs) |
| Sonarr | `https://sonarr.home.iamrpm.xyz` | (none) | (wizard) |
| Radarr | `https://radarr.home.iamrpm.xyz` | (none) | (wizard) |
| Bazarr | `https://bazarr.home.iamrpm.xyz` | (none) | (wizard) |
| Jellyfin | `https://jellyfin.home.iamrpm.xyz` | (none) | (onboarding) |
| Byparr (FlareSolverr-compat) | internal: `http://flaresolverr.media.svc.cluster.local:8191` | n/a | n/a |

---

## Internal Service URLs (use these when wiring apps together)

These are cluster-internal DNS names. Apps inside the `media` namespace use them:

| Service | URL | Used by |
|---------|-----|---------|
| Prowlarr API | `http://prowlarr.media.svc.cluster.local:9696` | Sonarr, Radarr |
| qBittorrent API | `http://qbittorrent.media.svc.cluster.local:8080` | Sonarr, Radarr |
| Sonarr API | `http://sonarr.media.svc.cluster.local:8989` | Bazarr |
| Radarr API | `http://radarr.media.svc.cluster.local:7878` | Bazarr |
| Byparr (Cloudflare proxy) | `http://flaresolverr.media.svc.cluster.local:8191` | Prowlarr (for CF-protected indexers) |

---

## Step 2a — Prowlarr

URL: `https://prowlarr.home.iamrpm.xyz`

### 1. Initial setup wizard

- Authentication Method: **Forms (Login Page)**
- Authentication Required: **Enabled**
- Set Username + Password → Save

### 2. Add Byparr as an Indexer Proxy (essential for Cloudflare sites)

This is **required** before adding 1337x or any CF-protected indexer. Without it, you get `Unable to access ... blocked by CloudFlare Protection`.

- **Settings → Indexers → "Indexer Proxies" tab → Add (+) → FlareSolverr**
- Name: `byparr`
- Tags: **add a tag** named `flaresolverr` (you'll attach this same tag to CF-protected indexers below)
- Host: `http://flaresolverr.media.svc.cluster.local:8191`
- Request Timeout: `60`
- Save → **Test** — should be green.

### 3. Add indexers

Settings → Indexers → "Indexers" tab → Add (+).

Recommended Tier 1 set (per `research-indexers-2026.md`):

| Indexer | Cloudflare? | Action |
|---------|-------------|--------|
| 1337x | YES | Add → **Tags field → flaresolverr** → Test |
| TheRARBG | YES | Add → Tags → flaresolverr → Test |
| TorrentGalaxy | YES | Add → Tags → flaresolverr → Test |
| The Pirate Bay | sometimes | Add → Tags → flaresolverr → Test |
| YTS | NO | Add → Test (no tag needed) |
| Nyaa.si | NO | Add → Test (anime only) |

**Critical**: Cloudflare-protected indexers will fail their Test unless they share a tag with the FlareSolverr/Byparr proxy. The tag is the wiring.

### 4. Copy API Key

**Settings → General → Security → API Key** → copy this. Used by Sonarr/Radarr below.

### 5. (Optional) Skip Definition Updates errors

In logs you may see `IndexerDefinitionUpdateService: Definition update failed` to `indexers.prowlarr.com:443`. Non-fatal background fetch. Ignore.

---

## Step 2b — qBittorrent

URL: `https://qbit.home.iamrpm.xyz`

### 1. Get the auto-generated temporary password

```bash
kubectl -n media logs deploy/qbittorrent | grep -i 'temporary password'
```
Example: `A temporary password is provided for this session: ABC123xyz`

### 2. Login + change password

- Login: `admin` / `<temp-pwd>`
- **Tools → Options → Web UI → Authentication**
  - Username: `admin` (or change)
  - Change password → set new permanent pwd → Save
- UI restarts → log back in with new pwd.

### 3. Download paths

**Tools → Options → Downloads**:

- Default Save Path: `/downloads/complete`
- Keep incomplete torrents in: `/downloads/incomplete` → enable
- Append `.!qB` extension to incomplete files: ON
- "Run external program on torrent completion": leave empty (Sonarr/Radarr poll the API)

### 4. Categories (required for *arr import paths)

Sidebar (left) → right-click "Categories" → **Add category**:

| Name | Save path |
|------|-----------|
| `tv` | `/downloads/complete/tv` |
| `movies` | `/downloads/complete/movies` |
| `music` | `/downloads/complete/music` (defer if not using Lidarr) |

### 5. BitTorrent settings

**Tools → Options → BitTorrent**:
- DHT: ON
- PeX: ON
- LSD: ON
- Encryption: Prefer encryption
- Max active downloads: 5 (adjust)
- Max active torrents: 10

### 6. Connection — port forwarding (optional)

By default, qBit listens on incoming TCP/UDP port 6881 inside the cluster. The values.yaml exposes it via NodePort `30881`. For inbound peer connections (better seed ratio), forward port 30881 from your router to the node IP. Skip if you only care about downloading.

---

## Step 2c — Sonarr

URL: `https://sonarr.home.iamrpm.xyz`

### 1. Initial setup wizard

- Authentication Method: **Forms (Login Page)**
- Authentication Required: Enabled
- Set admin + pwd → Save

### 2. Wire Prowlarr → Sonarr (indexer sync)

**Settings → Indexers → "+" → Sync from Prowlarr** (or in Prowlarr → Settings → Apps → Sonarr):

If from Sonarr side:
- Prowlarr URL: `http://prowlarr.media.svc.cluster.local:9696`
- API Key: (from Prowlarr Settings → General → Security)
- Test → green → Save.

Better: do it from Prowlarr side so the wiring is bidirectional:
- Prowlarr → **Settings → Apps → "+" → Sonarr**
- Prowlarr Server: `http://prowlarr.media.svc.cluster.local:9696` (this is sent TO Sonarr so Sonarr knows where to reach Prowlarr)
- Sonarr Server: `http://sonarr.media.svc.cluster.local:8989`
- API Key: (from Sonarr's Settings → General → Security)
- Sync Level: **Full Sync**
- Test → green → Save.

After this, Prowlarr pushes all its indexers to Sonarr automatically.

### 3. Wire qBittorrent → Sonarr

**Settings → Download Clients → "+" → qBittorrent**:

- Host: `qbittorrent.media.svc.cluster.local`
- Port: `8080`
- URL Base: (empty)
- Username: `admin`
- Password: (your permanent qBit pwd)
- Category: `tv`
- Recent Priority: Last
- Older Priority: Last
- Initial State: Start
- Sequential Order: OFF
- First and Last First: OFF
- Test → green → Save.

### 4. Media Management settings

**Settings → Media Management**:

| Setting | Value |
|---------|-------|
| Use Hardlinks instead of Copy | **ON** (critical — saves disk space, makes import instant) |
| Import Extra Files | `srt, nfo` (optional) |
| Rename Episodes | ON |
| Replace Illegal Characters | ON |
| Standard Episode Format | `{Series Title} - S{season:00}E{episode:00} - {Episode Title} {Quality Full}` |
| Daily Episode Format | `{Series Title} - {Air-Date} - {Episode Title} {Quality Full}` |
| Series Folder Format | `{Series Title}` |
| Season Folder Format | `Season {season:00}` |
| Create empty series folders | OFF |
| Delete empty folders | ON |

### 5. Root folder

**Settings → Media Management → Root Folders → "+"**:

- Path: `/media/tv`
- Save

### 6. Quality profile (optional tweak)

**Settings → Profiles → Quality Profiles**:

- HD-1080p is the standard default. Edit it if needed:
  - Allowed: WEBDL-1080p, Bluray-1080p, HDTV-1080p
  - Cutoff: Bluray-1080p (Sonarr will upgrade to this if better release appears)

### 7. Copy Sonarr's API Key

**Settings → General → Security → API Key** → copy. Bazarr needs this.

### 8. Test — add a series

**Series → Add New**:

- Search any show.
- Path: `/media/tv`
- Monitor: "Future Episodes" (or "All Episodes" to backfill)
- Quality Profile: HD-1080p
- Save → Sonarr immediately searches indexers.

---

## Step 2d — Radarr

URL: `https://radarr.home.iamrpm.xyz`

Mirror Sonarr setup exactly. Key differences:

| Setting | Value |
|---------|-------|
| App URL in Prowlarr → Apps | Add as `Radarr`, URL `http://radarr.media.svc.cluster.local:7878` |
| qBit category | `movies` (not `tv`) |
| Root folder | `/media/movies` (not `/media/tv`) |
| Movie Folder Format | `{Movie Title} ({Release Year})` |
| Movie Format | `{Movie Title} ({Release Year}) {Quality Full}` |

Copy Radarr's API key after setup — Bazarr needs it.

---

## Step 2e — Bazarr

URL: `https://bazarr.home.iamrpm.xyz`

### 1. Initial setup wizard

- Authentication: Forms → admin + pwd → Save.

### 2. Wire Sonarr → Bazarr

**Settings → Sonarr**:

- Use Sonarr: ON
- Address: `sonarr.media.svc.cluster.local`
- Port: `8989`
- Base URL: (empty)
- SSL: OFF
- API Key: (from Sonarr)
- Update from Sonarr: Use folders only / Full update — pick "Full" for now
- Save → Test → green.

### 3. Wire Radarr → Bazarr

**Settings → Radarr** — same pattern with port `7878` and Radarr's API key.

### 4. Languages

**Settings → Languages → "Languages Profiles" → Add**:

- Profile name: `English`
- Languages: English → Save

Mark as default in **Settings → Sonarr → Default Settings → Default Language Profile = English** (same for Radarr).

### 5. Providers

**Settings → Providers** — pick 2-3, all free (some need signup):

| Provider | Notes |
|----------|-------|
| **OpenSubtitles.com** | Free signup at https://www.opensubtitles.com — enter username + pwd here |
| **Subscene** | No signup, works as-is |
| **Yifysubtitles** | No signup, but only movies |
| **Addic7ed** | TV-focused, no signup |

Enable, fill creds, Save.

### 6. Subtitles settings

**Settings → Subtitles**:

- Use Embedded Subtitles: ON (skip download if file already has subs)
- Forced Subtitles: OFF (unless you want them)
- Subtitle Folder: alongside media (default)
- Encoding: UTF-8

### 7. Test

After Sonarr/Radarr libraries populate (~30s), Bazarr's Series/Movies tabs show your media. Pick one item → click → "Manual Search" → results appear → click download → .srt file appears next to the video on disk.

---

## Step 2f — Jellyfin

URL: `https://jellyfin.home.iamrpm.xyz`

### 1. Onboarding wizard

1. **Display language**: English (or preferred).
2. **Create admin user**: username + pwd.
3. **Setup media libraries** — click "+" for each:

   **Library 1: TV Shows**
   - Content type: **TV Shows**
   - Display name: `TV Shows`
   - Folders: add `/media/tv`
   - Metadata downloaders: **TheTVDB** (primary), TheMovieDb (fallback)
   - Image fetchers: TheTVDB, FanArt
   - Save.

   **Library 2: Movies**
   - Content type: **Movies**
   - Display name: `Movies`
   - Folders: add `/media/movies`
   - Metadata: **TheMovieDB** (primary), OMDb (fallback)
   - Save.

4. **Preferred Display Language**: English (or your locale).
5. **Allow Remote Connections**: ON.
6. Finish.

### 2. Login + initial config

After wizard, log in as admin → top-right user icon → Dashboard.

### 3. Playback / hardware acceleration

**Dashboard → Playback → Transcoding**:

For this Optiplex (i3-2120 Sandy Bridge):
- Hardware acceleration: **None** (NO QSV — Sandy Bridge is too old for proper hardware encode)
- Software transcoding will be CPU-heavy. Avoid by **direct-play** — choose H.264 / AAC / MP4 releases when picking torrents.

### 4. Network settings

**Dashboard → Networking**:
- Published Server URL: `https://jellyfin.home.iamrpm.xyz`
- Known proxies: add `10.42.0.0/16` (your pod CIDR — so X-Forwarded-For works)
- Save

### 5. (Optional) Auto-organize TV scans

**Dashboard → Libraries → TV Shows → Manage Library**:
- Real-time monitoring: ON
- Library scan interval: every 30 min

### 6. User management

**Dashboard → Users → "+"** — create personal accounts for family/housemates. Each gets their own watch history, resume position, recommendations.

### 7. Plugins (optional but recommended)

**Dashboard → Plugins → Catalog**:

- **Subtitle Extract** — auto-extract subs from MKV containers
- **Trakt** — sync watch history to trakt.tv
- **OPDS** — for e-book libraries (only if you add Readarr)
- **Prometheus Metrics** — exposes `/metrics` for Grafana (see `11-monitoring.md` if/when added)

---

## End-to-end smoke test

Pick a small free/legal item to verify the pipeline:

### 1. Add a test movie

Radarr → **Movies → Add New** → search "Big Buck Bunny" → Add. Path: `/media/movies`. Save.

### 2. Trigger manual search

In Radarr → click the movie → **Search** (magnifying glass icon) → **Interactive Search**. Browse results → pick a 1080p H.264 release with green flags → click download.

### 3. Watch the pipeline

```bash
# qBit shows the download progressing
open https://qbit.home.iamrpm.xyz

# Radarr shows the activity (Activity → Queue)
open https://radarr.home.iamrpm.xyz

# On node, see the file land:
kubectl -n media exec deploy/sonarr -- ls -la /data/downloads/complete/movies/
# (after completion)
kubectl -n media exec deploy/sonarr -- ls -la /data/media/movies/
```

### 4. Verify hardlink (the critical optimization)

```bash
ssh devbox
INODE_DL=$(stat -c '%i' /data/downloads/complete/movies/<title>/*.mkv)
INODE_LIB=$(stat -c '%i' /data/media/movies/<title>/*.mkv)
[ "$INODE_DL" = "$INODE_LIB" ] && echo "HARDLINK OK" || echo "COPIED (wasted space)"
```

### 5. Stream in Jellyfin

Jellyfin → Movies → wait for library scan (~30s) → click the movie → Play.

Confirm in Jellyfin **Dashboard → Active Sessions** that the stream shows:
- `Direct Play` ✓ (no CPU transcoding)
- `Transcoding (Software)` ✗ (this Optiplex can't handle it well — pick different release)

---

## API key reference (save these somewhere)

Each *arr generates a stable API key at first run. Keep them handy:

| App | Location to find |
|-----|------------------|
| Prowlarr | Settings → General → Security |
| Sonarr | Settings → General → Security |
| Radarr | Settings → General → Security |
| Bazarr | Settings → General → Security |
| Jellyfin | Dashboard → API Keys → New |

These keys are used:
- For wiring apps together (Sonarr ↔ Prowlarr, Bazarr ↔ Sonarr/Radarr).
- For monitoring exporters (exportarr sidecars if you add `11-monitoring.md`).
- For Stremio addons / external integrations.
- For any homemade scripts that talk to the *arr APIs.

Store them in your password manager or a `secrets/` directory (NOT committed to git).
