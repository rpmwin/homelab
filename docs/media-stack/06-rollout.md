# 06 — Rollout (Step-by-Step Execution)

Follow in order. Each step has a verify command — don't proceed until verify passes.

---

## Step 0 — Confirm prereqs ✓ (2026-06-13)

- [x] Node probed — i3-2120 Sandy Bridge, SSD=`/opt/media`, HDD=`/data`. **No QSV.** Direct-play only.
- [x] Wildcard DNS — `*.home.iamrpm.xyz → 16.112.38.35` confirmed.
- [ ] **Swap still needed** — SSH to node and run:
  ```bash
  fallocate -l 8G /swapfile2
  chmod 600 /swapfile2
  mkswap /swapfile2
  swapon /swapfile2
  echo '/swapfile2 none swap sw 0 0' >> /etc/fstab
  sysctl vm.swappiness=10
  echo 'vm.swappiness=10' >> /etc/sysctl.conf
  ```
- [x] Node dirs created (`/opt/media/config/*`, `/data/media/*`, `/data/downloads/*`, `chown 1000:1000`).
- [x] Intel GPU plugin — SKIPPED (Sandy Bridge, no QSV).
- [x] All manifests written to repo.

---

## Step 1 — Push repo + apply ApplicationSet (CURRENT STEP)

We use an **ApplicationSet** (app-of-apps pattern) — one resource registers all media apps and any future ones added under `k8s/apps/media/*/`.

```bash
cd /Users/iamrpm/Developer/projects/homelab
git push   # confirm with user before running

# 1. Create media namespace
kubectl apply -f k8s/apps/media/_shared/namespace.yaml

# 2. Apply the ApplicationSet — one-time. Generates Application CRs for every
#    subdir under k8s/apps/media/ (excluding _shared).
kubectl apply -f k8s/_platform/argocd/media-applicationset.yaml

# 3. Watch rollout
kubectl -n argocd get appset media-stack
kubectl -n argocd get apps -l app.kubernetes.io/part-of=media-stack
kubectl -n media get pods -w
```

**What happens**: ApplicationSet controller scans the repo, finds 7 subdirs under `k8s/apps/media/*` (skipping `_shared`), generates 7 ArgoCD `Application` CRs (one per dir, named after the dir). Each Application pulls bjw-s/app-template + the matching `values.yaml` from this repo.

**Adding a new app later**: drop a new subdir with `values.yaml` into `k8s/apps/media/`, commit, push. ApplicationSet auto-creates the Application within ~3 min. Zero kubectl.

**Removing an app**: delete its subdir, commit, push. ApplicationSet prunes the Application + its workloads (because syncPolicy has `prune: true`).

**Note**: ApplicationSet + multi-source require ArgoCD ≥ v2.6. Check:
```bash
kubectl -n argocd get deploy argocd-server -o jsonpath='{.spec.template.spec.containers[0].image}'
```

---

## Step 2 (was 1) — Directory scaffolding (DONE — skip)

Already created. Reference:

```bash
mkdir -p /Users/iamrpm/Developer/projects/homelab/k8s/apps/media/{_shared,prowlarr,flaresolverr,sonarr,radarr,bazarr,qbittorrent,jellyfin}
mkdir -p /Users/iamrpm/Developer/projects/homelab/k8s/_platform/{argocd,traefik}
```

Add the `Namespace` manifest:

```yaml
# k8s/apps/media/_shared/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: media
```

Apply directly (or via ArgoCD if you wire ApplicationSet first):
```bash
kubectl apply -f k8s/apps/media/_shared/namespace.yaml
kubectl get ns media   # → Active
```

---

## Step 2 — Prowlarr (alone, smoke test)

1. Write `k8s/apps/media/prowlarr/values.yaml` (see [`05-apps.md`](05-apps.md)).
2. Write `k8s/apps/media/prowlarr/application.yaml` (see [`04-argocd-helm.md`](04-argocd-helm.md)).
3. Substitute `/opt/media` placeholder in values.yaml with actual mount.
4. Apply:
   ```bash
   kubectl apply -f k8s/apps/media/prowlarr/application.yaml
   ```
5. Watch:
   ```bash
   kubectl -n argocd get app prowlarr -w
   kubectl -n media get pods -w
   ```
6. Verify:
   ```bash
   curl -sI https://prowlarr.home.iamrpm.xyz   # → HTTP/2 200 (or 302 to /UI)
   ```
7. Open `https://prowlarr.home.iamrpm.xyz` → set admin password → Settings → General → record API key (used by Sonarr/Radarr later).
8. Add 3 indexers via UI: **1337x**, **TorrentGalaxy**, **The Pirate Bay** (or any 3 publics). Test → all green.

---

## Step 3 — FlareSolverr

1. Write `k8s/apps/media/flaresolverr/{application,values}.yaml`.
2. Apply application.
3. Verify pod Running:
   ```bash
   kubectl -n media get pods -l app.kubernetes.io/name=flaresolverr
   kubectl -n media logs deploy/flaresolverr | tail -20
   # Expect: "FlareSolverr is ready!" at port 8191
   ```
4. Smoke test from inside cluster:
   ```bash
   kubectl -n media run curltest --rm -it --image=curlimages/curl -- \
     curl -sS http://flaresolverr.media.svc.cluster.local:8191/v1 \
     -H 'Content-Type: application/json' \
     -d '{"cmd":"sessions.list"}'
   # → {"status":"ok",...}
   ```
