# *arr Indexer Research — India, 2025/2026

Scope: k3s self-host, no VPN, streaming-first, Hollywood + Indian theatrical (Tamil/Telugu/Hindi/Malayalam) + some anime. Already running Prowlarr + FlareSolverr.

This is opinionated. Where multiple paths exist I pick one and say why.

---

## TL;DR (read this first)

1. **Public-tracker *arr alone will not get you fresh Indian theatrical content.** The releases live on TamilMV / 1TamilBlasters / TamilRockers, which (a) rotate domains weekly, (b) sit behind Cloudflare, and (c) have **no working built-in Prowlarr definition**. The community-accepted solution is **MediaFusion**, which scrapes those sites server-side and exposes a single Torznab endpoint.
2. **For Hollywood + English TV**, the default Prowlarr indexer set is mostly fine — but ~10 of the bundled ones are dead in 2025. Trim to ~6 reliable ones.
3. **Swap FlareSolverr for Byparr.** FlareSolverr is effectively unmaintained vs. Cloudflare's current challenges; Byparr is the active drop-in replacement and is what r/selfhosted has been migrating to since April 2025.
4. **Buy Real-Debrid ($3/mo) and front it with Decypharr.** This is the single biggest quality-of-life upgrade for the stack — instant streaming, no seeding obligation, no ratio. INR payments work via 247premiumcart / 365premium resellers.
5. **For "Vikram dropped Friday, I want it Sunday"**: the *arr stack is not the right primary tool. Use **Stremio + MediaFusion + Real-Debrid** for fresh-content discovery and one-off plays, keep *arr for the library you actually curate.

---

## 1. Public-tracker indexer set for 2025/2026

### What's dead or unreliable

