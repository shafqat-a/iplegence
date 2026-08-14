# Phase 0+1 MVP Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a daily Go pipeline that downloads free/open IP sources, field-merges them, and publishes `Superior-IP.mmdb` (country + ASN + hosting/privacy flags, < 100 MB) with checksums, attribution, a lookup CLI, and a validator.

**Architecture:** Fresh Go module (do not fork NetworkCats/Merged-IP-Data). Each source adapter emits `[]source.Block` (`netip.Prefix` + partial `schema.Record`). A sweep-line merger walks IPv4 and IPv6 separately, field-merges overlapping ranges using `configs/priority.yaml`, collapses adjacent identical records, and writes one MMDB via `mmdbwriter`. Flags OR. Abort if the file is larger than 180 MB.

**Tech Stack:** Go 1.24+, `github.com/maxmind/mmdbwriter`, `github.com/oschwald/maxminddb-golang/v2`, `go4.org/netipx`, `gopkg.in/yaml.v3`, GitHub Actions, GitHub Releases.

## Global Constraints

- Output artifact name is exactly `Superior-IP.mmdb` (plus `Superior-IP.mmdb.sha256` and `ATTRIBUTION.md`).
- MMDB schema must match `doc/spec/superior-open-ip-intelligence.md` §3: nested `country` / `continent` / `asn` / `traits`. Use `traits`, not NetworkCats' `proxy` object.
- Trait field names (exact): `is_anonymous`, `is_anonymous_vpn`, `is_hosting_provider`, `is_public_proxy`, `is_tor_exit_node`, `is_cdn`, `is_relay`.
- English-only names (`names.en` only). Do not store de/es/fr/ja/zh-CN.
- Prefer CIDR ranges. Never ingest single-IP residential proxy dumps.
- Abort the build if `Superior-IP.mmdb` is **> 180 MB**. Phase 1 target **< 100 MB**.
- ASN priority: IPinfo Lite > sapics origin-asn / iptoasn > GeoLite2-ASN.
- Country priority in Phase 1 (no GeoLite2-City yet): sapics country > IPinfo Lite > iptoasn.
- Trait flags use OR: true if any source says true.
- Always generate `ATTRIBUTION.md` listing every source that contributed. Never strip credits.
- Do not claim “commercial-grade accuracy” in README, release notes, or API copy.
- Official downloads only: MaxMind via `MAXMIND_LICENSE_KEY`, IPinfo via `IPINFO_TOKEN`. Do not use P3TERX or other unofficial GeoLite redistributes.
- Phase 1 sources only: IPinfo Lite, sapics origin-asn + country, iptoasn, GeoLite2-ASN (optional if key missing), OpenProxyDB, official cloud ranges (AWS, GCP, Azure, Cloudflare). No city DBs, no VPN lists, no QQWry.
- GeoLite2 and IPinfo Lite require visible attribution (CC BY-SA 4.0). sapics PDDL, OpenProxyDB CC0, cloud ranges public.
- Module path: `github.com/shafqat-a/iplegence`.
- Do not commit downloaded databases, `dist/`, or secrets. `.env` is gitignored.

## Follow-on plans (do not implement in this plan)

| Later plan | Scope |
|------------|--------|
| Phase 2 Full Enrichment | GeoLite2-City, DB-IP City, ipapi.is, VPN/relay lists, optional QQWry |
| Phase 3 Service Layer | HTTP lookup API, rate limit, Docker, client docs |
| Phase 4 Quality & Ops | Ground-truth checks, daily diffs, coverage dashboard, R2/jsDelivr mirror |

Phase 1 country without GeoLite2-City uses sapics → IPinfo → iptoasn. Phase 2 replaces that with GeoLite2-City > DB-IP > others.

---

## File structure

| Path | Responsibility |
|------|----------------|
| `go.mod` / `go.sum` | Module + deps |
| `Makefile` | `test`, `build`, `lookup`, `validate` |
| `.gitignore` | `download/`, `dist/`, `.env` |
| `.env.example` | `IPINFO_TOKEN`, `MAXMIND_LICENSE_KEY` |
| `LICENSE` | Apache-2.0 for **code** (data stays per-source) |
| `README.md` | Rebuild, consume, license, no accuracy claims |
| `configs/sources.yaml` | Enabled sources, URLs, licenses, required env |
| `configs/priority.yaml` | Field-level source order |
| `internal/schema/record.go` | `Record` + `ToMMDBType()` |
| `internal/schema/record_test.go` | Empty-field omission, trait mapping |
| `internal/merge/priority.go` | Load `priority.yaml` |
| `internal/merge/fields.go` | `Merge(dst, src, sourceID)` |
| `internal/merge/fields_test.go` | Priority + OR tests |
| `internal/cidr/range.go` | Prefix ↔ half-open `[start,end)` |
| `internal/cidr/sweep.go` | Sweep-line field merge |
| `internal/cidr/collapse.go` | Adjacent identical collapse + CIDR emit |
| `internal/cidr/sweep_test.go` | Overlap / most-specific / collapse |
| `internal/download/client.go` | HTTP GET, retries, atomic write |
| `internal/download/client_test.go` | `httptest` retry + checksum |
| `internal/source/source.go` | `Source` interface + `Block` |
| `internal/source/ipinfo.go` | IPinfo Lite MMDB |
| `internal/source/sapics.go` | origin-asn + country MMDB |
| `internal/source/iptoasn.go` | TSV `start end asn cc org` |
| `internal/source/geolite.go` | GeoLite2-ASN (skip if no key) |
| `internal/source/openproxy.go` | OpenProxyDB CSV flags |
| `internal/source/cloud.go` | AWS/GCP/Azure/Cloudflare ranges |
| `internal/source/*_test.go` | Fixture-based adapters |
| `internal/mmdb/write.go` | Write tree, size gate, sha256 |
| `internal/mmdb/write_test.go` | Round-trip + size abort |
| `internal/attrib/generate.go` | `ATTRIBUTION.md` from enabled sources |
| `internal/attrib/generate_test.go` | Required credit strings present |
| `cmd/build/main.go` | download → merge → write → attrib |
| `cmd/lookup/main.go` | `lookup <ip> [mmdb]` JSON to stdout |
| `cmd/validate/main.go` | Golden IPs must match expected country/ASN/flags |
| `testdata/` | Tiny MMDB/CSV/TSV/JSON fixtures |
| `.github/workflows/ci.yml` | `go test ./...` on PR |
| `.github/workflows/daily.yml` | Daily 01:00 UTC build + Release |

