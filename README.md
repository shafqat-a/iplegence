# iplegence

Code is Apache-2.0; bundled/released data remains under each source license listed in `ATTRIBUTION.md`.

Private project: merge the best free/open IP intelligence sources into one
MaxMind-compatible MMDB (`Superior-IP.mmdb`), ship it daily, then add a lookup API.

This is a merged free/open dataset; do not treat it as commercial-grade accuracy.

## What you get

A daily pipeline that downloads official IPinfo Lite, sapics country/ASN,
iptoasn, optional GeoLite2-ASN, OpenProxyDB, and official cloud ranges
(AWS, GCP, Azure, Cloudflare), field-merges them, and writes:

- `dist/Superior-IP.mmdb`
- `dist/Superior-IP.mmdb.sha256`
- `dist/ATTRIBUTION.md`

Phase 2 coverage is country + city + coordinates, ASN, hosting/privacy flags,
Tor exits, and iCloud Private Relay. Lookups also expose an inferred
`traits.usage_type` (`residential` / `mobile` / `business` / `education` /
`government` / `hosting`) from PeeringDB network types plus ASN-name
heuristics, with `traits.usage_type_source` set to `peeringdb`, `asn_name`,
or `prefix_flag`. Empty means unknown. This is not commercial IP2Proxy
`usage_type`.

## Rebuild locally

1. Copy `.env.example` to `.env` and set `IPINFO_TOKEN` (required for a live
   IPinfo Lite download). `MAXMIND_LICENSE_KEY` is optional; GeoLite2-ASN is
   skipped when unset.
2. Load the env vars (`set -a; source .env; set +a`).
3. Run:

```bash
make test
make build
make lookup IP=8.8.8.8
make validate
```

GitHub Actions secrets (Settings → Secrets and variables → Actions):

- `IPINFO_TOKEN` — required for the daily release job
- `MAXMIND_LICENSE_KEY` — optional

Do not commit `.env`, `download/`, `dist/`, or production `*.mmdb` files.

## Docker

One image contains `serve`, `lookup`, `validate`, `build`, the configs, and
the current `Superior-IP.mmdb`.

```bash
docker build -t iplegence:latest .
docker run --rm -p 8080:8080 iplegence:latest
# or: docker compose up --build
# published image:
# docker pull ghcr.io/shafqat-a/iplegence:latest
```

```bash
curl -s http://127.0.0.1:8080/health
curl -s http://127.0.0.1:8080/lookup/8.8.8.8
docker run --rm --entrypoint lookup iplegence:latest 8.8.8.8 /data/Superior-IP.mmdb
```

To rebuild the database inside the image, pass tokens and a writable data dir:

```bash
docker run --rm --entrypoint build \
  -e IPINFO_TOKEN -e MAXMIND_LICENSE_KEY \
  -v "$PWD/dist:/dist" -v "$PWD/download:/download" \
  -v "$PWD/configs:/configs:ro" \
  iplegence:latest
```

## Consume the MMDB

Any MaxMind-compatible reader works. Example with this repo's CLI:

```bash
go run ./cmd/lookup 8.8.8.8 dist/Superior-IP.mmdb
```

In Go, open `Superior-IP.mmdb` with `github.com/oschwald/maxminddb-golang`.
The record shape is nested `country` / `continent` / `asn` / `traits`
(see `doc/spec/superior-open-ip-intelligence.md` §3).

## License & Attribution

- **Code** is licensed under the Apache License 2.0 (`LICENSE`).
- **Data** stays under each source license. A generated `ATTRIBUTION.md` is
  shipped with every release and names every source that contributed.
- This product includes IP data from [IPinfo](https://ipinfo.io) (CC BY-SA 4.0).
- This product includes GeoLite2 Data created by [MaxMind](https://www.maxmind.com)
  (CC BY-SA 4.0 + MaxMind GeoLite2 EULA) when that source is enabled.

Never strip source credits. This is a merged free/open dataset; do not treat
it as commercial-grade accuracy.

## Spec and plan

1. [`AGENTS.md`](AGENTS.md) — locked decisions
2. [`doc/spec/superior-open-ip-intelligence.md`](doc/spec/superior-open-ip-intelligence.md)
3. [`docs/superpowers/plans/2026-08-14-phase0-1-mvp-pipeline.md`](docs/superpowers/plans/2026-08-14-phase0-1-mvp-pipeline.md)
