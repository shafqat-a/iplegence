# AGENTS.md — iplegence handoff

Updated 2026-08-14 after Phase 0+1 implementation.

## What this is

**iplegence** is a private GitHub repo (`shafqat-a/iplegence`) for a
**Superior Open IP Intelligence** product: merge free/open IP geolocation and
privacy/hosting datasets into one MaxMind-compatible MMDB, publish it daily,
then serve lookups.

The original spec is `doc/spec/grok_report.pdf` (jsPDF, 8 image pages, no
extractable text). A full transcript is
`doc/spec/superior-open-ip-intelligence.md`.

## What is already done

1. Private repo created (`shafqat-a/iplegence`).
2. Spec transcribed; Phase 0+1 plan written.
3. Phase 0+1 Go pipeline implemented and pushed to `main` (12 commits after
   the docs-only bootstrap). `go test ./...` passes locally.
4. Artifacts produced by `make build`: `dist/Superior-IP.mmdb`,
   `dist/Superior-IP.mmdb.sha256`, `dist/ATTRIBUTION.md`.
5. Lookup CLI (`cmd/lookup`), golden-IP validator (`cmd/validate`),
   CI (`.github/workflows/ci.yml`), daily release (`.github/workflows/daily.yml`).

**No live production MMDB has been built yet.** There is no local `.env` and
no GitHub Actions secrets, so IPinfo Lite cannot be downloaded.

## What to do next

1. Create `.env` from `.env.example` with `IPINFO_TOKEN` (required) and
   optional `MAXMIND_LICENSE_KEY`.
2. Add the same values as GitHub Actions secrets
   (Settings → Secrets and variables → Actions).
Phase 1 live build works. Next optional work is a new plan for Phase 2+
(city/VPN), Phase 3 (HTTP API), or Phase 4 (dashboards). Do not start those
until asked.

Daily 01:00 UTC GitHub Actions still needs `IPINFO_TOKEN` (already set).

## Locked decisions (do not reopen unless the owner asks)

| Topic | Decision |
|-------|----------|
| Language | Go 1.24+ (`mmdbwriter` + `maxminddb-golang/v2`). `go.mod` is 1.25 because `maxminddb-golang/v2 v2.5.0` requires it. |
| Module path | `github.com/shafqat-a/iplegence` |
| Base | **Start fresh.** Do not fork NetworkCats/Merged-IP-Data. Reference its pipeline shape only. |
| Output name | `Superior-IP.mmdb` + `.sha256` + `ATTRIBUTION.md` |
| Schema | Spec `traits` object (not NetworkCats `proxy`) |
| Trait fields | `is_anonymous`, `is_anonymous_vpn`, `is_hosting_provider`, `is_public_proxy`, `is_tor_exit_node`, `is_cdn`, `is_relay` |
| Names | English only (`names.en`) |
| Downloads | Official IPinfo + official MaxMind only. No P3TERX / unofficial GeoLite mirrors. |
| Merge | Most-specific prefix = row identity; field priority from YAML; flags OR |
| Country priority (Phase 2) | GeoLite2-City > DB-IP City > sapics > IPinfo Lite > iptoasn |
| ASN priority | IPinfo Lite > sapics origin-asn / iptoasn > GeoLite2-ASN |
| Size | Abort if > 180 MB. Phase 1 target < 100 MB. Final sweet spot 80–150 MB. |
| Phase 1 sources | IPinfo Lite, sapics origin-asn + geo-whois-asn-country, iptoasn, optional GeoLite2-ASN, OpenProxyDB, AWS/GCP/Azure/Cloudflare ranges |
| Out of Phase 1 | GeoLite2-City, DB-IP City, ipapi.is, VPN lists, QQWry, HTTP API |
| Code license | Apache-2.0. Data remains per-source. |
| Copy rule | Never claim “commercial-grade accuracy”. Never hide source credits. |

## Secrets (not in this repo)

Create `.env` locally (gitignored) and GitHub Actions secrets:

- `IPINFO_TOKEN` — required for a live IPinfo Lite download
- `MAXMIND_LICENSE_KEY` — optional in Phase 1; GeoLite2-ASN is skipped if unset

Do not commit tokens. Do not commit `download/` or `dist/` or `*.mmdb` except fixtures under `testdata/`.

## Useful URLs

| What | URL |
|------|-----|
| This repo | https://github.com/shafqat-a/iplegence |
| IPinfo Lite download | `https://ipinfo.io/data/ipinfo_lite.mmdb?token=$IPINFO_TOKEN` |
| MaxMind GeoLite2-ASN | `https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=$MAXMIND_LICENSE_KEY&suffix=tar.gz` |
| sapics origin-asn | https://github.com/sapics/ip-location-db/releases/download/latest/origin-asn.mmdb |
| sapics country | https://github.com/sapics/ip-location-db/releases/download/latest/user-country.mmdb |
| iptoasn combined | https://iptoasn.com/data/ip2asn-combined.tsv.gz (fallback: `ip2asn-v4.tsv.gz` + `ip2asn-v6.tsv.gz`) |
| OpenProxyDB | https://github.com/NetworkCats/OpenProxyDB/releases/latest/download/proxy_blocks.csv |
| AWS ranges | https://ip-ranges.amazonaws.com/ip-ranges.json |
| GCP ranges | https://www.gstatic.com/ipranges/cloud.json |
| Azure Service Tags | Microsoft download id 56519 |
| Cloudflare v4/v6 | https://www.cloudflare.com/ips-v4 and `/ips-v6` |
| Reference pipeline (do not fork) | https://github.com/NetworkCats/Merged-IP-Data |
| mmdbwriter | https://github.com/maxmind/mmdbwriter |

## Reference implementation notes

NetworkCats/Merged-IP-Data is Go 1.25, layout `cmd/merge` + `internal/{config,downloader,interner,merger,reader,writer}`.
It already merges GeoLite2, IPinfo, sapics, OpenProxyDB, VPN lists, etc., and writes MMDB.

Do **not** copy it wholesale because:

- Output schema uses `proxy` (is_proxy, is_vpn, is_tor, is_hosting, is_cdn, is_school, is_anonymous), not spec `traits`.
- GeoLite2 is pulled from unofficial P3TERX mirrors (EULA/license risk).
- IPinfo is pulled from a third-party GitHub mirror.
- Priorities are hardcoded, no YAML, no 180 MB abort, no generated `ATTRIBUTION.md`.
- Output file is `Merged-IP.mmdb`, not `Superior-IP.mmdb`.

Steal the ideas: download → normalize to CIDR → field merge → collapse adjacent → `mmdbwriter` atomic write.

## Session provenance

- Date: 2026-08-14
- First machine workspace: `/home/shafqat`
- Implementation workspace: `/home/shafqat/git/iplegence`
- First Grok session id: `01a000cd-ad0d-7963-9355-57e14518e3e9`
- Chat export: `doc/session/chat.md`
- Knowledge dump: `doc/handoff/2026-08-14-knowledge.md`

## Operator context (not part of the product, just where this started)

The first session ran on the cloudlabs homelab operator account (`shafqat`). That
environment’s infra notes live in `~/AGENTS.md` on that machine (ESXi / OPNsense /
cloudlabs.live). They are **not** required to implement iplegence and were not
copied here.