---

### Task 1: Repo skeleton, licenses, config files

**Files:**
- Create: `go.mod`, `Makefile`, `.gitignore`, `.env.example`, `LICENSE`, `README.md`, `configs/sources.yaml`, `configs/priority.yaml`
- Test: none yet (config is parsed in Task 4)

**Interfaces:**
- Consumes: nothing
- Produces: module `github.com/shafqat-a/iplegence`; YAML contracts below

- [ ] **Step 1: Init the Go module**

```bash
cd /home/shafqat/git/iplegence
go mod init github.com/shafqat-a/iplegence
```

Edit `go.mod` so the first lines are:

```
module github.com/shafqat-a/iplegence

go 1.24.0
```

- [ ] **Step 2: Write `.gitignore`**

```
/download/
/dist/
.env
*.mmdb
!testdata/**/*.mmdb
```

- [ ] **Step 3: Write `.env.example`**

```
IPINFO_TOKEN=
MAXMIND_LICENSE_KEY=
```

- [ ] **Step 4: Write `LICENSE`**

Apache-2.0 text. First comment line in README must say: code is Apache-2.0; bundled/released data remains under each source license listed in `ATTRIBUTION.md`.

- [ ] **Step 5: Write `configs/sources.yaml`**

```yaml
download_dir: download
output_dir: dist
output_name: Superior-IP.mmdb
max_bytes: 188743680   # 180 MiB
download_timeout_seconds: 300
download_retries: 3
download_concurrency: 6

sources:
  - id: ipinfo_lite
    enabled: true
    kind: mmdb
    url: "https://ipinfo.io/data/ipinfo_lite.mmdb?token=${IPINFO_TOKEN}"
    path: download/ipinfo_lite.mmdb
    license: "CC BY-SA 4.0"
    attribution: "This product includes IP data from IPinfo (https://ipinfo.io)."
    requires_env: [IPINFO_TOKEN]
    required: true

  - id: sapics_origin_asn
    enabled: true
    kind: mmdb
    url: "https://github.com/sapics/ip-location-db/releases/download/latest/origin-asn.mmdb"
    path: download/origin-asn.mmdb
    license: "PDDL"
    attribution: "ASN data from sapics/ip-location-db origin-asn (https://github.com/sapics/ip-location-db)."
    required: true

  - id: sapics_country
    enabled: true
    kind: mmdb
    url: "https://github.com/sapics/ip-location-db/releases/download/latest/geo-whois-asn-country.mmdb"
    path: download/geo-whois-asn-country.mmdb
    license: "PDDL"
    attribution: "Country data from sapics/ip-location-db geo-whois-asn-country (https://github.com/sapics/ip-location-db)."
    required: true

  - id: iptoasn
    enabled: true
    kind: tsv_gz
    url: "https://iptoasn.com/data/ip2asn-combined.tsv.gz"
    path: download/ip2asn-combined.tsv.gz
    license: "Free (iptoasn.com)"
    attribution: "ASN/country ranges from IPtoASN (https://iptoasn.com/)."
    required: true

  - id: geolite2_asn
    enabled: true
    kind: maxmind_tar_gz
    url: "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=${MAXMIND_LICENSE_KEY}&suffix=tar.gz"
    path: download/GeoLite2-ASN.tar.gz
    extract_glob: "**/GeoLite2-ASN.mmdb"
    extract_to: download/GeoLite2-ASN.mmdb
    license: "CC BY-SA 4.0 + MaxMind GeoLite2 EULA"
    attribution: "This product includes GeoLite2 Data created by MaxMind, available from https://www.maxmind.com."
    requires_env: [MAXMIND_LICENSE_KEY]
    required: false

  - id: openproxydb
    enabled: true
    kind: csv
    url: "https://github.com/NetworkCats/OpenProxyDB/releases/latest/download/proxy_blocks.csv"
    path: download/proxy_blocks.csv
    license: "CC0 1.0"
    attribution: "Privacy/hosting flags from OpenProxyDB (https://github.com/NetworkCats/OpenProxyDB)."
    required: true

  - id: aws_ranges
    enabled: true
    kind: json
    url: "https://ip-ranges.amazonaws.com/ip-ranges.json"
    path: download/aws-ip-ranges.json
    license: "Public"
    attribution: "AWS IP ranges from https://ip-ranges.amazonaws.com/ip-ranges.json."
    flags: { is_hosting_provider: true }
    required: true

  - id: gcp_ranges
    enabled: true
    kind: json
    url: "https://www.gstatic.com/ipranges/cloud.json"
    path: download/gcp-cloud.json
    license: "Public"
    attribution: "Google Cloud IP ranges from https://www.gstatic.com/ipranges/cloud.json."
    flags: { is_hosting_provider: true }
    required: true

  - id: azure_ranges
    enabled: true
    kind: azure_servicetags
    url: "https://www.microsoft.com/download/details.aspx?id=56519"
    path: download/azure-servicetags.json
    license: "Public"
    attribution: "Azure IP ranges from Microsoft Service Tags (download id 56519)."
    flags: { is_hosting_provider: true }
    required: false

  - id: cloudflare_v4
    enabled: true
    kind: lines
    url: "https://www.cloudflare.com/ips-v4"
    path: download/cloudflare-ips-v4.txt
    license: "Public"
    attribution: "Cloudflare IPv4 ranges from https://www.cloudflare.com/ips-v4."
    flags: { is_cdn: true, is_hosting_provider: true }
    required: true

  - id: cloudflare_v6
    enabled: true
    kind: lines
    url: "https://www.cloudflare.com/ips-v6"
    path: download/cloudflare-ips-v6.txt
    license: "Public"
    attribution: "Cloudflare IPv6 ranges from https://www.cloudflare.com/ips-v6."
    flags: { is_cdn: true, is_hosting_provider: true }
    required: true
```

If `ip2asn-combined.tsv.gz` 404s, switch `iptoasn` to two sources: `https://iptoasn.com/data/ip2asn-v4.tsv.gz` and `https://iptoasn.com/data/ip2asn-v6.tsv.gz`.

- [ ] **Step 6: Write `configs/priority.yaml`**

