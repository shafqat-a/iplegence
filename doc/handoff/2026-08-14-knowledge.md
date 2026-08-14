# Knowledge export — 2026-08-14 session

This is everything the first session knew that a second machine needs.
It is not a substitute for the spec or the plan; it is the context around them.

## Session facts

- **When:** 2026-08-14 (UTC afternoon / BD evening)
- **Grok session id:** `01a000cd-ad0d-7963-9355-57e14518e3e9`
- **Model:** grok-4.6, reasoning effort high
- **GitHub owner:** `shafqat-a`
- **Repo:** private `https://github.com/shafqat-a/iplegence`
- **Clone path on first machine:** `/home/shafqat/git/iplegence`
- **Git protocol:** SSH (`git@github.com:shafqat-a/iplegence.git`)
- **gh auth on first machine:** logged in as `shafqat-a` via keyring, scopes include `repo`

## User turns (verbatim intent)

1. Create a private GitHub project named `iplegence` and clone it under `~/git`.
2. Switch to the repo directory.
3. Copy `~/Downloads/grok_report.pdf` into `./doc/spec`.
4. Read the PDF as the spec and plan the project.
5. “I will run this on another machine. Export all knowledge and the session chat, then push into git.”

## What was created on disk

| Path | What |
|------|------|
| `doc/spec/grok_report.pdf` | Original 8-page image spec (2.5 MB, jsPDF 4.0.0, created 2026-08-14 21:30 +06). `pdftotext` extracts nothing. |
| `doc/spec/superior-open-ip-intelligence.md` | Full text transcript of that PDF. |
| `docs/superpowers/plans/2026-08-14-phase0-1-mvp-pipeline.md` | 12-task Phase 0+1 implementation plan. |
| `AGENTS.md` | Next-agent handoff. |
| `README.md` | New-machine entry point. |
| `doc/session/chat.md` | Readable conversation. |
| `doc/session/chat.jsonl` | Structured user/assistant messages from this session only. |
| `.gitignore` | Ignores `download/`, `dist/`, `.env`, production `*.mmdb`. |

## Product in one paragraph

Build a freely redistributable (attribution-required) IP intelligence database
by merging the best free/open sources. Primary artifact is a MaxMind-compatible
MMDB named `Superior-IP.mmdb` (80–150 MB) with country/city geo, ASN+org+domain,
and hosting/VPN/proxy/Tor/CDN/relay flags. Daily automated rebuild and GitHub
Release. Later: lookup API and self-host docs.

## Spec highlights the plan already encodes

Core capabilities: country+city geo, ASN, hosting/cloud detection, VPN/proxy/Tor/relay/CDN/anonymous flags, MMDB + later CSV/Parquet, daily updates, lookup API + offline libs.

Sources (priority order): IPinfo Lite, GeoLite2-City, GeoLite2-ASN, DB-IP City Lite, sapics user-country + origin-asn, iptoasn.com, OpenProxyDB, ipapi.is, official cloud ranges, public VPN lists, optional QQWry.

Merge: download → CIDR+records → most-specific prefix → field priority (geo: GeoLite2-City > DB-IP > others; ASN: IPinfo > sapics/iptoasn > GeoLite2-ASN; flags OR) → collapse adjacent identical → write MMDB (record size 24 or 28) → checksums + attribution.

Size: prefer CIDRs, English-only names, abort > 180 MB.

Phases: 0 setup, 1 MVP pipeline (country+ASN+basic flags, <100 MB), 2 full enrichment, 3 service layer, 4 quality/ops.

The PDF itself says: hand this to Grok 4.6 and **begin with Phase 0 + Phase 1**.

## Research done while planning (so you do not redo it)

### NetworkCats/Merged-IP-Data

- Public Go merger: https://github.com/NetworkCats/Merged-IP-Data
- Module `merged-ip-data`, Go 1.25.6
- Deps: `mmdbwriter v1.2.0`, `oschwald/maxminddb-golang v1.13.1`, `go4.org/netipx`, `ipipdotnet/ipdb-go`
- Layout: `cmd/merge/main.go`, `internal/{config,downloader,interner,merger,reader,writer}`
- Flags: `-skip-download`, `-output` (default `Merged-IP.mmdb`)
- Download timeout 30 minutes, concurrency 7
- Config hardcodes unofficial GeoLite URLs (`P3TERX/GeoLite.mmdb`) and a NetworkCats IPinfo mirror
- Output schema uses a `proxy` object, not spec `traits`
- `writer.WriteToPath` does atomic `.tmp` rename + 4 MB buffer
- `MergedRecord.ToMMDBType` omits empty/false fields and interns strings

### IPinfo Lite

- License: CC BY-SA 4.0, attribution required
- Official MMDB: `https://ipinfo.io/data/ipinfo_lite.mmdb?token=$TOKEN`
- Fields typically: country / country_code, continent / continent_code, asn, as_name, as_domain
- Token required even for the free database

### MaxMind GeoLite2

- Needs a free account + license key
- Official: `https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=$KEY&suffix=tar.gz`
- Same pattern for `GeoLite2-City`
- Attribution: “This product includes GeoLite2 Data created by MaxMind, available from https://www.maxmind.com.”
- Do **not** use P3TERX/git.io redistributes as the primary source

### sapics/ip-location-db

- Daily GitHub Releases under `/releases/download/latest/`
- Plan maps spec “user-country” → `geo-whois-asn-country.mmdb` (PDDL)
- `origin-asn.mmdb` for public-domain ASN
- License: PDDL (public domain) for those datasets — prefer them

### iptoasn.com

- Hourly TSV: `range_start range_end AS_number country_code AS_description`
- Combined: `https://iptoasn.com/data/ip2asn-combined.tsv.gz`
- Fallback: `ip2asn-v4.tsv.gz` and `ip2asn-v6.tsv.gz`
- Skip ASN 0

### OpenProxyDB

- `https://github.com/NetworkCats/OpenProxyDB/releases/latest/download/proxy_blocks.csv`
- CC0
- Map CSV flags onto spec traits (`is_vpn` → `is_anonymous_vpn`, `is_tor` → `is_tor_exit_node`, `is_hosting` → `is_hosting_provider`, `is_proxy` → `is_public_proxy`). Ignore `is_school` in Phase 1.

### Cloud ranges

- AWS JSON prefixes + ipv6_prefixes
- GCP `cloud.json` ipv4Prefix / ipv6Prefix
- Azure Service Tags download id 56519 (URL on the details page rotates; parse it; source is optional)
- Cloudflare plaintext CIDR lists
- All set `is_hosting_provider` and/or `is_cdn`

### First machine toolchain

- `go version go1.26.4 linux/amd64` was installed
- `gh` worked with SSH git
- `pdfplumber` was **not** installed; PDF was read by rendering pages

## Execution choice left open

After the plan was saved, the user was offered:

1. Subagent-driven (recommended)
2. Inline execution

They did not pick. They asked to export and push instead. The next session should ask again, or default to subagent-driven if they say “implement” / “execute the plan”.

## Things that must not leak into git

- `/home/shafqat/.env` (homelab secrets: ESXi, iLO, OPNsense — unrelated to this product)
- `IPINFO_TOKEN`, `MAXMIND_LICENSE_KEY`
- Other Grok sessions on that machine
- System prompt / internal product guidelines

## Suggested first commands on the new machine

```bash
git clone git@github.com:shafqat-a/iplegence.git
cd iplegence
# confirm you have Go 1.24+ and gh auth
go version
gh auth status
# then execute the plan, starting at Task 1
```