A March 2025 GitHub issue ([Prowlarr/Prowlarr#2358](https://github.com/Prowlarr/Prowlarr/issues/2358)) flags these as either dead sites or sites Prowlarr can no longer scrape even with FlareSolverr:

- Anidex
- BitSearch (syncs but returns nothing — [issue #2210](https://github.com/Prowlarr/Prowlarr/issues/2210))
- ExtraTorrent.st
- GloDLS
- Isohunt2
- kickasstorrents.to / kickasstorrents.ws (the original KAT died years ago; current "kickass" sites are clones with poor data)
- SceneTime
- ShowRSS
- Torrentz2nz

**Disable these in Prowlarr.** They burn API budget and cause Sonarr/Radarr to mark searches as failed.

### What works reliably (community consensus, mid-2025 → 2026)

Cross-referenced from the [programming.dev "what indexers do you use" thread](https://programming.dev/post/41364587), the [Bytesized Hosting Prowlarr guide](https://bytesized-hosting.com/guides/prowlarr-guide-one-indexer-manager-for-all-your-arr-apps), and TorrentFreak's 2025 popularity rankings ([1337x is #2 most-popular public tracker](https://en.wikipedia.org/wiki/1337x)):

| Indexer | Use | Cloudflare? | Notes |
|---|---|---|---|
| **1337x** | Movies + TV, general | Yes — needs FlareSolverr/Byparr | The single most reliable public source in 2025 |
| **TorrentGalaxy (TGx)** | Movies + TV | Sometimes | Up and down through 2024–25; check mirror status |
| **TheRARBG** | Movies + TV | No | RARBG successor; archive is the value, new releases hit-or-miss |
| **YTS** | Movies only, small sizes | No | Useful for low-bandwidth Hollywood; quality is what it is |
| **EZTV** | TV only | No | RSS works, but its catalog is narrowing |
| **The Pirate Bay** | Everything | No | Patchy uptime; keep enabled as backup |
| **Nyaa.si** | Anime | No | The anime tracker. Period. |

That's your **Tier 1** for public English/anime content. Six to seven indexers is the sweet spot — more than that and you're paying latency without adding unique releases (heavy cross-posting between public trackers).

### Custom YAML indexers (Cardigann)

Prowlarr supports custom Cardigann YML definitions in `Definitions/Custom/`. The de-facto repo is **[dreulavelle/Prowlarr-Indexers](https://github.com/dreulavelle/Prowlarr-Indexers)**, which ships YAMLs for:

- `torrentio.yml` — Torrentio (Stremio addon) as a Torznab source
- `zilean.yml` — DMM hash index (Real-Debrid cached content)
- `comet.yml`, `aiostreams.yml`, `knightcrawler.yml`, `mediafusion`-friendly endpoints
- `elfhosted-torrentio.yml`, `elfhosted-public.yml` — ElfHosted aggregator endpoints
- `orionoid.yml`, `torbox.yml`, `stremthru.yml`, `debridio.yml`

**Important**: there is **no community YAML for TamilMV / TamilBlasters in this repo or in Prowlarr/Indexers**. People have asked ([MediaFusion#216](https://github.com/mhdzumair/MediaFusion/issues/216)) and the answer is consistently "use MediaFusion, the domain rotation makes a static YAML impossible to maintain."

### Cloudflare: drop FlareSolverr, use Byparr

FlareSolverr is no longer keeping up with Cloudflare's challenges and is widely reported as "non-functional, won't be fixed" ([ElfHosted blog, Apr 2025](https://store.elfhosted.com/blog/2025/04/16/byparr-bypasses-flaresolverr/)). **[Byparr](https://github.com/ThePhaseless/Byparr)** is the drop-in replacement — same port, same API contract — but uses Camoufox (anti-detect Firefox) under the hood. Migration is a single image swap in your k8s manifest. Solvearr ([nabil-ak/Solvearr](https://github.com/nabil-ak/Solvearr)) is a lighter option if you're memory-constrained, but Byparr handles harder challenges.

---

## 2. Indian content — the actual problem

### State of native Prowlarr support

**There is no working built-in or community-maintained Prowlarr indexer for TamilMV, 1TamilBlasters, TamilRockers, BollyShare, or KatMovieHD as of 2026.** The Prowlarr maintainers won't ship one because:

- The sites change domains weekly to dodge ISP blocks (Jio, Airtel block them DNS-level — relevant for you in India)
- Most are behind Cloudflare with aggressive challenges
- Cardigann YML is static and can't track moving domains

This has been the answer in every issue thread I checked, including [MediaFusion#216](https://github.com/mhdzumair/MediaFusion/issues/216) and the 1tamilmv GitHub topic.

### The MediaFusion path

**[MediaFusion (mhdzumair/MediaFusion)](https://github.com/mhdzumair/MediaFusion)** is a FastAPI service that:

- Scrapes 40+ sources server-side, including **TamilMV, 1TamilBlasters, Movierulz, regional Telugu/Malayalam sites**
- Tracks domain rotation centrally so you don't have to
- Aggregates Torrentio, Zilean, Jackett, Prowlarr, RSS, BT4G, DHT
- Exposes one **Torznab endpoint** you plug into Prowlarr/Sonarr/Radarr
- Also runs as a **Stremio addon** with the same backend

This is the only path that meaningfully solves IN content in 2026. The native Prowlarr integration is documented and live, but there's a known gotcha — [issue #404](https://github.com/mhdzumair/MediaFusion/issues/404): MediaFusion-as-Prowlarr-indexer sometimes syncs to Radarr but not Sonarr; workaround is to add it as a manual Torznab indexer directly inside Sonarr.

Public instance: `https://mediafusion.elfhosted.com` (ElfHosted-hosted, rate-limited).
Self-host: Docker Compose is documented; the deployment dir has Kubernetes manifests too. For your k3s setup, MediaFusion deploys cleanly — it's a FastAPI app + MongoDB + Redis + scraper workers. No Helm chart upstream, you'd write a small Kustomize or convert their compose.

### How fast does fresh content actually appear?

Based on TamilMV release-pattern observation discussed across HiFiVision, Stremio Reddit, and BitBrowser writeups ([reference](https://www.bitbrowser.net/blog/how-to-access-tamilmv-in-2025)):

| Stage | Typical lag from theatrical |
|---|---|
| HDCAM / TS rip (low quality) | **6–48 hours** |
| Clean HQ HDCAM | 2–5 days |
| WEB-DL / HDRip (the one you actually want) | **OTT release date** — typically 21–35 days post-theatrical for Tamil/Telugu, 14–28 for Hindi |
| BluRay | 60–90 days |

Reality check: if you're a streaming-first watcher who hates CAM rips, **the *arr stack will not give you Vikram-on-Sunday in good quality** — that quality literally doesn't exist yet for 1–3 weeks. What MediaFusion + Stremio + Real-Debrid gives you is the ability to *grab the cam the moment it lands* if you want it, and then have Radarr auto-upgrade to WEB-DL automatically when the OTT rip arrives, via Radarr's quality-cutoff upgrade behavior.

### Cross-posting from TamilMV → 1337x/TGx

Slow and inconsistent. Tamil/Telugu rips do eventually appear on 1337x (sometimes same-day for big releases, often 2–3 days), but the metadata is messier and naming inconsistent, which causes Sonarr/Radarr matching failures. Don't rely on it.

---

## 3. Day-of-release IN workflow — what people actually do

Pieced together from the [HiFiVision regional-content thread](https://www.hifivision.com/threads/mind-blown-by-stremio-but-how-to-watch-regional-contents.97863/) and Stremio addon guide writeups:

**The de-facto Indian selfhost workflow in 2025/2026**:

1. **Stremio** as the player UI on phone/TV/desktop
2. **MediaFusion** addon (public ElfHosted instance to start, self-hosted later)
3. **Real-Debrid** account ($3/mo, INR via reseller)
4. Optionally **AIOStreams** in front of MediaFusion + Torrentio + Comet, to deduplicate

Total cost: ~₹300/mo. Latency from "pick a movie" to "playing in 4K": ~5 seconds (the Real-Debrid magic — the file is already cached on their CDN for popular content).

For library curation (the part where you *do* want *arr): Radarr + Sonarr + Prowlarr + MediaFusion-as-Torznab, downloading to qBittorrent or Decypharr-via-Real-Debrid. New IN releases get pulled in as they appear; quality auto-upgrades work normally.

---

## 4. Private trackers — worth it from India?

Short answer: **not the highest leverage thing for your use case.**

- **HDBits** — ultra-closed, invite only, almost impossible to get into ([InstallGentoo wiki](https://igwiki.lyci.de/wiki/Private_trackers)). Skip.
- **PassThePopcorn (PTP)** — movies-only, invite only, IRC interviews happen but rare. Skip unless you have a friend with an invite.
- **BroadcastTheNet (BTN)** — TV only, harder than PTP. Skip.
- **TorrentLeech / IPTorrents** — occasionally open signups (or paid donation entry ~$10). Mid-tier; some value for English content but adds little over a working public set + Real-Debrid.
- **MyAnonamouse (MAM)** — books/audiobooks. Not relevant here.

For Indian content specifically: there is no significant private tracker scene. TamilMV *is* the canonical source. Private trackers don't help you here.

Track open signups via [InviteHawk](https://www.invitehawk.com/topic/135842-private-trackers-signup-schedule-open-application-and-irc-signups/) if curious, but **for your stated use case, Real-Debrid + MediaFusion delivers more than any private tracker would**.

---

## 5. Stremio vs *arr — use both

The Indian self-host community largely uses Stremio *instead of* *arr because:

- *arr expects matching to TMDB/TVDB metadata — regional Tamil/Telugu/Malayalam metadata is patchy on those, especially for new releases
- Sonarr/Radarr fight you on non-English release naming conventions
- Stremio + MediaFusion + RD is "click and play", no library management overhead

But running both side-by-side is the right move for someone who already has the *arr stack up:

- **Stremio for**: fresh releases, casual browsing, regional content, anything you watch once
- **\*arr for**: shows/movies you want permanently in Jellyfin, want to upgrade qualities over time, want offline access, family library

They share MediaFusion as the indexer, so there's no duplicated infra cost.

---

## 6. Real-Debrid + Decypharr — the unlock

[Decypharr](https://github.com/sirrobot01/decypharr) and [RDT-Client](https://github.com/rogerfar/rdt-client) both emulate qBittorrent's WebAPI but route to Real-Debrid. Sonarr/Radarr think they're talking to qBittorrent; downloads actually resolve from RD's cached pool.

Why it's worth it for India:

1. **No seeding** — your ISP can't see you participating in a swarm (you're just HTTPS-downloading from RD's servers)
2. **Cached content is instant** — most popular Hollywood and big-budget IN releases are pre-cached on RD; download = ~1 second to "queued in library"
3. **No VPN needed for the download path** (you still want one for casually browsing TamilMV in your browser, but the *arr path is HTTPS to RD only)
4. **Zilean indexer** ([dreulavelle/Prowlarr-Indexers zilean.yml](https://github.com/dreulavelle/Prowlarr-Indexers/blob/main/Custom/zilean.yml)) lets Prowlarr query RD's cached hash database directly — fast searches that only return things RD already has

**Decypharr vs RDT-Client**: Decypharr is newer, actively maintained, multi-debrid (RD + AllDebrid + Torbox), and is what ElfHosted standardized on in 2025. Pick Decypharr.

INR payment: confirmed working via [365premium.com](https://www.365premium.com/) and [247premiumcart.com](https://247premiumcart.com/product-category/real-debrid/) — UPI/Paytm/Indian debit cards. ~₹250/mo for 30 days.

---

## 7. Final concrete recommendation

### Tier 1 — Enable immediately (Prowlarr)

| Indexer | Why |
|---|---|
| **1337x** | Top public tracker 2025; needs Byparr |
| **TheRARBG** | Solid catalog, no CF issues |
| **TorrentGalaxy** | Good for fresh Hollywood releases |
| **The Pirate Bay** | Backup; no CF |
| **YTS** | Cheap-bandwidth Hollywood movies |
| **Nyaa.si** | Anime — must-have |
| **MediaFusion** (Torznab via custom URL) | All Indian content, plus Stremio scraper aggregation |
| **Zilean** (via dreulavelle YAML) | Only if you've added Real-Debrid |

### Tier 2 — Nice to have

- **EZTV** — TV-only, narrowing catalog but cheap to keep on
- **BT4G** — DHT-based, sometimes finds what others miss
- **Torrentio** (via dreulavelle YAML) — Stremio scraper exposed as Torznab; redundant if you already have MediaFusion but a useful cross-check
- **Bitmagnet** — self-hosted DHT crawler; long-term play for resilience

### Disable / remove

Anidex, BitSearch, ExtraTorrent.st, GloDLS, Isohunt2, all KAT clones, SceneTime, ShowRSS, Torrentz2nz. They eat API budget and produce errors.

### Single-best workflow for "I want the new Vikram movie within 2 days"

**Deploy this order**:

1. Swap FlareSolverr → Byparr in your k3s manifests (one-line image change)
2. Buy Real-Debrid via 247premiumcart (~₹250/mo, INR/UPI)
3. Deploy **Decypharr** as a sidecar; configure Sonarr/Radarr to use it as the qBittorrent download client (in parallel with your existing qBit — set Decypharr as primary)
4. **Self-host MediaFusion** on k3s (FastAPI + Mongo + Redis); add its Torznab URL into Prowlarr as a generic Torznab indexer. Configure it to scrape TamilMV / 1TamilBlasters / Movierulz / regional Telugu sites
5. Add **Zilean** custom YAML to Prowlarr for fast Real-Debrid hash lookups
6. Install **Stremio** on your phone + TV; configure with MediaFusion addon pointing at your self-hosted instance, plus your RD key. This is your fresh-release path
7. Keep *arr running as your library curator — same MediaFusion endpoint, Radarr auto-upgrades CAM → WEB-DL → BluRay over time

Realistic outcome: HDCAM/TS rip within 1–2 days, auto-upgrade to WEB-DL on OTT release ~3 weeks later, all automated.

### Should you deploy MediaFusion now?

**Yes, today.** There is no simpler alternative. The "just add TamilMV to Prowlarr" path does not exist in 2026 and won't. Start with the **public ElfHosted instance** (zero setup, rate-limited) to validate the workflow end-to-end, then self-host on k3s once you've confirmed it solves your problem. The self-hosted version removes rate limits and gives you control over which scrapers run.

---

## Sources

- [Prowlarr indexer removal tracking — Prowlarr/Prowlarr#2358 (Mar 2025)](https://github.com/Prowlarr/Prowlarr/issues/2358)
- [BitSearch sync issue — Prowlarr/Prowlarr#2210](https://github.com/Prowlarr/Prowlarr/issues/2210)
- [dreulavelle/Prowlarr-Indexers (custom Cardigann YAMLs)](https://github.com/dreulavelle/Prowlarr-Indexers)
- [MediaFusion main repo](https://github.com/mhdzumair/MediaFusion)
- [MediaFusion#216 — TamilMV/TamilBlasters indexer feasibility](https://github.com/mhdzumair/MediaFusion/issues/216)
- [MediaFusion#404 — Prowlarr→Sonarr sync gotcha](https://github.com/mhdzumair/MediaFusion/issues/404)
- [MediaFusion deployment Docker Compose docs](https://github.com/mhdzumair/MediaFusion/blob/main/deployment/docker-compose/README.md)
- [ElfHosted — Byparr bypasses Flaresolverr (Apr 2025)](https://store.elfhosted.com/blog/2025/04/16/byparr-bypasses-flaresolverr/)
- [Solvearr — lightweight FlareSolverr alternative](https://github.com/nabil-ak/Solvearr)
- [Decypharr ElfHosted product page](https://store.elfhosted.com/product/decypharr/)
- [rogerfar/rdt-client](https://github.com/rogerfar/rdt-client)
- [ElfHosted Plex + Radarr/Sonarr + Real-Debrid guide](https://docs.elfhosted.com/guides/media/plex-realdebrid-aars/)
- [Best Stremio addons 2026 — ElfHosted](https://docs.elfhosted.com/stremio-addons/guide/recommended-addons/)
- [HiFiVision — Stremio regional content thread](https://www.hifivision.com/threads/mind-blown-by-stremio-but-how-to-watch-regional-contents.97863/)
- [AIOStreams (Viren070)](https://github.com/Viren070/AIOStreams)
- [Viren070 AIOStreams setup guide](https://guides.viren070.me/stremio/addons/aiostreams/setup)
- [Real-Debrid INR reseller — 247premiumcart](https://247premiumcart.com/product-category/real-debrid/)
- [Real-Debrid INR reseller — 365premium](https://www.365premium.com/)
- [TRaSH Guides — Prowlarr + FlareSolverr setup](https://trash-guides.info/Prowlarr/prowlarr-setup-flaresolverr/)
- [Bytesized Hosting — Prowlarr Guide 2026](https://bytesized-hosting.com/guides/prowlarr-guide-one-indexer-manager-for-all-your-arr-apps)
- [programming.dev — What indexers do you use](https://programming.dev/post/41364587)
- [InviteHawk — private tracker open signup schedule](https://www.invitehawk.com/topic/135842-private-trackers-signup-schedule-open-application-and-irc-signups/)
- [InstallGentoo wiki — private trackers](https://igwiki.lyci.de/wiki/Private_trackers)
- [BitBrowser — How to access TamilMV 2025](https://www.bitbrowser.net/blog/how-to-access-tamilmv-in-2025)
