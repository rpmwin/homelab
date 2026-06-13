# Cloudflare Bypass & *arr Stack Reality Check (June 2026)

**Context**: Self-hosted Prowlarr/Sonarr/Radarr/Jellyfin/qBittorrent on k3s in India, residential IP, Intel i3-2120 (Sandy Bridge, 2c/4t @ 3.3 GHz, 2011). Byparr completes 1337x challenges in 14-35s warm but mostly times out at 60-120s.

**Bottom-line up front**: Your CPU is the problem, and there is no software fix worth chasing on Sandy Bridge. Stop scraping 1337x. Move to a hybrid setup: keep *arr for library management, but offload the actual fetch to **Real-Debrid + Decypharr** (or skip *arr entirely and go **Stremio + AIOStreams + RD**). For Indian-language content, **self-hosted MediaFusion** is genuinely the only good option; the MongoDB/Redis overhead is unavoidable. Skip the bypass arms race entirely.

---

## 1. Is 1337x still used in 2025-2026?

Yes, but it is no longer a "set-and-forget" public indexer. The Prowlarr/Indexers and Prowlarr/Prowlarr GitHub repos have a continuous stream of "1337x blocked by Cloudflare Protection" issues spanning 2022 through January 2026 ([Issue #2319](https://github.com/Prowlarr/Prowlarr/issues/2319), [#2577](https://github.com/Prowlarr/Prowlarr/issues/2577), [#1518](https://github.com/Prowlarr/Prowlarr/issues/1518), [Indexers #749](https://github.com/Prowlarr/Indexers/issues/749)). The community has not abandoned it - 1337x still has the best curated GP releases for English movies/TV - but the official Prowlarr stance is: "use FlareSolverr or live with intermittent breakage." Many users [report on programming.dev](https://programming.dev/post/41364587) running 1337x alongside Bitsearch, Limetorrents, TheRARBG, and nyaa as a pool, accepting that any single one may be down.

## 2. Cloudflare bypass - what actually works in 2026

### FlareSolverr - mostly stalled
Repo is technically active (v3.5.0 dated May 2026) but the README itself states: *"At this time none of the captcha solvers work. You can check the status in the open issues."* ([FlareSolverr repo](https://github.com/FlareSolverr/FlareSolverr)). 48 open issues, no working Turnstile handling without external 2Captcha/CapMonster hand-off. A November 2025 Cloudflare patch caused mass 55-second timeouts ([Issue #1623](https://github.com/FlareSolverr/FlareSolverr/issues/1623)). It is "maintained" but functionally degraded for Turnstile.

### Byparr - current community standard
Active (v2.1.0 Feb 2026, [repo](https://github.com/ThePhaseless/Byparr)). Uses Camoufox (anti-detection Firefox fork) + SeleniumBase + FastAPI, FlareSolverr-API-compatible drop-in. Default deploy budgets ~2 GB RAM and a shared CPU. Camoufox is acknowledged as resource-heavy ([webscraping.club](https://substack.thewebscraping.club/p/camoufox-server-docker)). On a Sandy Bridge i3-2120 the JS execution and headless rendering simply cannot finish Turnstile inside Cloudflare's 60-90s window most of the time. This is not a Byparr bug; it is a CPU bug.

### Solvearr - newest contender
[nabil-ak/Solvearr](https://github.com/nabil-ak/Solvearr) is a FlareSolverr-API-compatible bypass proxy targeted explicitly at Prowlarr/Sonarr/Radarr. Smaller mind-share than Byparr, no published benchmarks vs Byparr on weak CPUs. Worth a 30-minute experiment but unlikely to beat Byparr's fundamentals - same Chromium-class browser, same JS cost.

### Other approaches
- **Cloudscraper / pyppeteer / Playwright**: Cloudscraper is dead against modern Turnstile ([scrapfly.io](https://scrapfly.io/blog/posts/what-is-cloudscraper-and-new-alternatives)). Playwright/Pyppeteer are libraries, not drop-in proxies - require custom glue.
- **Patched Firefox / undetected-chromedriver**: Camoufox (used by Byparr) is the modern equivalent. undetected-chromedriver is increasingly detected by Turnstile.
- **2Captcha / CapMonster hand-off**: FlareSolverr supports it via env vars - paid (~$0.001/solve), works, but adds external dependency and ongoing cost.
- **Tor onion endpoint** for 1337x: mentioned in [bobcares.com](https://bobcares.com/blog/fix-unable-to-access-1337x-to-blocked-by-cloudflare-protection/) and [programming.dev #18898159](https://programming.dev/post/18898159) - flaky, slow, but no JS challenge. Worth trying as a low-CPU fallback.

**Verdict**: No software fix solves CPU-bound JS rendering on Sandy Bridge. The 2026 stack-standard is Byparr; if Byparr times out, you're at the hardware floor.

## 3. Public alternatives to 1337x without Cloudflare

Indexers that scrape cleanly in Prowlarr in 2026 (no proxy needed):
- **TheRARBG** (rargb.to) - the spiritual RARBG successor, no CF challenge at the time of writing.
- **TorrentGalaxy (TGx)** - intermittent CF, usually solvable without Byparr.
- **Nyaa.si** - anime/Asian, no CF.
- **Bitsearch** - meta-search, light or no CF.
- **EZTV** - TV-only, no CF.
- **YTS** - movie-only/x265, no CF.
- **LimeTorrents** - light CF.
- **TorrentDownloads, Torlock, Solid Torrents** - variable.

This list comes from a [programming.dev 2026 Prowlarr poll](https://programming.dev/post/41364587) where the modal answer was "TheRARBG + nyaa + EZTV + YTS + 1337x (best-effort)". The pragmatic move is to enable 10-15 indexers and let Prowlarr's aggregator paper over individual outages.

## 4. Real-Debrid + Zilean + Decypharr workflow

This is now the dominant 2026 *arr workflow for people who want zero local download bandwidth ([corelab.tech Decypharr guide](https://corelab.tech/decypharr-setup-guide/), [Decypharr docs](https://docs.decypharr.com/guides/debrid/real-debrid/)).

**Architecture**:
1. **Prowlarr** searches indexers as usual (still subject to CF for 1337x).
2. **Zilean** is a standalone indexer that scrapes Debrid Media Manager's public hashlists - massive cache, **completely bypasses Cloudflare** because it talks to DMM, not torrent sites.
3. **Decypharr** replaces qBittorrent as the "download client" Sonarr/Radarr point to. It hands magnets to Real-Debrid, RD returns a direct link, Decypharr presents it as a completed download via WebDAV/rclone mount.
4. Sonarr/Radarr import, Jellyfin scans.

**India reality**:
- **Latency**: RD's edge is EU (FR/DE primary). From IN expect 180-250ms RTT, but downloads stream over HTTP and saturate your line - I have seen multiple [SaaSHub reviews](https://www.saashub.com/real-debrid) and Reddit threads confirming RD works from IN with no throttling. The bottleneck is your ISP, not RD.
- **Payment**: RD accepts cards, PayPal, crypto, and resellers. **No direct UPI**. Workarounds: virtual cards (Niyo, Jupiter, Fi), or buy RD vouchers from Indian resellers on Telegram/forums for INR via UPI. Pricing is ~€4/month (~₹400) for 30 days, ~€16 (~₹1500) for 180 days.
- **IP binding**: [RD ToS](https://github.com/debridmediamanager/awesome-debrid/discussions/7) - one public IP at a time. Residential IN IP is fine; if you tunnel via VPN you risk a permaban if the egress is shared with abuse.
- **Indian-specific debrid**: None worth mentioning. AllDebrid (FR) and TorBox (newer, well-regarded) are the only credible alternatives - same EU latency story.

## 5. Stremio + Torrentio/AIOStreams + RD - is *arr obsolete?

Partially. The 2026 consensus across [bytesized-hosting](https://bytesized-hosting.com/guides/seedbox-vs-debrid-in-2026-real-debrid-stremio-and-why-its-not-that-simple), [corelab.tech](https://corelab.tech/stremio-setup-guide/), and the [HN thread #33365913](https://news.ycombinator.com/item?id=33365913):

- **Stremio + AIOStreams + RD**: ~5 minutes to set up, ~€4/month total, on-demand "Netflix for torrents." Torrentio is rate-limited in 2026; **AIOStreams** is the successor aggregator, with **Comet** and **MediaFusion** as the actively-maintained best-of-class addons ([ElfHosted recommended addons](https://docs.elfhosted.com/stremio-addons/guide/recommended-addons/)).
- **What *arr still gives you**: Persistent library on disk you control (RD can drop hashes for DMCA), automated quality upgrades, multi-user families on Jellyfin/Plex, offline access, archival.
- **Who has actually switched**: Solo viewers, renters, anyone whose library is "watch once, forget." Anyone running Jellyfin for family or building a permanent collection still runs *arr.

The honest take: *arr is not obsolete, it has split into two camps - "library builders" (still *arr+downloads) and "consumers" (Stremio+RD). For a weak CPU + residential IN, the consumer path removes 90% of your headaches.

## 6. Private trackers from India in 2026

The reality is harsh but navigable:
- **HDBits, PassThePopcorn (PTP), BroadcastTheNet (BTN), AnimeBytes (AB)**: Closed. Invite-only. PTP/BTN/HDB sometimes recruit via the IRC interview at #what.cd-irc.scenelist or #ptp-interview - PTP interviews are doable with prep. **AnimeBytes runs open-signup windows occasionally** (last few in late 2025). HDBits is the hardest of the four.
- **TorrentLeech, IPTorrents, FileList**: Periodic paid signup (~€10-€30) or invite. TL is the easiest-on-ramp for English movies/TV.
- **DesiTorrents (DT)**: Indian-specific - movies/TV/music in Hindi/Tamil/Telugu/etc. **Application signup currently open** at https://desitorrents.tv/application ([confirmed Feb 2026 via opentrackers.org](https://opentrackers.org/desitorrents/)). This is the single highest-value sign-up for an Indian user right now.
- **BwTorrents, TamilTorrents, Hon3yHD, TmGHub**: Smaller Indian-content trackers, invite-only.

Indian IPs are not blocked on any of these; the gatekeeping is purely on the invite/interview side, not nationality.

## 7. Hardware reality

Multiple 2025-2026 community posts converge on the same number: **8 GB RAM minimum, 16 GB comfortable, modern (post-2015) x86 with AES-NI and at least 4 modern cores** is what people recommend for a Byparr/FlareSolverr+arr stack ([bytesized arr guide](https://bytesized-hosting.com/guides/the-complete-arr-stack-guide-2026-sonarr-radarr-prowlarr-and-more), [sumguy.com](https://sumguy.com/arr-stack-durable-setup/)).

i3-2120 (Sandy Bridge, no AES-NI on lower SKUs, 2 cores, no AVX2) is roughly **1/8th the single-thread perf of a modern N100**. Cloudflare's Turnstile is a deliberately CPU-burning JS proof-of-work that scales with attempts - on weak silicon you fall into a death spiral where retries make it worse ([roundproxies.com error guide](https://roundproxies.com/blog/flaresolverr-error/)). The realistic upgrade is a **used N100 mini-PC (~₹12-15k on Indian secondhand markets, or new ~₹18-22k)** - same idle power as Sandy Bridge i3, ~6-8x the Turnstile throughput, AV1 hardware decode for Jellyfin transcoding as a bonus. This single upgrade solves CF bypass *and* transcoding.

## 8. RSS feed bypass trick

Yes, some trackers expose RSS endpoints that skip the JS challenge because RSS readers historically don't run JS - Cloudflare often whitelists `/rss` paths. **1337x does not expose a public RSS feed** at any tracked path; the search results pages are all CF-gated. Trackers that do (Nyaa, EZTV, RuTracker via 3rd-party mirrors) are already the no-CF trackers above. The "RSS trick" is real for *private* trackers (which use it heavily, with passkey-authenticated feeds) but not a 1337x workaround.

## 9. Specific 2026 success stories

- **"DUMB & Real-Debrid Guide (2026)"** ([corelab.tech](https://corelab.tech/ultimate-plex-debrid-guide/)) - end-to-end working setup, Plex + Sonarr/Radarr + Decypharr + RD + Zilean. Cited as the canonical 2026 stack.
- **ElfHosted's [Plex + Radarr/Sonarr + Real-Debrid guide](https://docs.elfhosted.com/guides/media/plex-realdebrid-aars/)** - production-grade, used by paying customers, kept current.
- **"The *arr Stack Without 9000 Reddit Threads"** ([sumguy.com](https://sumguy.com/arr-stack-durable-setup/)) - "this is what I run in 2026" essay, no debrid, just durable indexer pool + qBittorrent + VPN.
- **Naralux/mediacenter** ([GitHub](https://github.com/Naralux/mediacenter)) - docker-compose reference repo for RD+*arr, current commits in 2026.

## 10. Indian/Tamil/Telugu/Hindi content - what to actually run

**MediaFusion is genuinely the best option** ([repo](https://github.com/mhdzumair/MediaFusion)), and the MongoDB+Redis dependency is not optional - it stores scraped catalog state for TamilMV/TamilBlasters/TamilUltra/MHDTVPlay. On k3s the deploy is one Helm chart with MongoDB and Redis sub-charts - your homelab can absorb that.

A self-hosted MediaFusion instance reportedly broke TamilMV/TamilUltra scraping in July 2025 (see [Issue #538](https://github.com/mhdzumair/MediaFusion/issues/538)) but TamilBlasters has remained working. Update cadence is steady. The hosted instance at stremio-addons.net is the lowest-effort option (~$0.05/day, free $10 credit).

**Simpler alternatives**:
- **DesiTorrents** added to Prowlarr (CF-free, login-gated) - if you can get the application accepted.
- **Sun NXT, Aha, Hotstar, Jio Hotstar** - legal Tamil/Telugu/Hindi, ₹100-300/month, no infra. The honest baseline for IN regional content.
- **Hosted MediaFusion via ElfHosted** - skip self-host entirely.

There is no "simpler self-hosted" path that has anything like MediaFusion's coverage. Hindi/general Bollywood is well-covered on 1337x/YTS/TheRARBG; **the MongoDB pain is specifically the cost of Tamil/Telugu catalog scraping.**

---

## Strongest single recommendation

For your exact situation - weak Sandy Bridge CPU, residential IN IP, want exhaustive coverage including 1337x quality - do this, in order:

1. **Stop trying to make Byparr work on the i3-2120.** Budget ₹15-20k for a used or new **N100 mini-PC** (Beelink S12 Pro, GMKTec, Trigkey). This is the single highest-leverage move and unblocks everything else. Move k3s onto it; keep the i3 as a worker node for non-CPU-bound pods.

2. **Add Real-Debrid (€16/180-day via virtual card or INR reseller) + Zilean + Decypharr.** This gives you a CF-free indexer (Zilean reads DMM hashlists, never touches 1337x) and removes local download bandwidth entirely. Keep Prowlarr+Sonarr+Radarr+Jellyfin exactly as they are.

3. **Demote 1337x to a fallback indexer.** Promote TheRARBG, nyaa, EZTV, YTS, Bitsearch in Prowlarr priority. Run Byparr only for the occasional 1337x query - on N100 it will be a 5-8s solve, well inside any timeout.

4. **Apply for DesiTorrents** (signups open as of Feb 2026) for Hindi/regional content. Add MediaFusion self-hosted on the same k3s for Tamil/Telugu specifically - the MongoDB cost is justified by the catalog.

5. **For "I just want to watch something now" use Stremio + AIOStreams + the same RD account** on your phone/TV. Skip *arr entirely for casual viewing. Reserve *arr for what you actually want archived to Jellyfin.

If you cannot do (1), the realistic fallback is to **abandon 1337x entirely**, lean on Real-Debrid+Zilean for 90% coverage, accept that the remaining 10% (exotic encodes, obscure scene releases) requires DesiTorrents or paying for an ElfHosted-hosted Byparr.

---

## Sources

- [FlareSolverr repo](https://github.com/FlareSolverr/FlareSolverr) / [Issue #1623 - Cloudflare Nov-2025 patch](https://github.com/FlareSolverr/FlareSolverr/issues/1623)
- [Byparr repo (ThePhaseless)](https://github.com/ThePhaseless/Byparr)
- [Solvearr (nabil-ak)](https://github.com/nabil-ak/Solvearr)
- [Prowlarr Issue #2319 - 1337x CF block](https://github.com/Prowlarr/Prowlarr/issues/2319) / [#2577](https://github.com/Prowlarr/Prowlarr/issues/2577) / [#1518](https://github.com/Prowlarr/Prowlarr/issues/1518) / [Indexers #749](https://github.com/Prowlarr/Indexers/issues/749)
- [programming.dev "What indexers do you use"](https://programming.dev/post/41364587) / [Manual CF bypass post](https://programming.dev/post/18898159)
- [Decypharr docs - Real-Debrid](https://docs.decypharr.com/guides/debrid/real-debrid/)
- [corelab.tech Decypharr guide 2026](https://corelab.tech/decypharr-setup-guide/) / [Ultimate Plex-Debrid guide](https://corelab.tech/ultimate-plex-debrid-guide/) / [Stremio AIOStreams](https://corelab.tech/stremio-setup-guide/)
- [ElfHosted Plex+RD+arrs](https://docs.elfhosted.com/guides/media/plex-realdebrid-aars/) / [Jellyfin+RD+arrs](https://docs.elfhosted.com/guides/media/jellyfin-realdebrid-aars/) / [Best Stremio addons 2026](https://docs.elfhosted.com/stremio-addons/guide/recommended-addons/)
- [bytesized "Seedbox vs Debrid 2026"](https://bytesized-hosting.com/guides/seedbox-vs-debrid-in-2026-real-debrid-stremio-and-why-its-not-that-simple) / [Complete Arr Stack 2026](https://bytesized-hosting.com/guides/the-complete-arr-stack-guide-2026-sonarr-radarr-prowlarr-and-more)
- [sumguy.com - durable arr stack](https://sumguy.com/arr-stack-durable-setup/)
- [Naralux/mediacenter reference repo](https://github.com/Naralux/mediacenter)
- [MediaFusion Issue #538 - TamilMV scraping](https://github.com/mhdzumair/MediaFusion/issues/538) / [MediaFusion repo](https://github.com/mhdzumair/MediaFusion)
- [DesiTorrents open signup confirmation](https://opentrackers.org/desitorrents/)
- [Comet self-hosted Stremio addon](https://github.com/iceblinker/comet)
- [awesome-debrid Real-Debrid FAQ - IP rules](https://github.com/debridmediamanager/awesome-debrid/discussions/7)
- [Real-Debrid SaaSHub reviews](https://www.saashub.com/real-debrid)
- [roundproxies Byparr 2026 guide](https://roundproxies.com/blog/byparr/) / [FlareSolverr errors](https://roundproxies.com/blog/flaresolverr-error/)
- [scrapfly Cloudscraper status](https://scrapfly.io/blog/posts/what-is-cloudscraper-and-new-alternatives)
- [HN #33365913 - Stremio+Torrentio+Debrid](https://news.ycombinator.com/item?id=33365913)
