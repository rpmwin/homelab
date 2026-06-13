# Traefik

## Install

Bundled with k3s — no separate installation needed. Runs as a DaemonSet in `kube-system`.

```bash
kubectl -n kube-system get pods -l app.kubernetes.io/name=traefik
kubectl -n kube-system get svc traefik
```

## TLS / Let's Encrypt

Traefik manages LE certificates via its built-in certresolver.

Certresolver name: **`letsencrypt`**

Annotations used on every TLS ingress:
```yaml
annotations:
  traefik.ingress.kubernetes.io/router.entrypoints: web,websecure
  traefik.ingress.kubernetes.io/router.tls: "true"
  traefik.ingress.kubernetes.io/router.tls.certresolver: letsencrypt
```

No `spec.tls` block needed — certresolver auto-provisions.

## DNS

Domain: `home.iamrpm.xyz`  
Wildcard DNS: `*.home.iamrpm.xyz → 16.112.38.35` (set at registrar, DNS-only, no Cloudflare proxy)

One wildcard covers all subdomains — no per-app DNS changes needed.

## Node IP

- External (EC2 elastic / NAT): `16.112.38.35`
- Internal: `192.168.1.7`
- Tailscale: `100.114.206.128`

## Existing service URLs

| Service | URL |
|---------|-----|
| backend-1 | https://back1.home.iamrpm.xyz |
| backend-2 | https://back2.home.iamrpm.xyz |
| Prometheus | https://prom.home.iamrpm.xyz |
| Grafana | https://grafana.home.iamrpm.xyz |
| Media stack (WIP) | `*.home.iamrpm.xyz` — see k8s/apps/media/ |
