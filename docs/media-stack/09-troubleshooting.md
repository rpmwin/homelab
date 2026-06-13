# 09 — Troubleshooting (Issues Hit During Rollout)

Real issues encountered during the 2026-06-13 build, with diagnosis + fix. Organized by symptom.

---

## 1. ArgoCD ApplicationSet generated 0 apps

### Symptom
```
kubectl -n argocd get appset media-stack
NAME          AGE
media-stack   58s
kubectl -n argocd get apps -l app.kubernetes.io/part-of=media-stack
No resources found in argocd namespace.
```

### Diagnosis
```
kubectl -n argocd get appset media-stack -o yaml | tail -40
```
Status condition: `unable to resolve git revision: failed to list refs: error creating SSH agent: "SSH agent requested but SSH_AUTH_SOCK not-specified"`

Root cause: appset used `git@github.com:rpmwin/homelab.git` (SSH) but ArgoCD had no SSH key registered. Existing apps (backend-1, etc.) use HTTPS — repo is public.

### Fix
Switched both `repoURL` entries in `k8s/_platform/argocd/media-applicationset.yaml` from `git@github.com:rpmwin/homelab.git` to `https://github.com/rpmwin/homelab.git`. Commit `610bd6d`.

### Lesson
Always match the auth scheme of existing working apps. Public repos = HTTPS, no creds needed. SSH requires registering a key via `argocd repo add` first.

---

## 2. ImagePullBackOff — DNS "Try again" errors

### Symptom
```
Failed to pull image "lscr.io/linuxserver/prowlarr:latest": ... dial tcp: lookup lscr.io: Try again
```

### Diagnosis
Tailscale was injecting `100.100.100.100` (MagicDNS) into systemd-resolved as one of the DNS servers. When kubelet's containerd hit it, MagicDNS didn't know `lscr.io` and returned SERVFAIL/timeout instead of falling through cleanly.

```
resolvectl status | grep 'DNS Server'
# showed: 1.1.1.1, 8.8.8.8, AND 100.100.100.100 (Tailscale)
```

### Fix
On node:
```bash
sudo tailscale set --accept-dns=false
sudo systemctl restart systemd-resolved
sudo systemctl restart k3s
kubectl -n media delete pods --all  # force re-pull
```

### Lesson
Tailscale + k3s + systemd-resolved is a known footgun. If pods can't resolve external names but the node can, suspect Tailscale DNS injection first. `--accept-dns=false` keeps Tailscale connectivity without it touching DNS.

---

## 3. CoreDNS upstream flakiness ("server misbehaving")

### Symptom
From Traefik / Prowlarr / any media pod:
```
dial tcp: lookup acme-v02.api.letsencrypt.org on 10.43.0.10:53: server misbehaving
dial tcp: lookup 1337x.to on 10.43.0.10:53: i/o timeout
```

Direct test from a busybox pod:
```bash
kubectl -n media run nettest --rm -i --restart=Never --image=busybox:1.36 -- \
  nslookup acme-v02.api.letsencrypt.org
# ;; connection timed out; no servers could be reached
```

### Diagnosis
CoreDNS Corefile had `forward . /etc/resolv.conf` — picked up whatever the node's resolv.conf had. Airtel home internet flakes on UDP/53 to public DNS under bursts (when Traefik fires 6 parallel queries during cert renewal, several time out).

### Fix
Override CoreDNS with explicit upstreams + health checks via `coredns-custom` ConfigMap:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns-custom
  namespace: kube-system
data:
  upstream.override: |
    forward . 1.1.1.1 8.8.8.8 9.9.9.9 {
      prefer_udp
      max_concurrent 1000
      policy random
      health_check 5s
    }