```yaml
country: [sapics_country, ipinfo_lite, iptoasn]
continent: [ipinfo_lite, sapics_country]
asn_number: [ipinfo_lite, sapics_origin_asn, iptoasn, geolite2_asn]
asn_org: [ipinfo_lite, iptoasn, geolite2_asn]
asn_domain: [ipinfo_lite]
city: [geolite2_city, dbip_city]          # unused in Phase 1; reserved
location: [geolite2_city, dbip_city]
subdivisions: [geolite2_city]
postal: [geolite2_city]
traits: or
```

- [ ] **Step 7: Write `Makefile`**

```makefile
.PHONY: test build lookup validate
test:
	go test ./...
build:
	go run ./cmd/build
lookup:
	go run ./cmd/lookup $(IP)
validate:
	go run ./cmd/validate dist/Superior-IP.mmdb
```

- [ ] **Step 8: Write a short `README.md`**

Must include: what the project is, how to set `.env`, `make test` / `make build` / `make lookup IP=8.8.8.8`, how to read the MMDB with any MaxMind reader, a **License & Attribution** section that names IPinfo and MaxMind, and the sentence: “This is a merged free/open dataset; do not treat it as commercial-grade accuracy.”

- [ ] **Step 9: Commit**

```bash
git add go.mod Makefile .gitignore .env.example LICENSE README.md configs/sources.yaml configs/priority.yaml
git commit -m "chore: scaffold module, licenses, and source/priority config"
```

---

### Task 2: Schema record + MMDB encoding

**Files:**
- Create: `internal/schema/record.go`, `internal/schema/record_test.go`
- Test: `internal/schema/record_test.go`

**Interfaces:**
- Consumes: nothing
- Produces:

```go
package schema

type Names struct {
    En string
}

type Country struct {
    ISOCode   string
    GeonameID uint32
    Names     Names
}

type Continent struct {
    Code      string
    GeonameID uint32
    Names     Names
}

type City struct {
    GeonameID uint32
    Names     Names
}

type Location struct {
    Latitude       float64
    Longitude      float64
    AccuracyRadius uint16
    TimeZone       string
    HasCoordinates bool
}

type Subdivision struct {
    GeonameID uint32
    ISOCode   string
    Names     Names
}

type Postal struct {
    Code string
}

type ASN struct {
    Number       uint32
    Organization string
    Domain       string
}

type Traits struct {
    IsAnonymous        bool
    IsAnonymousVPN     bool
    IsHostingProvider  bool
    IsPublicProxy      bool
    IsTorExitNode      bool
    IsCDN              bool
    IsRelay            bool
}

type Record struct {
    Country      Country
    Continent    Continent
    City         City
    Location     Location
    Subdivisions []Subdivision
    Postal       Postal
    ASN          ASN
    Traits       Traits
}

func (r Record) IsEmpty() bool
func (r Record) Equal(other Record) bool
func (r Record) ToMMDBType() mmdbtype.Map // omit empty / false fields
```

MMDB keys must be exactly: `country`, `continent`, `city`, `location`, `subdivisions`, `postal`, `asn`, `traits`, `iso_code`, `names`, `en`, `code`, `geoname_id`, `latitude`, `longitude`, `accuracy_radius`, `time_zone`, `autonomous_system_number`, `autonomous_system_organization`, `as_domain`, `is_anonymous`, `is_anonymous_vpn`, `is_hosting_provider`, `is_public_proxy`, `is_tor_exit_node`, `is_cdn`, `is_relay`.

- [ ] **Step 1: Write the failing test**

```go
package schema_test

import (
    "testing"

    "github.com/maxmind/mmdbwriter/mmdbtype"
    "github.com/shafqat-a/iplegence/internal/schema"
)

func TestToMMDBTypeOmitsEmptyAndFalse(t *testing.T) {
    rec := schema.Record{
        Country: schema.Country{ISOCode: "US", Names: schema.Names{En: "United States"}},
        ASN:     schema.ASN{Number: 15169, Organization: "Google LLC", Domain: "google.com"},
        Traits:  schema.Traits{IsHostingProvider: true},
    }
    m := rec.ToMMDBType()
    if _, ok := m[mmdbtype.String("city")]; ok {
        t.Fatal("empty city must be omitted")
    }
    traits := m[mmdbtype.String("traits")].(mmdbtype.Map)
    if _, ok := traits[mmdbtype.String("is_relay")]; ok {
        t.Fatal("false traits must be omitted")
    }
    if traits[mmdbtype.String("is_hosting_provider")] != mmdbtype.Bool(true) {
        t.Fatal("true hosting flag missing")
    }
    country := m[mmdbtype.String("country")].(mmdbtype.Map)
    names := country[mmdbtype.String("names")].(mmdbtype.Map)
    if names[mmdbtype.String("en")] != mmdbtype.String("United States") {
        t.Fatal("english country name missing")
    }
}

func TestEqualAndEmpty(t *testing.T) {
    var z schema.Record
    if !z.IsEmpty() {
        t.Fatal("zero record should be empty")
    }
    a := schema.Record{Country: schema.Country{ISOCode: "US"}}
    b := a
    if !a.Equal(b) {
        t.Fatal("equal records")
    }
    b.Country.ISOCode = "DE"
    if a.Equal(b) {
        t.Fatal("different ISO should not be equal")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/schema -v
```

Expected: FAIL, package or `ToMMDBType` undefined.

- [ ] **Step 3: Implement `internal/schema/record.go`**

Implement the types above. `ToMMDBType` allocates maps only for present fields (same pattern as NetworkCats `record.go`, but `traits` instead of `proxy`, and only `names.en`). `Equal` compares all exported fields including trait bools and `Location.HasCoordinates`.

- [ ] **Step 4: Run tests**

```bash
go get github.com/maxmind/mmdbwriter@v1.2.0
go test ./internal/schema -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/schema
git commit -m "feat: add MMDB record schema with traits encoding"
```

---

### Task 3: Field-level merge + priority config

**Files:**
- Create: `internal/merge/priority.go`, `internal/merge/fields.go`, `internal/merge/fields_test.go`, `internal/merge/priority_test.go`
- Test: those `*_test.go` files

**Interfaces:**
- Consumes: `schema.Record`; YAML at `configs/priority.yaml`
- Produces:

```go
package merge

type Priority struct {
    Country     []string
    Continent   []string
    ASNNumber   []string
    ASNOrg      []string
    ASNDomain   []string
    City        []string
    Location    []string
    Subdivisions []string
    Postal      []string
    Traits      string // must be "or"
}

func LoadPriority(path string) (Priority, error)

// Merge copies non-empty fields from src into dst when srcID ranks
// higher than the current winner for that field. winners is updated in place.
// Trait bools are OR'd regardless of rank.
func Merge(dst *schema.Record, src schema.Record, srcID string, p Priority, winners map[string]string)
```

Winner map keys: `"country"`, `"continent"`, `"asn_number"`, `"asn_org"`, `"asn_domain"`, `"city"`, `"location"`, `"subdivisions"`, `"postal"`.

A source ranks higher when its index in the field's list is **smaller**. A source not in the list never wins that field (except traits).

- [ ] **Step 1: Write the failing tests**

```go
package merge_test

import (
    "path/filepath"
    "runtime"
    "testing"

    "github.com/shafqat-a/iplegence/internal/merge"
    "github.com/shafqat-a/iplegence/internal/schema"
)

func testdataPriority(t *testing.T) merge.Priority {
    t.Helper()
    _, file, _, _ := runtime.Caller(0)
    root := filepath.Join(filepath.Dir(file), "..", "..")
    p, err := merge.LoadPriority(filepath.Join(root, "configs", "priority.yaml"))
    if err != nil {
        t.Fatal(err)
    }
    return p
}

func TestASNPrefersIPinfoOverSapics(t *testing.T) {
    p := testdataPriority(t)
    var dst schema.Record
    w := map[string]string{}
    merge.Merge(&dst, schema.Record{ASN: schema.ASN{Number: 1, Organization: "sapics"}}, "sapics_origin_asn", p, w)
    merge.Merge(&dst, schema.Record{ASN: schema.ASN{Number: 15169, Organization: "Google LLC", Domain: "google.com"}}, "ipinfo_lite", p, w)
    if dst.ASN.Number != 15169 || dst.ASN.Domain != "google.com" {
        t.Fatalf("got %#v", dst.ASN)
    }
}

func TestCountryPrefersSapicsOverIPinfo(t *testing.T) {
    p := testdataPriority(t)
    var dst schema.Record
    w := map[string]string{}
    merge.Merge(&dst, schema.Record{Country: schema.Country{ISOCode: "US"}}, "ipinfo_lite", p, w)
    merge.Merge(&dst, schema.Record{Country: schema.Country{ISOCode: "DE"}}, "sapics_country", p, w)
    if dst.Country.ISOCode != "DE" {
        t.Fatalf("got %s", dst.Country.ISOCode)
    }
}

func TestTraitsOR(t *testing.T) {
    p := testdataPriority(t)
    var dst schema.Record
    w := map[string]string{}
    merge.Merge(&dst, schema.Record{Traits: schema.Traits{IsCDN: true}}, "cloudflare_v4", p, w)
    merge.Merge(&dst, schema.Record{Traits: schema.Traits{IsHostingProvider: true}}, "aws_ranges", p, w)
    if !dst.Traits.IsCDN || !dst.Traits.IsHostingProvider {
        t.Fatalf("OR failed: %#v", dst.Traits)
    }
}

func TestUnknownSourceCannotOverwrite(t *testing.T) {
    p := testdataPriority(t)
    var dst schema.Record
    w := map[string]string{}
    merge.Merge(&dst, schema.Record{Country: schema.Country{ISOCode: "US"}}, "ipinfo_lite", p, w)
    merge.Merge(&dst, schema.Record{Country: schema.Country{ISOCode: "XX"}}, "random", p, w)
    if dst.Country.ISOCode != "US" {
        t.Fatal("unknown source must not win")
    }
}
```

- [ ] **Step 2: Run tests — expect FAIL** (`merge` package missing)

```bash
go test ./internal/merge -v
```

- [ ] **Step 3: Implement**

`LoadPriority` unmarshals YAML tags `country`, `continent`, `asn_number`, `asn_org`, `asn_domain`, `city`, `location`, `subdivisions`, `postal`, `traits`.

`rank(list []string, id string) (int, bool)` returns index and true if present.

`Merge`:
- country: if `src.Country.ISOCode != ""` and rank(srcID) is better than `winners["country"]`, copy Country
- same pattern for continent (Code), ASN number/org/domain independently, city, location (`HasCoordinates` or TimeZone), subdivisions, postal
- traits: `dst.Traits.X = dst.Traits.X || src.Traits.X` for each bool

- [ ] **Step 4: Run tests — expect PASS**

```bash
go get gopkg.in/yaml.v3
go test ./internal/merge -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/merge go.mod go.sum
git commit -m "feat: field-level merge with configurable source priority"
```

---

### Task 4: CIDR sweep, most-specific rows, collapse

**Files:**
- Create: `internal/cidr/range.go`, `internal/cidr/sweep.go`, `internal/cidr/collapse.go`, `internal/cidr/sweep_test.go`
- Test: `internal/cidr/sweep_test.go`

**Interfaces:**
- Consumes: `schema.Record`, `merge.Priority`, `merge.Merge`
- Produces:

```go
package cidr

import "net/netip"

type Block struct {
    Prefix netip.Prefix
    Rec    schema.Record
    Source string
}

type Row struct {
    Prefix netip.Prefix
    Rec    schema.Record
}

func PrefixToHalfOpen(p netip.Prefix) (start, endExclusive netip.Addr, ok bool)
func MergeBlocks(blocks []Block, p merge.Priority) ([]Row, error)
```

Algorithm (IPv4 and IPv6 separately):

1. Convert each `Block.Prefix` to a half-open `[start, end)` using `netipx.PrefixIPSet` or by computing the last address + 1.
2. Collect every `start` and `end` as breakpoints; sort unique.
3. Sweep. Maintain the set of active blocks whose `[start,end)` covers the current interval.
4. For each non-empty interval, `merge.Merge` all active records (order does not matter; ranks decide). Skip empty merged records.
5. Collapse adjacent intervals with `Record.Equal`.
6. Convert each remaining `[start,end)` to one or more CIDRs via `netipx.IPRange{From: start, To: lastInclusive}.Prefixes()` and emit `[]Row`.

- [ ] **Step 1: Write the failing test**

