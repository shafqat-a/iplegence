# iplegence

One MaxMind-compatible database built from free and open IP sources, plus a lookup you can call with `curl`.

iplegence downloads country, city, ASN, and privacy feeds, merges them by the most specific prefix, and writes `Superior-IP.mmdb`. Look up an address and you get country, city, coordinates, ASN, and traits such as hosting, CDN, Tor exit, and an inferred usage type.

The database is a merge of public feeds. Treat it as a strong open dataset, and check it against your own traffic before you depend on it for fraud or compliance decisions. `traits.usage_type` is inferred from PeeringDB and ASN names. An empty value means unknown. It is not a commercial usage-type feed.

Code is Apache-2.0. The data stays under each source's license. Every build writes `ATTRIBUTION.md` next to the database. Keep that file with the database.

## Look up an IP

The HTTP server listens on `:8080` and reads `dist/Superior-IP.mmdb` (override with `LISTEN_ADDR` and `MMDB_PATH`).

```bash
make serve
```

```bash
curl -s http://127.0.0.1:8080/health
curl -s http://127.0.0.1:8080/lookup/8.8.8.8
curl -s http://127.0.0.1:8080/1.1.1.1
```

A hit is JSON. Empty fields are omitted, and a trait flag appears only when it is true. A miss is `404 {"error":"not found"}`. A bad address is `400 {"error":"bad ip"}`.

`testdata/golden.json` checks `8.8.8.8` as country `US`, city `Mountain View`, ASN `15169`. A response with those fields looks like this. Extra traits show up when a source has them, and they are omitted when they do not:

```json
{
  "country": { "iso_code": "US", "names": { "en": "United States" } },
  "city": { "names": { "en": "Mountain View" } },
  "asn": { "autonomous_system_number": 15169 }
}
```

`usage_type` is one of `residential`, `mobile`, `business`, `education`, `government`, `hosting`. `usage_type_source` is `peeringdb`, `asn_name`, or `prefix_flag`.

There is no auth on this port. Put it behind your own proxy if it is reachable beyond localhost.

The same record from the CLI, without the server:

```bash
make lookup IP=8.8.8.8
# or: go run ./cmd/lookup 8.8.8.8 dist/Superior-IP.mmdb
```

### From Go

Any MaxMind reader works. This matches the reader the server uses (`github.com/oschwald/maxminddb-golang/v2`):

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"os"

	"github.com/oschwald/maxminddb-golang/v2"
)

func main() {
	db, err := maxminddb.Open("dist/Superior-IP.mmdb")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	addr, err := netip.ParseAddr(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rec map[string]any
	if err := db.Lookup(addr).Decode(&rec); err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rec); err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, "ok")
}
```

### From Python

```bash
pip install maxminddb
```

```python
import maxminddb

with maxminddb.open_database("dist/Superior-IP.mmdb") as reader:
    print(reader.get("8.8.8.8"))
```

Field names inside the database are the MaxMind-style keys in `internal/schema/read.go`: `country.iso_code`, `location.latitude`, `asn.autonomous_system_number`, `traits.is_tor_exit_node`, `traits.usage_type`.

### Docker

The image contains `serve`, `lookup`, `validate`, `build`, and a copy of `Superior-IP.mmdb`.

```bash
docker build -t iplegence:latest .
docker run --rm -p 8080:8080 iplegence:latest

docker run --rm --entrypoint lookup iplegence:latest 8.8.8.8 /data/Superior-IP.mmdb
```

`docker compose up --build` publishes the same port. A published image is `ghcr.io/shafqat-a/iplegence:latest`.

## What is in the record

| Field | Contents |
|---|---|
| `country` | `iso_code`, `geoname_id`, `names.en` |
| `continent` | `code`, `geoname_id`, `names.en` |
| `city` | `geoname_id`, `names.en` |
| `location` | `latitude`, `longitude`, `accuracy_radius`, `time_zone` |
| `subdivisions` | region list, same name shape as country |
| `postal.code` | postal code when a source has one |
| `asn` | number, organization, `as_domain` |
| `traits` | anonymous, VPN, hosting, public proxy, Tor exit, CDN, iCloud Private Relay, usage type |

Names are English only.

## Rebuild the database

Go 1.25. Live downloads need an IPinfo token. GeoLite2 is included only when `MAXMIND_LICENSE_KEY` is set. Both names are in `.env.example`.

```bash
cp .env.example .env          # set IPINFO_TOKEN
set -a; source .env; set +a
make test
make build                    # writes dist/Superior-IP.mmdb and ATTRIBUTION.md
make validate
```

Do not commit `.env`, `download/`, or `dist/`.

The daily GitHub Action publishes a dated release. Its secrets are `IPINFO_TOKEN` (required) and `MAXMIND_LICENSE_KEY` (optional).

Sources are listed in `configs/sources.yaml`, including IPinfo Lite, sapics country/ASN, iptoasn, DB-IP City Lite, ipapi.is, official cloud ranges (AWS, GCP, Azure, Cloudflare, and others), Tor exits, proxy and VPN lists, and PeeringDB. A source marked `required: false` is skipped when its token or download is missing. Country and ASN fields follow `configs/priority.yaml`. Trait flags are OR-merged.

Rebuild inside Docker when you want the image to produce a fresh file:

```bash
docker run --rm --entrypoint build \
  -e IPINFO_TOKEN -e MAXMIND_LICENSE_KEY \
  -v "$PWD/dist:/dist" -v "$PWD/download:/download" \
  -v "$PWD/configs:/configs:ro" \
  iplegence:latest
```

## Develop

```bash
make test          # go test ./...
```

| Path | What it is |
|---|---|
| `cmd/build` | Download, merge, write the MMDB |
| `cmd/serve` | HTTP lookup |
| `cmd/lookup` | One-shot CLI |
| `cmd/validate` | Compare the MMDB with `testdata/golden.json` |
| `internal/schema` | Record shape |
| `doc/spec/superior-open-ip-intelligence.md` | Field spec |

Packages under `internal/` are not an importable SDK. Call the HTTP API, the CLI, or a MaxMind reader.

## License

- Code: [Apache License 2.0](LICENSE).
- Data: each upstream license, named in the generated `ATTRIBUTION.md`.
- IPinfo data is [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/).
- GeoLite2 data, when that source is enabled, is created by [MaxMind](https://www.maxmind.com) and is covered by CC BY-SA 4.0 and the GeoLite2 EULA.

Ship the attribution file with every copy of the database.