```
Then `kubectl -n kube-system rollout restart deploy/coredns`.

The `import /etc/coredns/custom/*.override` directive in k3s's default Corefile auto-picks up override files.

### Lesson
For homelab clusters on flaky residential internet, never let CoreDNS auto-discover upstream. Pin to multiple known-reliable resolvers with health checks. Cache aggressively.

---

## 4a. Byparr `NS_ERROR_UNKNOWN_HOST` for every target site

### Symptom
After deploying Byparr, smoke test fails with `playwright._impl._errors.Error: Page.goto: NS_ERROR_UNKNOWN_HOST`. Prowlarr UI shows `Unable to access ..., blocked by CloudFlare Protection` (misleading — the real error is DNS).

### Diagnosis
1. CoreDNS upstream over Airtel UDP/53 was unreliable under Firefox's parallel DNS query load. Replaced with explicit upstream nameservers via `dnsPolicy: None`:
   ```yaml
   defaultPodOptions:
     dnsPolicy: "None"
     dnsConfig:
       nameservers: [1.1.1.1, 8.8.8.8, 9.9.9.9]
   ```

2. But still failing — `getent hosts 1337x.to` returned only **IPv6** addresses (`2606:4700:...`). Pod has no IPv6 connectivity (k3s cluster is IPv4-only). Firefox tried IPv6 first, failed, reported as `UNKNOWN_HOST`.

### Fix
Add `no-aaaa` + `single-request-reopen` options to dnsConfig:
```yaml
defaultPodOptions:
  dnsPolicy: "None"
  dnsConfig:
    nameservers: [1.1.1.1, 8.8.8.8, 9.9.9.9]
    options:
      - { name: ndots, value: "1" }
      - { name: timeout, value: "2" }
      - { name: attempts, value: "2" }
      - { name: single-request-reopen }
      - { name: no-aaaa }
```

After this, `getent hosts 1337x.to` returns only IPv4 (Cloudflare CDN IPs `172.67.188.67`, `104.21.40.193`). Byparr smoke test returns `"status":"ok","message":"Success"` with valid `cf_clearance` cookie.

Commits `8a9b41e` (dnsPolicy:None) + `ae7fb6a` (no-aaaa).

### Lesson
IPv4-only clusters + dual-stack DNS responses = silent failures. Always set `no-aaaa` on pods that need to reach external dual-stack hosts. The `UNKNOWN_HOST` error is misleading — it's actually "host resolved but I picked an unreachable IP family."

---

## 4. FlareSolverr can't bypass Cloudflare anymore

### Symptom
Prowlarr UI → Indexers → 1337x → Test:
```
Unable to access 1337x.to, blocked by CloudFlare Protection.
```

### Diagnosis
FlareSolverr (`ghcr.io/flaresolverr/flaresolverr:latest`) was running, but Cloudflare upgraded to Turnstile in 2024-2025. FlareSolverr's headless-Chrome fingerprint is detected and blocked. FlareSolverr's official repo is essentially archived.

### Fix
Swap to **Byparr** (community successor, same API, port 8191):
```yaml
# k8s/apps/media/flaresolverr/values.yaml
image:
  repository: ghcr.io/thephaseless/byparr   # was: ghcr.io/flaresolverr/flaresolverr
  tag: latest
resources:
  limits:
    memory: 1Gi   # heavier than FlareSolverr (Chromium-based)
```
No client-side config change — Prowlarr's existing "FlareSolverr" indexer-proxy entry still points at `http://flaresolverr.media.svc.cluster.local:8191` and works.

Commit `dd84f94`.

### Lesson
Anti-bot tech evolves; the proxy layer must too. Keep watch for Byparr alternatives if/when it falls behind too.

---

## 5. HTTPS returning Traefik default cert (no LE cert issued)

This was the largest debug session — multiple compounding issues.

### Symptom
```bash
echo | openssl s_client -servername prowlarr.home.iamrpm.xyz \
  -connect prowlarr.home.iamrpm.xyz:443 2>/dev/null | openssl x509 -noout -issuer
# issuer=CN=TRAEFIK DEFAULT CERT     ← not Let's Encrypt
```

### Diagnosis chain

#### 5a. `spec.tls` block was misleading us
First instinct: kubectl said `PORTS: 80` (not `80, 443`) — added `spec.tls.hosts` to bjw-s values.yaml. Ingresses then showed `80, 443`. But cert was still default.

Compared to working backend-1:
```bash
kubectl -n default get ingress backend-1-ingress -o jsonpath='{.spec.tls}'
# (empty — no spec.tls)
kubectl -n default get ingress backend-1-ingress
# PORTS: 80   (also just 80, yet has valid LE cert)
```

Conclusion: **Traefik handles TLS via annotations alone**. `spec.tls` is unnecessary and was actually breaking HTTP-01 challenge routing. Removed it. Commit `691e2a5`.

#### 5b. Real root cause: EC2 nginx not proxying port 80 ACME challenges
Tested ACME challenge path directly:
```bash
curl -sI http://prowlarr.home.iamrpm.xyz/.well-known/acme-challenge/test
# HTTP/1.1 404 Not Found
# Server: nginx/1.28.3        ← not Traefik, this is the EC2 nginx
```

The setup:
- DNS: `*.home.iamrpm.xyz → 16.112.38.35` (EC2 elastic IP)
- EC2 has nginx with Wireguard tunnel (`wg0`) to home cluster at `10.0.0.2`
- Port 443: `stream { ... proxy_pass 10.0.0.2:443; ssl_preread on; }` ✓ (SNI passthrough — works)
- Port 80: handled by HTTP server block that ONLY serves `/var/www/certbot` locally (for the EC2-local certbot that issues `ssh.home.iamrpm.xyz`)

Result: LE's HTTP-01 challenge for `prowlarr.home.iamrpm.xyz` hit EC2 nginx port 80 → returned 404 → LE marked domain failed.

Backend-1 had a valid cert only because it was issued in the past when port 80 routing was different. New certs couldn't be issued in this setup.

#### 5c. LE rate-limit lock-out
~5 failed authorizations per FQDN per hour triggered:
```
429 :: urn:ietf:params:acme:error:rateLimited
too many failed authorizations (5) for "radarr.home.iamrpm.xyz" in the last 1h0m0s
```
Locked us out for ~1 hour each time. Stopped retrying anything until window cleared.

### Fix

Edited `/etc/nginx/nginx.conf` on EC2 — split the port-80 server block:

```nginx
http {
    # ssh.home → local certbot (unchanged)
    server {
        listen 80;
        server_name ssh.home.iamrpm.xyz;
        location /.well-known/acme-challenge/ { root /var/www/certbot; }
        location / { return 301 https://$host$request_uri; }
    }

    # Everything else → proxy ACME to home Traefik via wg0
    server {
        listen 80 default_server;
        server_name _;
        location /.well-known/acme-challenge/ {
            proxy_pass http://10.0.0.2:80;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        location / { return 301 https://$host$request_uri; }
    }
    # ... (existing 8443 ttyd block unchanged)
}
```
`sudo nginx -t && sudo systemctl reload nginx`.

Then on cluster: removed `spec.tls` from values.yaml, restarted Traefik. After CoreDNS was also stable, all 6 certs issued within minutes.

### Verification
```bash
for h in prowlarr sonarr radarr bazarr qbit jellyfin; do
  echo "" | openssl s_client -servername ${h}.home.iamrpm.xyz \
    -connect ${h}.home.iamrpm.xyz:443 2>&1 | grep -E '^issuer=|^subject='
done
# subject=CN=prowlarr.home.iamrpm.xyz issuer=C=US, O=Let's Encrypt, CN=YR1
# (etc. for all 6)
```

### Lesson
When DNS points to a reverse proxy you don't fully control, **verify port 80 reaches the actual TLS-issuing server before troubleshooting at the cluster layer**. One `curl http://<host>/.well-known/acme-challenge/test` from outside reveals this in 5 seconds. The 5-hour debug spiral happened because we assumed port 80 worked since port 443 did.

Also: LE's rate limit punishes you fast. Get config right BEFORE letting Traefik retry repeatedly. Use LE staging during testing if possible.

---

## 6. Cert presented = "Traefik default" even after issuance

### Symptom
After `acme.json` clearly had all 6 cert entries, `openssl s_client` still returned default cert for some hosts.

### Diagnosis
Traefik caches cert mapping in memory. New certs in `acme.json` aren't auto-loaded until the next config reload trigger.

### Fix
```bash
kubectl -n kube-system rollout restart deploy/traefik
```
After restart, all 6 hosts presented correct LE certs.

### Lesson
Whenever `acme.json` is mutated but the wrong cert is served, restart Traefik. It's cheap and idempotent.

---

## 7. Prowlarr background errors at startup (non-fatal)

### Symptom
```
[Error] IndexerDefinitionUpdateService: Definition update failed
  Resource temporarily unavailable (indexers.prowlarr.com:443)
[Error] ServerSideNotificationService: Failed to retrieve notifications
  Resource temporarily unavailable (prowlarr.servarr.com:443)
```

### Diagnosis
These are background HTTP calls to Prowlarr's update service. The UI works fine, indexers can be added/tested. EAGAIN (errno 11) on socket connect = transient resource unavailable, often DNS related.

### Fix
Non-fatal. Restarting Prowlarr pod after CoreDNS was stable made them disappear. No action needed if UI works.

### Lesson
Don't chase non-fatal background errors. If the main app function works, ignore Prowlarr's chatter about its own update servers.

---

## Final state (post all fixes)

- HTTPS works with LE certs on all 6 media subdomains
- ApplicationSet generates apps from `k8s/apps/media/*/values.yaml`
- CoreDNS pinned to 1.1.1.1 / 8.8.8.8 / 9.9.9.9 with health checks
- Tailscale DNS off (`--accept-dns=false`) — doesn't interfere with kubelet
- EC2 nginx forwards `/.well-known/acme-challenge/` to home cluster via wg0
- Byparr running in place of FlareSolverr — bypasses current Cloudflare Turnstile

## Commits (chronological)

| Commit | What |
|--------|------|
| `a1bfa8f` | Initial media stack (manifests + docs) |
| `e58ef25` | Switch to ApplicationSet pattern |
| `610bd6d` | Repo URL: SSH → HTTPS (auth fix) |
| `a5131bc` | (Failed attempt) Add spec.tls — reverted later |
| `dd84f94` | FlareSolverr → Byparr |
| `691e2a5` | Remove spec.tls (Traefik annotations alone work) |

## Files touched on EC2 (not in git)

- `/etc/nginx/nginx.conf` — split port-80 server blocks for ACME forwarding
- Backup at `/etc/nginx/nginx.conf.bak.before-acme`

## Files touched on node (not in git)

- `/swapfile2` (8GB) + `/etc/fstab` entry — total swap bumped to 14GB
- `/etc/sysctl.conf` — `vm.swappiness=10`
- Tailscale: `sudo tailscale set --accept-dns=false`
- `/opt/media/*` and `/data/*` directories (chown 1000:1000)