```go
package cidr_test

import (
    "net/netip"
    "testing"

    "github.com/shafqat-a/iplegence/internal/cidr"
    "github.com/shafqat-a/iplegence/internal/merge"
    "github.com/shafqat-a/iplegence/internal/schema"
)

func mustP(s string) netip.Prefix {
    p, err := netip.ParsePrefix(s)
    if err != nil {
        panic(err)
    }
    return p
}

func pri() merge.Priority {
    return merge.Priority{
        Country:   []string{"sapics_country", "ipinfo_lite"},
        ASNNumber: []string{"ipinfo_lite", "sapics_origin_asn"},
        Traits:    "or",
    }
}

func TestMostSpecificPrefixKeepsItsIdentity(t *testing.T) {
    rows, err := cidr.MergeBlocks([]cidr.Block{
        {Prefix: mustP("10.0.0.0/8"), Rec: schema.Record{Country: schema.Country{ISOCode: "US"}}, Source: "ipinfo_lite"},
        {Prefix: mustP("10.1.0.0/16"), Rec: schema.Record{Country: schema.Country{ISOCode: "DE"}}, Source: "sapics_country"},
    }, pri())
    if err != nil {
        t.Fatal(err)
    }
    got := map[string]string{}
    for _, r := range rows {
        got[r.Prefix.String()] = r.Rec.Country.ISOCode
    }
    if got["10.1.0.0/16"] != "DE" {
        t.Fatalf("more specific should be DE, rows=%v", got)
    }
    // 10.0.0.0/8 is split; leftover pieces must stay US
    for pfx, iso := range got {
        if pfx != "10.1.0.0/16" && iso != "US" {
            t.Fatalf("%s should be US, got %s", pfx, iso)
        }
    }
}

func TestFieldMergeOnSamePrefix(t *testing.T) {
    rows, err := cidr.MergeBlocks([]cidr.Block{
        {Prefix: mustP("8.8.8.0/24"), Rec: schema.Record{ASN: schema.ASN{Number: 15169}}, Source: "ipinfo_lite"},
        {Prefix: mustP("8.8.8.0/24"), Rec: schema.Record{Country: schema.Country{ISOCode: "US"}}, Source: "sapics_country"},
    }, pri())
    if err != nil {
        t.Fatal(err)
    }
    if len(rows) != 1 {
        t.Fatalf("expected 1 row, got %d", len(rows))
    }
    if rows[0].Rec.ASN.Number != 15169 || rows[0].Rec.Country.ISOCode != "US" {
        t.Fatalf("got %#v", rows[0].Rec)
    }
}

func TestCollapseAdjacentIdentical(t *testing.T) {
    rec := schema.Record{Country: schema.Country{ISOCode: "US"}}
    rows, err := cidr.MergeBlocks([]cidr.Block{
        {Prefix: mustP("172.16.0.0/24"), Rec: rec, Source: "ipinfo_lite"},
        {Prefix: mustP("172.16.1.0/24"), Rec: rec, Source: "ipinfo_lite"},
    }, pri())
    if err != nil {
        t.Fatal(err)
    }
    if len(rows) != 1 || rows[0].Prefix.String() != "172.16.0.0/23" {
        t.Fatalf("should collapse to /23, got %#v", rows)
    }
}
```

- [ ] **Step 2: Run — expect FAIL**

```bash
go test ./internal/cidr -v
```

- [ ] **Step 3: Implement** using `go4.org/netipx` (`IPRange`, `AddrFrom16` as needed). Reject IPv4-mapped IPv6 prefixes; treat families separately. `endExclusive` for the last IPv4 address `255.255.255.255` can be represented by a sentinel (do not overflow). Document that sentinel in a comment.

- [ ] **Step 4: Run — expect PASS**

```bash
go get go4.org/netipx
go test ./internal/cidr -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/cidr go.mod go.sum
git commit -m "feat: sweep-line CIDR merge with collapse"
```

---

### Task 5: Downloader

**Files:**
- Create: `internal/config/config.go`, `internal/config/config_test.go`, `internal/download/client.go`, `internal/download/client_test.go`
- Test: those tests

**Interfaces:**
- Consumes: `configs/sources.yaml`
- Produces:

```go
package config

type Flags struct {
    IsHostingProvider bool `yaml:"is_hosting_provider"`
    IsCDN             bool `yaml:"is_cdn"`
}

type SourceSpec struct {
    ID          string   `yaml:"id"`
    Enabled     bool     `yaml:"enabled"`
    Kind        string   `yaml:"kind"`
    URL         string   `yaml:"url"`
    Path        string   `yaml:"path"`
    License     string   `yaml:"license"`
    Attribution string   `yaml:"attribution"`
    RequiresEnv []string `yaml:"requires_env"`
    Required    bool     `yaml:"required"`
    ExtractGlob string   `yaml:"extract_glob"`
    ExtractTo   string   `yaml:"extract_to"`
    Flags       Flags    `yaml:"flags"`
}

type File struct {
    DownloadDir             string       `yaml:"download_dir"`
    OutputDir               string       `yaml:"output_dir"`
    OutputName              string       `yaml:"output_name"`
    MaxBytes                int64        `yaml:"max_bytes"`
    DownloadTimeoutSeconds  int          `yaml:"download_timeout_seconds"`
    DownloadRetries         int          `yaml:"download_retries"`
    DownloadConcurrency     int          `yaml:"download_concurrency"`
    Sources                 []SourceSpec `yaml:"sources"`
}

func Load(path string) (File, error)
func (s SourceSpec) ExpandURL() string // os.ExpandEnv
func (s SourceSpec) MissingEnv() []string

package download

type Result struct {
    Source SourceSpec
    Path   string
    Err    error
}

func Fetch(ctx context.Context, spec config.SourceSpec) error
func FetchAll(ctx context.Context, cfg config.File) []Result
```

`Fetch` rules:
- Expand `${ENV}` in URL.
- If any `requires_env` is empty: if `required`, return error; else skip (nil error, no file).
- GET with context timeout, 3 retries, exponential backoff starting at 1s, only on 5xx / transport errors. 4xx (except 429) is final.
- Write to `path + ".tmp"` then rename.
- After success, file size must be > 0.
- `kind: maxmind_tar_gz`: download tar.gz, extract first match of `extract_glob` to `extract_to`.
- `kind: tsv_gz`: keep the `.gz` on disk; the adapter gunzips while reading.
- `kind: azure_servicetags`: GET the details page, regex-find the first `https://download.microsoft.com/download/.../ServiceTags_Public_*.json` URL, then download that. If none found, return a skippable error when `required: false`.