5. Wire into Prowlarr: Prowlarr UI → Settings → Indexers → Add Indexer → **FlareSolverr** → Host: `http://flaresolverr.media.svc.cluster.local:8191` → Save → Test → green.

---

## Step 4 — qBittorrent

1. Write `k8s/apps/media/qbittorrent/{application,values}.yaml`.
2. Apply.
3. Get the auto-generated password from logs:
   ```bash
   kubectl -n media logs deploy/qbittorrent | grep -i 'temporary password'
   ```
4. Open `https://qbit.home.iamrpm.xyz`:
   - Login `admin` + the temp password.
   - Tools → Options → Web UI → set a permanent password.
   - Options → Downloads:
     - Default save path: `/downloads/complete`
     - Keep incomplete in: `/downloads/incomplete`
   - Categories tab → add `tv`, `movies` with respective save subpaths under `/downloads/complete`.
5. Verify download works with a known-safe torrent (e.g., a Linux ISO):
   - Magnet link → confirm download progresses → completed file appears in `/data/downloads/complete/...` on node.

---

## Step 5 — Sonarr

1. Write `k8s/apps/media/sonarr/{application,values}.yaml`.
2. Apply.
3. Open `https://sonarr.home.iamrpm.xyz`:
   - Authentication: set "Forms" auth, create user. (Don't expose without auth.)
   - **Settings → Indexers → "+" → "Sync from Prowlarr"** — paste Prowlarr URL `http://prowlarr.media.svc.cluster.local:9696` + API key from step 2.
   - **Settings → Download Clients → "+" → qBittorrent**:
     - Host: `qbittorrent.media.svc.cluster.local`
     - Port: `8080`
     - Username/password as set in step 4.
     - Category: `tv`
     - Test → green.
   - **Settings → Media Management**:
     - Use hardlinks instead of copy: **ON**
     - Import extra files: as desired (subs/info)
     - Rename episodes: ON (standard format)
   - **Add Series** → search a test show → select → root folder `/media/tv` → Monitor "Future Episodes" → Save.
4. Verify end-to-end:
   - Pick a test episode → Sonarr searches Prowlarr → qBit gets the .torrent → completes → Sonarr imports via hardlink → file appears in `/data/media/tv/<show>/<season>/...` on node.
   - Confirm hardlink: `ls -li /data/downloads/complete/tv/<file>` and `ls -li /data/media/tv/<show>/.../<file>` → **same inode**.

---

## Step 6 — Radarr

Mirror Sonarr setup. Differences:
- Category in qBit download client: `movies`.
- Root folder: `/media/movies`.
- Sync indexers from Prowlarr same way.

---

## Step 7 — Bazarr

1. Write + apply.
2. Open `https://bazarr.home.iamrpm.xyz`:
   - Settings → Sonarr → URL `http://sonarr.media.svc.cluster.local:8989`, API key from Sonarr → Settings → General.
   - Settings → Radarr → URL `http://radarr.media.svc.cluster.local:7878`, API key.
   - Settings → Providers → enable OpenSubtitles, Subscene, Yifysubtitles (or others; some need accounts).
   - Settings → Languages → add English (+ any others).
3. Verify: Bazarr lists shows/movies from Sonarr/Radarr within ~30s. Pick one → manual search → subtitle downloads → appears next to media file as `.en.srt`.

---

## Step 8 — Jellyfin

1. Write + apply (with QSV stanzas commented OUT for first pass — get it running first, add QSV after).
2. Open `https://jellyfin.home.iamrpm.xyz` → onboarding:
   - Create admin user.
   - Add library "TV Shows" → folder `/media/tv` → metadata TheTVDB.
   - Add library "Movies" → folder `/media/movies` → metadata TheMovieDB.
   - Allow remote connections.
3. Wait for initial library scan to finish (Dashboard → Activity).
4. Play a file in a browser → confirm playback. Check Dashboard → Playback → confirm "Direct play" in the active session info.
5. (If Optiplex Gen 6+) Uncomment QSV stanzas in values.yaml, re-apply:
   - `gpu.intel.com/i915: 1` in resources.limits
   - `securityContext.supplementalGroups: [<render-gid>]`
   - `/dev/dri` hostPath mount
   - Resync: `argocd app sync jellyfin --force`
   - In Jellyfin UI → Dashboard → Playback → Hardware Acceleration → Intel QuickSync → enable codec decodes/encodes → Save → restart.
   - Play a high-bitrate file that needs transcode → Dashboard → confirm "Transcoding (QSV)" not "(software)".

---

## Step 9 — (Later) MediaFusion for Indian content

Skip in initial rollout. Add after the stack is stable:

1. Read MediaFusion's docker-compose to understand mongo + redis deps.
2. Translate into manifests (3 deployments + 3 services in `media` ns) OR use an alternate Helm chart if available.
3. Wire into Prowlarr as a Torznab indexer.
4. See [`08-legal-notes.md`](08-legal-notes.md) for what to expect re: notices.

---

## Step 10 — Commit & push

```bash
cd /Users/iamrpm/Developer/projects/homelab
git add k8s/apps/media k8s/_platform docs/media-stack
git status
git commit -m "media: add *arr stack + jellyfin (k8s manifests + docs)"
# do NOT push without user confirmation
```

Proceed to [`07-verification.md`](07-verification.md) for the end-to-end test checklist.
