# 07 — Verification (End-to-End)

Run all checks. Each should pass before declaring done.

## Cluster-level

```bash
# All media pods Running, no CrashLoopBackOff
kubectl -n media get pods
# Expect: prowlarr, flaresolverr, sonarr, radarr, bazarr, qbittorrent, jellyfin all 1/1 Running

# All ArgoCD apps Synced + Healthy
kubectl -n argocd get app | grep -E 'prowlarr|sonarr|radarr|bazarr|qbit|jellyfin|flaresolverr'
# Expect: all Synced, all Healthy

# Resource pressure
kubectl top pod -n media
kubectl top node
# Expect: no pod at memory limit, node < 90% RAM

# Swap
ssh <node> 'free -h && swapon --show'
# Expect: swap total ~12G, usage low
```

## TLS + Ingress

```bash
for sub in jellyfin sonarr radarr prowlarr bazarr qbit; do
  echo "=== $sub ==="
  curl -sI "https://${sub}.home.iamrpm.xyz" | head -1
  echo | openssl s_client -servername "${sub}.home.iamrpm.xyz" \
    -connect "${sub}.home.iamrpm.xyz:443" 2>/dev/null \
    | openssl x509 -noout -issuer -dates
done
# Expect: HTTP/2 200 or 302 for each. Issuer = Let's Encrypt. notAfter > 30 days out.
```

## Functional flow

1. **Prowlarr → indexers**:
   - UI → System → Indexer Stats → ≥3 indexers tested green within last hour.

2. **Sonarr/Radarr → Prowlarr**:
   - Sonarr UI → System → Health → no warnings about indexers/download client.
   - Same for Radarr.

3. **Test download end-to-end** (use a small free/legal content for the smoke test — e.g., a Creative Commons movie, or Sintel/Big Buck Bunny):
   - Add to Radarr → Sonarr/Radarr searches → release found → handed to qBit.
   - qBit downloads (visible at `https://qbit.home.iamrpm.xyz`).
   - On completion, Sonarr/Radarr imports — UI shows "Imported" event.
   - On node: file at `/data/media/movies/<title>/...` exists.

4. **Hardlink verification**:
   ```bash
   ssh <node>
   # Replace paths with actual completed download
   ls -li /data/downloads/complete/movies/<file>
   ls -li /data/media/movies/<title>/<file>
   # Same inode number on both → hardlink working
   stat -c '%h' /data/media/movies/<title>/<file>
   # Link count = 2 (or more)
   ```
   If inodes differ → Sonarr/Radarr is copying, not hardlinking. Causes: filesystems differ, Settings → Media Management → "Use hardlinks" is OFF, or PUID/PGID mismatch preventing cross-link.

5. **Bazarr subs**:
   - Bazarr UI → Movies/Series → pick the imported test item → Manual search → download sub → file `<title>.en.srt` appears next to the media file on disk.

6. **Jellyfin playback**:
   - Library scan in Dashboard → Activity → finished.
   - Test movie shows up in the Movies library.
   - Click play in browser → starts within 5s.
   - Dashboard → Active Sessions → confirm "**Direct Play**" (or "Transcoding (QSV)" if QSV enabled). Should NEVER be "Transcoding (software)" for routine content on this hardware — software transcode will stutter.

7. **(If QSV enabled) HW transcode check**:
   ```bash
   # Inside Jellyfin pod, when a transcode session is active:
   kubectl -n media exec deploy/jellyfin -- ls -la /dev/dri
   # Should show: card0, renderD128
   kubectl -n media exec deploy/jellyfin -- intel_gpu_top  # if available
   # Should show non-zero "Video/0" engine util while transcoding
   ```

## Persistence (survive pod restart)

```bash
kubectl -n media delete pod -l app.kubernetes.io/name=sonarr
# Pod restarts → UI still has all configured indexers, shows, settings.

kubectl -n media delete pod -l app.kubernetes.io/name=jellyfin
# Pod restarts → libraries still listed, no re-scan needed.
```

If config is lost → hostPath isn't mounted correctly, or PUID/PGID mismatch wiped permissions.

## GitOps round-trip

```bash
# Edit a value (e.g., bump Sonarr memory limit) in the repo
git diff k8s/apps/media/sonarr/values.yaml
git commit + push   # only if user authorized push

# ArgoCD picks up change within 3 min (or force refresh)
kubectl -n argocd get app sonarr
# Expect: OutOfSync → Synced after auto-sync wave
```

## Rollback drill

```bash
# Disable autoSync briefly to demo rollback
argocd app set sonarr --sync-policy none

# Revert a values change
git revert <bad-commit>
git push

# Manually sync
argocd app sync sonarr
```

## Done = all of the above pass

If any step fails, see [`09-troubleshooting.md`](09-troubleshooting.md) (not yet written — add as issues are hit).
