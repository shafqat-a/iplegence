# Project Plan: Superior Open IP Intelligence Service

Source: `doc/spec/grok_report.pdf` (image-only; this is the text transcript).
Ready for handoff to implementation. Begin with Phase 0 + Phase 1.

## 1. Project Vision & Goals

Build a high-quality, freely redistributable (or lightly attributed) IP intelligence
database + service that combines the best free/open sources into one superior product.

### Core capabilities

- Country + city-level geolocation (coordinates, timezone, subdivisions)
- ASN + organization name + domain
- Datacenter / hosting / cloud detection
- VPN / proxy / Tor / relay / CDN / anonymous detection
- Clean MMDB output (primary) + CSV/Parquet
- Daily automated updates
- Simple lookup API + offline library support
- Target final MMDB size: **80–150 MB**

### Success criteria

- Better combined coverage than any single free source
- Drop-in compatible with standard MaxMind MMDB readers
- Fully automated daily rebuild + release
- Clear attribution & license compliance
- Easy for others to self-host or consume

## 2. Data Sources (priority order)

| Pri | Source | Provides | Format | Update | License | Merge role |
|-----|--------|----------|--------|--------|---------|------------|
| 1 | IPinfo Lite | Country, continent, ASN, AS name, domain | MMDB/CSV | Daily | CC BY-SA 4.0 (attribution required) | Primary ASN; country fallback |
| 2 | GeoLite2-City | Full city geo (names, lat/lon, timezone, subdivisions) | MMDB | 2×/week | CC BY-SA 4.0 + MaxMind EULA (attribution + account) | Primary geo |
| 3 | GeoLite2-ASN | ASN fallback | MMDB | 2×/week | same | ASN fallback |
| 4 | DB-IP City Lite | Supplementary city data | MMDB | Monthly | CC BY 4.0 | Geo enrichment / conflict resolution |
| 5 | sapics/ip-location-db (`user-country`, `origin-asn`) | Clean country + ASN from public data | CSV/MMDB | Daily | PDDL (public domain) | High-priority country/ASN when available |
| 6 | iptoasn.com | IP→ASN + country | TSV | Hourly | Free | ASN enrichment |
| 7 | OpenProxyDB (NetworkCats) | Proxy / VPN / Tor / hosting / CDN / school flags | CSV | Daily | CC0 | Core privacy/hosting flags |
| 8 | ipapi.is free geo | IPv4/IPv6 geolocation | CSV | Frequent | Free download | Additional geo |
| 9 | Cloud provider ranges (AWS, GCP, Azure, DigitalOcean, Hetzner, OVH, Cloudflare, …) | Official hosting ranges | JSON/CSV | Daily | Public | Strong `is_hosting` / `is_cdn` |
| 10 | Public VPN lists (NordVPN, X4B, iCloud Private Relay, Tor exits, …) | Known VPN/relay exits | Various | Daily | Mixed (mostly public) | `is_vpn` / `is_relay` overlay |
| 11 | QQWry / Chunzhen (optional) | Better China coverage | Binary/CSV | — | Check license | China-specific enrichment |

Prefer sources that already publish MMDB or clean CIDR CSVs. Avoid single-IP
residential proxy dumps that explode size.

## 3. Target output schema (MMDB)

Keep close to MaxMind + IPinfo for compatibility. English-only names unless size
allows more. Drop unreliable fields.

```json
{
  "country": { "iso_code": "US", "names": { "en": "United States" } },
  "continent": { "code": "NA", "names": { "en": "North America" } },
  "city": { "names": { "en": "..." }, "geoname_id": 0 },
  "location": {
    "latitude": 0.0,
    "longitude": 0.0,
    "accuracy_radius": 0,
    "time_zone": "..."
  },
  "subdivisions": [],
  "postal": { "code": "..." },
  "asn": {
    "autonomous_system_number": 15169,
    "autonomous_system_organization": "Google LLC",
    "as_domain": "google.com"
  },
  "traits": {
    "is_anonymous": false,
    "is_anonymous_vpn": false,
    "is_hosting_provider": true,
    "is_public_proxy": false,
    "is_tor_exit_node": false,
    "is_cdn": false,
    "is_relay": false
  }
}
```

Output artifact name: `Superior-IP.mmdb`.

## 4. Architecture & tech stack

- **Language:** Go. MMDB via `github.com/maxmind/mmdbwriter` +
  `github.com/oschwald/maxminddb-golang`.
