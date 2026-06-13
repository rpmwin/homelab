# Public Prowlarr Indexers — Status as of Feb 2026

Scope: free, no-account, no-API-key, no FlareSolverr/Byparr required, available
out-of-the-box in Prowlarr's `+ Add Indexer` modal. Western movies/TV focus +
a few anime picks.

## TL;DR — Rock-solid in 2026

- **The Pirate Bay** — no CF, huge catalogue, healthiest general indexer.
- **EZTV** — TV-only but consistently up, no CF.
- **YTS** — movies-only (x264/x265 small encodes), no CF, very stable.
- **Nyaa.si** — anime gold standard, no CF.

## Recommended Indexer Matrix

| Name (in Prowlarr) | Search term | Coverage | Cloudflare | Health 2026-02 | Source |
|---|---|---|---|---|---|
| The Pirate Bay | `pirate` | General (Movies/TV/Music/Apps) | none (apibay endpoint) | healthy | [programming.dev thread](https://programming.dev/post/41364587) |
| EZTV | `eztv` | TV | none | healthy | [programming.dev thread](https://programming.dev/post/41364587) |
| YTS | `yts` | Movies | none (uses YTS JSON API) | healthy | [Servarr Wiki](https://wiki.servarr.com/prowlarr/supported-indexers) |
| Nyaa.si | `nyaa` | Anime | none | healthy | [programming.dev thread](https://programming.dev/post/41364587) |
| LimeTorrents | `lime` | General | none (clearnet mirror) | healthy | [programming.dev thread](https://programming.dev/post/41364587) |
| Bitsearch | `bitsearch` | General (meta-search) | none | degraded — HTTP 429 rate limits; set query limit 10/hr | [Prowlarr #2635](https://github.com/Prowlarr/Prowlarr/issues/2635) |
| Anidex | `anidex` | Anime | none | healthy | [Servarr Wiki](https://wiki.servarr.com/prowlarr/supported-indexers) |
| Subsplease | `subsplease` | Anime (current-season) | none | healthy | [programming.dev thread](https://programming.dev/post/41364587) |
| Shana Project | `shana` | Anime | none | healthy | [programming.dev thread](https://programming.dev/post/41364587) |
| LinuxTracker | `linux` | Linux ISOs (general fallback) | none | healthy | [programming.dev thread](https://programming.dev/post/41364587) |
| BT.etree | `etree` | Music (live recordings) | none | healthy | [programming.dev thread](https://programming.dev/post/41364587) |

> Quirk: even no-CF indexers can land in `Indexer Disabled` if Prowlarr fires
> queries too fast. Cap query/grab limits at 10/hr per [Prowlarr #2635](https://github.com/Prowlarr/Prowlarr/issues/2635).

## Requires FlareSolverr / Byparr — SKIP if you don't run one

| Name | Why it needs bypass |
|---|---|
| 1337x | Cloudflare challenge page — [Prowlarr #2577](https://github.com/Prowlarr/Prowlarr/issues/2577), [#1518](https://github.com/Prowlarr/Prowlarr/issues/1518) |
| KickassTorrents (kat.ws / kat.to) | Cloudflare intermittent |
| TorrentGalaxy (TGx) | Cloudflare + repeated outages mid-2025; see "dead/dying" below |
| RARBG-to / TheRARBG | Cloudflare + missing/broken Prowlarr definition |

## Looks active, actually wastes your time

- **RARBG** — preinstalled placeholder on fresh Prowlarr, but no definition;
  cannot be re-added. [Prowlarr #2595](https://github.com/Prowlarr/Prowlarr/issues/2595).
- **TheRARBG / rarbg-to** — appears in some lists, definition is broken in
  Prowlarr (works in Jackett only). Same issue thread.
- **Torrentz2** — meta-search; backing sources mostly dead in 2026, returns
  near-zero results for new releases.
- **Bitmagnet (public DHT crawler instance)** — runs but its public-instance
  index is sparse; ~100 grabs vs Nyaa's 150+ in real-world setups. Self-host
  for real value.

## Was good in 2024, dead/dying in 2025-2026

- **TorrentGalaxy** — extended outage through 2025 after Dutch ISP block; admins
  claim "not shutting down" but availability is unreliable. [Cybernews](https://cybernews.com/news/torrent-galaxy-down/).
  Treat as unreliable until further notice.
- **RARBG (original)** — shut 31 May 2023. Any "RARBG" entry in Prowlarr's
  built-in list is a stub. [Wikipedia](https://en.wikipedia.org/wiki/RARBG).
- **Zooqle** — gone since 2022, still occasionally referenced in old guides.
- **KickassTorrents (original kat.cr)** — gone; current `.ws`/`.to` mirrors are
  Cloudflare-walled clones, hit-or-miss.
- **EliteTorrent / DonTorrent (ES)** — still in Prowlarr list but Spanish-only;
  not in scope for Western general.

## Recommended minimum set (no FlareSolverr)

For a fresh Prowlarr with zero bypass infra:

1. The Pirate Bay
2. EZTV (TV)
3. YTS (movies)
4. LimeTorrents (general fallback)
5. Nyaa.si (anime, if relevant)

That covers >90% of mainstream Western movie/TV grabs without a CF proxy.
Add Bitsearch as a 6th once you've capped its rate limit.

## Sources

- [Servarr Wiki — Prowlarr Supported Indexers](https://wiki.servarr.com/prowlarr/supported-indexers)
- [Servarr/Wiki repo — supported-indexers.md](https://github.com/Servarr/Wiki/blob/master/prowlarr/supported-indexers.md)
- [TRaSH Guides — Prowlarr](https://trash-guides.info/Prowlarr/)
- [programming.dev — "What indexers do you use in Prowlarr"](https://programming.dev/post/41364587)
- [Prowlarr #2595 — RARBG definition missing](https://github.com/Prowlarr/Prowlarr/issues/2595)
- [Prowlarr #2577 — 1337x Cloudflare block](https://github.com/Prowlarr/Prowlarr/issues/2577)
- [Prowlarr #1518 — 1337x Cloudflare block (older)](https://github.com/Prowlarr/Prowlarr/issues/1518)
- [Prowlarr #2635 — rate-limit-induced indexer disabling](https://github.com/Prowlarr/Prowlarr/issues/2635)
- [Cybernews — TorrentGalaxy downtime](https://cybernews.com/news/torrent-galaxy-down/)
- [Bytesized Hosting — Prowlarr Guide 2026](https://bytesized-hosting.com/guides/prowlarr-guide-one-indexer-manager-for-all-your-arr-apps)