- [ ] **Step 1: Write `client_test.go` using `httptest`**

Cover: 200 writes file; first 500 then 200 succeeds; 404 on required source errors; missing `IPINFO_TOKEN` on required source errors; missing `MAXMIND_LICENSE_KEY` on optional source returns nil and creates no file.

- [ ] **Step 2: Run — expect FAIL**

```bash
go test ./internal/download ./internal/config -v
```

- [ ] **Step 3: Implement `Load`, `ExpandURL`, `Fetch`, `FetchAll`** (semaphore = `DownloadConcurrency`). `FetchAll` always returns per-source results; wrap a joined error if any **required** source failed.

- [ ] **Step 4: Run — expect PASS**

```bash
go test ./internal/download ./internal/config -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/download internal/config
git commit -m "feat: config loader and retrying source downloader"
```

---

### Task 6: Source adapters (fixtures only)

**Files:**
- Create: `internal/source/source.go`, `internal/source/ipinfo.go`, `internal/source/sapics.go`, `internal/source/iptoasn.go`, `internal/source/geolite.go`, `internal/source/openproxy.go`, `internal/source/cloud.go`, matching `*_test.go`, and `testdata/` fixtures
- Test: each adapter test file

**Interfaces:**
- Consumes: downloaded files + `config.SourceSpec`
- Produces:

```go
package source

type Block = cidr.Block // Prefix, Rec, Source (Source = spec.ID)

type Adapter interface {
    ID() string
    Load(path string, spec config.SourceSpec) ([]cidr.Block, error)
}

func Lookup(kind string) (Adapter, bool)
```

Register by `kind` **and** by `id` where one kind serves many IDs (`mmdb` is not enough — IPinfo vs sapics vs GeoLite decode differently). `Lookup` should key on `spec.ID`.

Decoders:

**ipinfo_lite** — iterate MMDB networks (`maxminddb.Networks`). Accept either nested or flat IPinfo Lite keys:

```go
type ipinfoRow struct {
    Country       string `maxminddb:"country"`
    CountryCode   string `maxminddb:"country_code"`
    Continent     string `maxminddb:"continent"`
    ContinentCode string `maxminddb:"continent_code"`
    ASN           string `maxminddb:"asn"`     // "AS15169" or "15169"
    ASName        string `maxminddb:"as_name"`
    ASDomain      string `maxminddb:"as_domain"`
}
```

Map `country`/`country_code` → `ISOCode` (uppercase 2-letter). Parse ASN by trimming `AS` prefix. Fill `Names.En` for country via a static ISO-3166 English map in `internal/schema/iso.go` (include at least US, DE, GB, FR, JP, CN, IN, BR, AU, CA, NL, IE, SG, RU, KR — plus `""` skip). Continent codes: AF, AN, AS, EU, NA, OC, SA.

**sapics_origin_asn** — sapics origin-asn MMDB typically has `autonomous_system_number` or `asn`. Decode both.

**sapics_country** — fields `country_code` or `country`. ISO only.

**geolite2_asn** — standard GeoLite2-ASN: `autonomous_system_number`, `autonomous_system_organization`.

**iptoasn** — gzip TSV, columns tab-separated: `range_start range_end AS_number country_code AS_description`. Skip `AS_number == 0`. Convert `[start,end]` inclusive to prefixes with `netipx.IPRange{From, To}.Prefixes()`. Source ID `iptoasn`.

**openproxydb** — do not assume column names. First row is a header. Detect columns named (case-insensitive) `network`, `cidr`, `start`, `end`, `is_proxy`, `is_vpn`, `is_tor`, `is_hosting`, `is_cdn`, `is_anonymous`, `is_relay`, `is_school`. Map:

| CSV | Trait |
|-----|--------|
| is_proxy | IsPublicProxy |
| is_vpn | IsAnonymousVPN |
| is_tor | IsTorExitNode |
| is_hosting | IsHostingProvider |
| is_cdn | IsCDN |
| is_anonymous | IsAnonymous |
| is_relay | IsRelay |
| is_school | ignore in Phase 1 (not in schema) |

If only `start`/`end` exist, convert the range to prefixes. Boolean cells: `1`, `true`, `yes` → true.

**cloud (aws/gcp/azure/cloudflare)** — set traits from `spec.Flags`. AWS: each `prefixes[].ip_prefix` and `ipv6_prefixes[].ipv6_prefix`. GCP: `prefixes[].ipv4Prefix` / `ipv6Prefix`. Azure: `values[].properties.addressPrefixes[]`. Cloudflare: one CIDR per non-empty non-`#` line.

- [ ] **Step 1: Build tiny fixtures under `testdata/`**

Do **not** download production DBs in unit tests. Write:

- `testdata/iptoasn.tsv` — two lines, e.g. `8.8.8.0  8.8.8.255  15169  US  GOOGLE` and `1.1.1.0  1.1.1.255  13335  AU  CLOUDFLARENET`
- `testdata/openproxy.csv` — header + `1.2.3.0/24,1,0,0,1,0,0,0`
- `testdata/aws.json` — one `10.0.0.0/8` prefix
- `testdata/gcp.json` — one `11.0.0.0/8`
- `testdata/azure.json` — one `12.0.0.0/8`
- `testdata/cloudflare-v4.txt` — `13.0.0.0/8`
- `testdata/ipinfo.mmdb` and `testdata/sapics-asn.mmdb` — generate in a `TestMain` or a helper using `mmdbwriter` (keep the helper in `internal/source/mmdbutil_test.go`)

Helper to write a 1-network MMDB:

```go
func writeTestMMDB(t *testing.T, path string, pfx string, rec mmdbtype.DataType) {
    tree, err := mmdbwriter.New(mmdbwriter.Options{IPVersion: 6, RecordSize: 28})
    // tree.Insert(net.ParseCIDR(pfx)) + rec; WriteTo file
}
```

- [ ] **Step 2: Write failing adapter tests** — one test per adapter asserting prefix count, ISO/ASN/traits, and `Block.Source == spec.ID`.

- [ ] **Step 3: Implement adapters + `Lookup`**

