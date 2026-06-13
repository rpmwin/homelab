# 08 — Legal Notes (India Context)

> Personal notes on torrent legal exposure for an India-based home user. Not legal advice. Reflects the realistic state in 2025-2026.

## Decision: no VPN

Plan deliberately drops Gluetun (VPN sidecar). Reasons:

1. Simpler pod (no shared netns, no kill-switch logic, no provider account).
2. No VPN throughput hit on streaming-adjacent traffic.
3. Realistic risk in India is low (see below).
4. User has explicitly opted in to this trade-off.

If risk tolerance changes later, Gluetun can be added in front of qBit as a sidecar without touching other components.

## How a notice would actually arrive

**Detection**: Anti-piracy firms (MarkMonitor, IP-Echelon, Aiplex India) join popular swarms as peers. They log every IP that connects, with timestamp + torrent hash.

**Mapping**: WHOIS lookup → identifies Airtel (or whatever ISP). Notice sent to ISP's abuse desk.

**Subscriber match**: Airtel checks CGNAT logs (retained 180+ days per DoT licensing) → finds the subscriber who held that IP at that timestamp.

**Delivery**: Email to the registered email on the broadband account (sometimes SMS). Format: "copyright infringement detected at <time> on <content>, please stop, repeated violations may result in throttling/termination."

**No court summons**, no police, no third-party CC. Just an ISP warning email.

## Severity in India (real-world, 2024-2026)

- **Very low for home users**. Airtel/Jio do not aggressively police torrenting like US ISPs (Comcast/AT&T DMCA forwarding).
- Most Indian users torrent for years without ever receiving a notice.
- When notices do arrive, they're warnings; throttling/disconnection happens only after repeated ignored notices over months.
- The bulk of "anti-piracy" enforcement in India is **John Doe / Ashok Kumar injunctions** that order ISPs to block torrent *sites* (DNS/SNI level — which is why TamilMV etc. constantly rotate domains). Individual subscribers are rarely targeted.
- Criminal prosecution of a home leecher: **effectively never happens**. Copyright Act Sec 63 has up to 3yr prison, but enforced against uploaders / commercial pirates / site operators, not home downloaders.
- Civil suits against home users: theoretically possible (statutory damages), practically unheard of.

## Higher-risk windows to know about

1. **Brand-new Bollywood / Tollywood / Kollywood releases** (day-of-theatrical to ~week 1). Aiplex and similar watch these aggressively. Wait 1+ week.
2. **Static-IP corporate broadband plans** — taken more seriously than consumer dynamic-IP plans.
3. **Office, college, hostel, landlord networks** — complaint goes to the network operator, not you. Don't torrent on those.

## Routing through AWS EC2 is WORSE, not better

If qBit traffic egresses via the EC2 elastic IP (via Wireguard policy routing or running qBit on EC2 itself):

- Swarm sees the EC2 elastic IP.
- Notices go to `abuse@amazon.com` — AWS is far stricter than Airtel.
- AWS AUP forbids "copyright infringement". First DMCA → 24-48h warning. Second/third → instance suspended. Repeated → entire AWS account suspended (could lose other services on that account too).
- **Conclusion**: leave qBit traffic on the Airtel route. Don't tunnel torrents through EC2.

The EC2 + Wireguard setup is for *inbound* access (you → homelab from outside). It does not currently affect *outbound* qBit peer connections. Leave it that way.

## Defensive measures already taken

- **No port-forwarded inbound** for qBit's BitTorrent port unless explicitly opened. Outbound-only torrenting still works but with smaller peer set.
- **Web UI behind Traefik + Let's Encrypt** TLS — no plain-HTTP qBit auth over the public internet.
- **Strong qBit Web UI password** on first run (default temp password rotated immediately).
- **Categories** ensure downloads land in known paths, not arbitrary disk locations.

## What to monitor

- The email registered on your Airtel broadband account — make sure it's an inbox you read. A notice missed is a notice that escalates.
- Airtel My Account portal → any "abuse / TOS" flags. Rare, but check occasionally.
- ISP throttling — `speedtest` periodically; sudden slowdown on torrent traffic specifically can be ISP traffic shaping.

## If a notice arrives

1. Don't panic — first notice is a warning, no action required other than "stop."
2. Pause / remove the offending torrents in qBit.
3. Optionally rotate WAN IP (reboot router) — Airtel CGNAT will probably hand a new IP.
4. Consider adding Gluetun + a VPN provider (Mullvad / Proton / AirVPN). Note that this only protects future traffic — past activity is already logged at the ISP.

## If you ever want to add VPN later

Gluetun sidecar pattern (deferred from this plan):

```yaml
# qBit values.yaml — sketched
controllers:
  main:
    containers:
      gluetun:
        image:
          repository: qmcgaw/gluetun
          tag: latest
        env:
          VPN_SERVICE_PROVIDER: mullvad
          VPN_TYPE: wireguard
          WIREGUARD_PRIVATE_KEY: <from secret>
          SERVER_COUNTRIES: Netherlands
        securityContext:
          capabilities:
            add: [NET_ADMIN]
        # ...
      main:
        # qBit container — networks: gluetun
```

Pods share network namespace via `shareProcessNamespace` + careful liveness probes. Many existing community examples on GitHub.