- **Do not fork** NetworkCats/Merged-IP-Data. Start a clean `iplegence` tree.
  Reference that project's pipeline shape (download → normalize → merge → write)
  but use this schema (`traits`, not `proxy`), official licensed downloads
  (not unofficial GeoLite mirrors), YAML priority config, size abort, and
  generated attribution.
- **Pipeline:** GitHub Actions daily, or cron + Docker.
- **Storage:** GitHub Releases for the MMDB + checksums + attribution.
  Optional later S3/R2 mirror.
- **Serving (Phase 3):** lightweight Go HTTP API, local library wrappers,
  optional mmap cache.

### Merge strategy

1. Download all enabled sources.
2. Normalize everything to CIDR + structured records.
3. Resolve overlapping networks: most-specific prefix wins as the row identity.
4. Field-level priority (configurable):
   - Geo: GeoLite2-City > DB-IP > others
   - ASN: IPinfo Lite > sapics / iptoasn > GeoLite2-ASN
   - Flags: OR — a trait is true if any source says so
5. Collapse adjacent identical records.
6. Write one optimized MMDB (record size 24 or 28).
7. Generate checksums + `ATTRIBUTION.md`.

### Size control

- Prefer CIDR ranges over individual IPs
- English-only names
- Drop low-value fields early
- Monitor size after write; **abort if > 180 MB**
- Phase 1 target **< 100 MB**; final sweet spot **80–150 MB**

## 5. Implementation phases

### Phase 0 — Setup

- Repo exists (`shafqat-a/iplegence`, private)
- Decide base: start fresh with mmdbwriter (locked)
- MaxMind account + license key (GeoLite2)
- IPinfo token (Lite database download)
- Document all licenses & required attributions

### Phase 1 — MVP data pipeline (this plan)

- Automated downloaders for every Phase-1 source
- Basic merge → single MMDB with country + ASN + basic flags
- GitHub Action that runs daily and publishes to Releases
- Validation script (8.8.8.8, 1.1.1.1, known VPN/hosting IPs)
- Target size < 100 MB

### Phase 2 — Full enrichment (follow-on plan)

- City-level data + coordinates (GeoLite2-City, DB-IP, ipapi.is)
- Full OpenProxyDB + cloud ranges + VPN lists
- China enhancement (optional, license-checked)
- Size optimization pass
- Attribution README auto-generated (already in Phase 1; keep)

### Phase 3 — Service layer (follow-on plan)

- HTTP API: lookup by IP → JSON
- Rate limiting + caching
- Docker image + docker-compose
- Client libraries (Go, Python, JS) or document MMDB usage
- Optional Prometheus metrics + health endpoint

### Phase 4 — Quality & operations (follow-on plan)

- Automated accuracy checks against known ground truth
- Diff reports between daily builds
- Size & coverage dashboards
- Mirror strategy (jsDelivr / Cloudflare R2 / GitHub)
- Public “How to use / License / Attribution” page

## 6. Licensing & compliance (non-negotiable)

- Always ship `ATTRIBUTION.md` and `LICENSE` listing every source
- GeoLite2 and IPinfo Lite require visible attribution
- Prefer PDDL / CC0 sources where possible
- Do **not** claim “commercial-grade accuracy”
- If redistributing commercially, review MaxMind EULA carefully
- Never remove or hide source credits

## 7. Success metrics

- Daily successful build rate > 99%
- Final MMDB size 80–150 MB
- Lookup latency < 1 ms (local MMDB)
- Coverage of major cloud providers + top VPN providers
- Clear superiority over single free sources on combined signals

## 8. Risks & mitigations

| Risk | Mitigation |
|------|------------|
| Source license changes | Monitor + have fallback sources |
| Size explosion | Strict CIDR preference + field pruning |
| Conflicting data | Explicit priority config + conflict logging |
| MaxMind account/key issues | Cache last-good GeoLite2 + use alternatives |
| False positives on VPN flags | OR logic + confidence later if needed |
| Maintenance burden | Fully automated pipeline + clear docs |

## 9. Deliverables checklist

- [ ] Daily pipeline that produces `Superior-IP.mmdb`
- [ ] GitHub Releases with MMDB + checksum + attribution
- [ ] Configurable merge priority file
- [ ] Basic lookup CLI + example code
- [ ] Simple API server (Phase 3)
- [ ] Full documentation (sources, licenses, how to rebuild, how to consume)
- [ ] Size & quality reports