- [ ] **Step 4: `go test ./internal/source -v` — PASS**

- [ ] **Step 5: Commit**

```bash
git add internal/source testdata
git commit -m "feat: source adapters for Phase 1 datasets"
```

---

### Task 7: MMDB writer, size gate, checksum

**Files:**
- Create: `internal/mmdb/write.go`, `internal/mmdb/write_test.go`
- Test: `internal/mmdb/write_test.go`

**Interfaces:**
- Consumes: `[]cidr.Row`, `config.File.MaxBytes`
- Produces:

```go
package mmdb

type WriteResult struct {
    Path   string
    Bytes  int64
    SHA256 string
}

var ErrTooLarge = errors.New("mmdb exceeds max_bytes")

func Write(rows []cidr.Row, dest string, maxBytes int64) (WriteResult, error)
```

Implementation:
- `mmdbwriter.New(Options{DatabaseType: "iplegence-Superior-IP", Description: map[string]string{"en": "Merged free/open IP intelligence (country, ASN, traits)"}, IPVersion: 6, RecordSize: 28, DisableIPv4Aliasing: false, IncludeReservedNetworks: true})`
- Insert each `row.Prefix` with `row.Rec.ToMMDBType()`. Skip nil maps.
- Atomic write (`dest.tmp` → `dest`) with a 4 MiB bufio writer.
- If `result.Bytes > maxBytes` (or file size > maxBytes): delete dest, return `ErrTooLarge`.
- SHA-256 hex of the final file into `dest + ".sha256"` as `<hex>  Superior-IP.mmdb\n`.

- [ ] **Step 1: Failing tests**

1. Write two rows (`8.8.8.0/24` Google hosting, `1.1.1.0/24` Cloudflare CDN), reopen with `maxminddb.Open`, lookup `8.8.8.8` and `1.1.1.1`, assert country/ASN/traits.
2. `Write(..., maxBytes=1)` returns `ErrTooLarge` and leaves no dest file.

Use `github.com/oschwald/maxminddb-golang/v2` if v1 import path fails; pin whichever compiles with mmdbwriter v1.2.0. Decode into `schema.Record` using `maxminddb` struct tags on a **read** DTO if the writer types do not tag for reading — in that case add `internal/schema/read.go`:

```go
type LookupRecord struct {
    Country   struct {
        ISOCode string            `maxminddb:"iso_code"`
        Names   map[string]string `maxminddb:"names"`
    } `maxminddb:"country"`
    ASN struct {
        Number uint32 `maxminddb:"autonomous_system_number"`
        Org    string `maxminddb:"autonomous_system_organization"`
        Domain string `maxminddb:"as_domain"`
    } `maxminddb:"asn"`
    Traits struct {
        IsHostingProvider bool `maxminddb:"is_hosting_provider"`
        IsCDN             bool `maxminddb:"is_cdn"`
    } `maxminddb:"traits"`
}
```

- [ ] **Step 2: Run — FAIL**
- [ ] **Step 3: Implement**
- [ ] **Step 4: `go test ./internal/mmdb -v` — PASS**
- [ ] **Step 5: Commit** `feat: write Superior-IP.mmdb with size gate and sha256`

---

### Task 8: Attribution generator

**Files:**
- Create: `internal/attrib/generate.go`, `internal/attrib/generate_test.go`
- Test: `internal/attrib/generate_test.go`

**Interfaces:**

```go
package attrib

func Render(cfg config.File, used []string, dest string) error
```

`used` is the set of source IDs that actually produced ≥1 block. The file must start with:

```
# Attribution

This product merges publicly available IP intelligence sources.
It does not claim commercial-grade accuracy.

## Sources used
```

Then one subsection per used source: name/id, license, the exact `attribution` string from YAML. Sources not used are listed under `## Configured but unused`. Always mention IPinfo and MaxMind **if those IDs are in `used`**.

- [ ] **Step 1: Test that a cfg with ipinfo_lite + sapics_origin_asn in `used` writes a file containing `IPinfo` and `sapics/ip-location-db` and `commercial-grade` (the disclaimer).**
- [ ] **Step 2: FAIL → implement → PASS**
- [ ] **Step 3: Commit** `feat: generate ATTRIBUTION.md from enabled sources`

---

### Task 9: `cmd/build` pipeline

**Files:**
- Create: `cmd/build/main.go`
- Test: `internal` packages already cover units; add `cmd/build/main_test.go` only if you extract `Run(cfg) error` to `internal/pipeline/pipeline.go` (preferred).

**Interfaces:**

```go
package pipeline

type Options struct {
    ConfigPath   string
    SkipDownload bool
    PriorityPath string
}

func Run(ctx context.Context, opt Options) error
```

`Run`:
1. Load config + priority.
2. Unless `SkipDownload`, `download.FetchAll`. Fail if any required source failed.
3. For each enabled source whose file exists, `adapter.Load`. Log block counts. Collect `used` IDs with len(blocks)>0.
4. `cidr.MergeBlocks`.
5. `mmdb.Write` to `filepath.Join(cfg.OutputDir, cfg.OutputName)`.
6. `attrib.Render` to `filepath.Join(cfg.OutputDir, "ATTRIBUTION.md")`.
7. Print size MB, row count, elapsed.

`cmd/build/main.go` flags: `-config configs/sources.yaml`, `-priority configs/priority.yaml`, `-skip-download`.

- [ ] **Step 1: Write `internal/pipeline/pipeline_test.go`** that points at `testdata` files by constructing a `config.File` in memory (or a temp YAML). Skip live HTTP. Assert dest MMDB exists, sha256 exists, attribution exists, lookup of fixture IPs works.
- [ ] **Step 2: FAIL → implement `internal/pipeline` + thin `cmd/build` → PASS**
- [ ] **Step 3: Commit** `feat: wire download-merge-write pipeline`

---

### Task 10: Lookup CLI

**Files:**
- Create: `cmd/lookup/main.go`, `internal/lookup/lookup.go`, `internal/lookup/lookup_test.go`

**Interfaces:**

```go
package lookup

func Open(path string) (*DB, error)
func (d *DB) Close() error
func (d *DB) Lookup(ip string) (schema.LookupRecord, bool, error)
```

CLI: `go run ./cmd/lookup 8.8.8.8 [dist/Superior-IP.mmdb]` prints indented JSON to stdout. Unknown IP → exit 2, `{"error":"not found"}`. Bad IP → exit 1.

- [ ] **Step 1: Test Open+Lookup against the MMDB written in Task 7's testdata (or rewrite a tiny one).**
- [ ] **Step 2: Implement**
- [ ] **Step 3: Commit** `feat: add lookup CLI`

---

### Task 11: Validator

**Files:**
- Create: `cmd/validate/main.go`, `internal/validate/validate.go`, `internal/validate/validate_test.go`, `testdata/golden.json`

**Interfaces:**

```go
package validate

type Expect struct {
    IP          string `json:"ip"`
    CountryISO  string `json:"country_iso"`            // empty = ignore
    ASN         uint32 `json:"asn"`                    // 0 = ignore
    Hosting     *bool  `json:"is_hosting_provider"`
    CDN         *bool  `json:"is_cdn"`
}

func Check(dbPath string, cases []Expect) []error
```

`testdata/golden.json` (used after a **real** build; unit test uses a fixture MMDB):

```json
[
  {"ip": "8.8.8.8", "country_iso": "US", "asn": 15169, "is_hosting_provider": true},
  {"ip": "1.1.1.1", "country_iso": "AU", "is_cdn": true, "is_hosting_provider": true},
  {"ip": "9.9.9.9", "country_iso": "US"}
]
```

`cmd/validate` exits 1 if any case fails. Print one `OK`/`FAIL` line per IP.

Unit test: write a fixture MMDB with 8.8.8.8 = US/15169/hosting and assert `Check` is empty; flip expected ASN and assert a non-empty error.

- [ ] **Step 1–4: test / fail / implement / pass**
- [ ] **Step 5: Commit** `feat: add golden-IP validator`

---

### Task 12: CI + daily release

**Files:**
- Create: `.github/workflows/ci.yml`, `.github/workflows/daily.yml`

**Interfaces:**
- Consumes: repo secrets `IPINFO_TOKEN` (required), `MAXMIND_LICENSE_KEY` (optional)
- Produces: GitHub Release tag `vYYYY.MM.DD` with `Superior-IP.mmdb`, `Superior-IP.mmdb.sha256`, `ATTRIBUTION.md`

- [ ] **Step 1: Write `ci.yml`**

```yaml
name: ci
on:
  pull_request:
  push:
    branches: [main]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24.x"
      - run: go test ./...
```

- [ ] **Step 2: Write `daily.yml`**

```yaml
name: daily
on:
  schedule:
    - cron: "0 1 * * *"
  workflow_dispatch:
permissions:
  contents: write
jobs:
  build:
    runs-on: ubuntu-latest
    timeout-minutes: 180
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.24.x" }
      - name: Build database
        env:
          IPINFO_TOKEN: ${{ secrets.IPINFO_TOKEN }}
          MAXMIND_LICENSE_KEY: ${{ secrets.MAXMIND_LICENSE_KEY }}
        run: go run ./cmd/build
      - name: Validate
        run: go run ./cmd/validate dist/Superior-IP.mmdb
      - name: Release
        env:
          GH_TOKEN: ${{ github.token }}
        run: |
          TAG="v$(date -u +%Y.%m.%d)"
          gh release delete "$TAG" --yes || true
          git tag -f "$TAG"
          git push -f origin "$TAG"
          gh release create "$TAG" \
            dist/Superior-IP.mmdb \
            dist/Superior-IP.mmdb.sha256 \
            dist/ATTRIBUTION.md \
            --title "$TAG" \
            --notes "Daily iplegence build. See ATTRIBUTION.md for sources and licenses. This is a merged free/open dataset; do not treat it as commercial-grade accuracy."
```

- [ ] **Step 3: Document required secrets in README** (`Settings → Secrets and variables → Actions`).
- [ ] **Step 4: Commit** `ci: add unit workflow and daily release pipeline`

Do not run a live full build in this task unless tokens are already in the environment. Unit tests are the gate.

---

## Self-review

### Spec coverage (Phase 0 + Phase 1)

| Spec item | Task |
|-----------|------|
| Repo + start fresh with mmdbwriter | 1 |
| MaxMind + IPinfo credentials documented | 1, 5, 12 |
| Licenses & attributions | 1, 8 |
| Downloaders for Phase-1 sources | 5, 6 |
| Normalize to CIDR + structured records | 4, 6 |
| Most-specific prefix as row identity | 4 |
| Field priority YAML (ASN / country / OR flags) | 3 |
| Collapse adjacent identical records | 4 |
| Single MMDB, record size 28 | 7 |
| Checksums + ATTRIBUTION.md | 7, 8 |
| Size abort > 180 MB; Phase 1 < 100 MB target | 7, README |
| Daily GitHub Action + Releases | 12 |
| Validator 8.8.8.8 / 1.1.1.1 / hosting | 11 |
| Lookup CLI + MaxMind-compatible schema | 2, 10 |
| English-only names | 2 |
| No commercial-grade claim | 1, 8, 12 |
| `traits` schema (not NetworkCats `proxy`) | 2 |
| Official downloads only | 1, 5 |
| City / VPN / QQWry / API / dashboards | **out of scope** — Phase 2–4 plans |

### Placeholder scan

No TBD/TODO left in task steps. Azure Service Tags URL discovery is specified (parse download page; optional source). iptoasn combined vs split files has a concrete fallback.

### Type consistency

`cidr.Block` is the interchange type (`Prefix`, `Rec schema.Record`, `Source string`). `source.Block` is an alias. `merge.Merge` + `Priority` are used by `cidr.MergeBlocks`. `mmdb.Write` consumes `[]cidr.Row`. `schema.LookupRecord` is the read DTO for CLI + validator.

---

## Execution notes for the implementer

1. Never download full production databases inside `go test`.
2. A first live `make build` needs `IPINFO_TOKEN`. `MAXMIND_LICENSE_KEY` is optional; GeoLite2-ASN is skipped when unset.
3. First live build may take tens of minutes and several GB RAM (full IPv4+IPv6 sweep). Run it on a machine with ≥ 16 GB RAM.
4. If MMDB > 100 MB in Phase 1, drop iptoasn rows that only duplicate IPinfo ASN+country before writing (log how many were dropped). Do not drop unique prefixes.
5. After Task 12, stop. Do not start Phase 2 unless a new plan is written.
