package source_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/source"
)

func testdata(t *testing.T, name string) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", name)
}

func TestIPtoASN(t *testing.T) {
	a, ok := source.Lookup("iptoasn")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(testdata(t, "iptoasn.tsv"), config.SourceSpec{ID: "iptoasn"})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("no blocks")
	}
	var saw bool
	for _, b := range blocks {
		if b.Source != "iptoasn" {
			t.Fatalf("source %q", b.Source)
		}
		if b.Rec.ASN.Number == 15169 && b.Rec.Country.ISOCode == "US" {
			saw = true
		}
	}
	if !saw {
		t.Fatalf("missing 8.8.8.0 GOOGLE, blocks=%d", len(blocks))
	}
}

func TestOpenProxy(t *testing.T) {
	a, ok := source.Lookup("openproxydb")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(testdata(t, "openproxy.csv"), config.SourceSpec{ID: "openproxydb"})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("got %d", len(blocks))
	}
	b := blocks[0]
	if b.Prefix.String() != "1.2.3.0/24" || !b.Rec.Traits.IsPublicProxy || !b.Rec.Traits.IsHostingProvider {
		t.Fatalf("got %#v", b)
	}
	if b.Rec.Traits.IsAnonymousVPN || b.Source != "openproxydb" {
		t.Fatalf("got %#v", b)
	}
}

func TestOpenProxyCurrentHeader(t *testing.T) {
	a, ok := source.Lookup("openproxydb")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(testdata(t, "openproxy-current.csv"), config.SourceSpec{ID: "openproxydb"})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("got %d", len(blocks))
	}
	b := blocks[0]
	if b.Prefix.String() != "4.5.6.0/24" || !b.Rec.Traits.IsPublicProxy || !b.Rec.Traits.IsCDN || !b.Rec.Traits.IsHostingProvider {
		t.Fatalf("got %#v", b)
	}
}

func TestCloudAdapters(t *testing.T) {
	cases := []struct {
		id, file, pfx string
		cdn           bool
	}{
		{"aws_ranges", "aws.json", "10.0.0.0/8", false},
		{"gcp_ranges", "gcp.json", "11.0.0.0/8", false},
		{"azure_ranges", "azure.json", "12.0.0.0/8", false},
		{"cloudflare_v4", "cloudflare-v4.txt", "13.0.0.0/8", true},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			a, ok := source.Lookup(tc.id)
			if !ok {
				t.Fatal("missing adapter")
			}
			spec := config.SourceSpec{ID: tc.id, Flags: config.Flags{IsHostingProvider: true, IsCDN: tc.cdn}}
			blocks, err := a.Load(testdata(t, tc.file), spec)
			if err != nil {
				t.Fatal(err)
			}
			if len(blocks) != 1 {
				t.Fatalf("got %d", len(blocks))
			}
			if blocks[0].Prefix.String() != tc.pfx || !blocks[0].Rec.Traits.IsHostingProvider || blocks[0].Source != tc.id {
				t.Fatalf("got %#v", blocks[0])
			}
			if blocks[0].Rec.Traits.IsCDN != tc.cdn {
				t.Fatalf("cdn=%v", blocks[0].Rec.Traits.IsCDN)
			}
		})
	}
}

func TestIPinfoMMDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ipinfo.mmdb")
	writeTestMMDB(t, path, "8.8.8.0/24", mmdbtype.Map{
		mmdbtype.String("country_code"): mmdbtype.String("US"),
		mmdbtype.String("continent"):    mmdbtype.String("NA"),
		mmdbtype.String("asn"):          mmdbtype.String("AS15169"),
		mmdbtype.String("as_name"):      mmdbtype.String("Google LLC"),
		mmdbtype.String("as_domain"):    mmdbtype.String("google.com"),
	})
	a, ok := source.Lookup("ipinfo_lite")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(path, config.SourceSpec{ID: "ipinfo_lite"})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("got %d", len(blocks))
	}
	b := blocks[0]
	if b.Source != "ipinfo_lite" || b.Rec.Country.ISOCode != "US" || b.Rec.ASN.Number != 15169 || b.Rec.ASN.Domain != "google.com" {
		t.Fatalf("got %#v", b)
	}
	if b.Rec.Continent.Code != "NA" {
		t.Fatalf("continent %#v", b.Rec.Continent)
	}
}

func TestSapicsASNMMDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sapics-asn.mmdb")
	writeTestMMDB(t, path, "8.8.8.0/24", mmdbtype.Map{
		mmdbtype.String("autonomous_system_number"): mmdbtype.Uint32(15169),
	})
	a, ok := source.Lookup("sapics_origin_asn")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(path, config.SourceSpec{ID: "sapics_origin_asn"})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].Rec.ASN.Number != 15169 || blocks[0].Source != "sapics_origin_asn" {
		t.Fatalf("got %#v", blocks)
	}
}

func TestSapicsCountryMMDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sapics-cc.mmdb")
	writeTestMMDB(t, path, "8.8.8.0/24", mmdbtype.Map{
		mmdbtype.String("country_code"): mmdbtype.String("de"),
	})
	a, ok := source.Lookup("sapics_country")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(path, config.SourceSpec{ID: "sapics_country"})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].Rec.Country.ISOCode != "DE" {
		t.Fatalf("got %#v", blocks)
	}
}

func TestIPapiCSV(t *testing.T) {
	a, ok := source.Lookup("ipapi_city_v4")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(testdata(t, "ipapi.csv"), config.SourceSpec{ID: "ipapi_city_v4"})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("no blocks")
	}
	b := blocks[0]
	if b.Source != "ipapi_city_v4" || b.Rec.Country.ISOCode != "US" || b.Rec.City.Names.En != "Mountain View" {
		t.Fatalf("got %#v", b)
	}
	if !b.Rec.Location.HasCoordinates || b.Rec.Location.TimeZone != "America/Los_Angeles" {
		t.Fatalf("location %#v", b.Rec.Location)
	}
}

func TestCityMMDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "city.mmdb")
	writeTestMMDB(t, path, "8.8.8.0/24", mmdbtype.Map{
		mmdbtype.String("country"): mmdbtype.Map{
			mmdbtype.String("iso_code"): mmdbtype.String("US"),
			mmdbtype.String("names"):    mmdbtype.Map{mmdbtype.String("en"): mmdbtype.String("United States")},
		},
		mmdbtype.String("city"): mmdbtype.Map{
			mmdbtype.String("names"): mmdbtype.Map{mmdbtype.String("en"): mmdbtype.String("Mountain View")},
		},
		mmdbtype.String("location"): mmdbtype.Map{
			mmdbtype.String("latitude"):  mmdbtype.Float64(37.386),
			mmdbtype.String("longitude"): mmdbtype.Float64(-122.0838),
			mmdbtype.String("time_zone"): mmdbtype.String("America/Los_Angeles"),
		},
	})
	a, ok := source.Lookup("geolite2_city")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(path, config.SourceSpec{ID: "geolite2_city"})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("got %d", len(blocks))
	}
	b := blocks[0]
	if b.Rec.City.Names.En != "Mountain View" || b.Rec.Country.ISOCode != "US" || !b.Rec.Location.HasCoordinates {
		t.Fatalf("got %#v", b.Rec)
	}
}

func TestNordVPNJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "nord.json")
	if err := os.WriteFile(p, []byte(`[{"station":"1.2.3.4","ips":[{"ip":{"ip":"5.6.7.8"}}]}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	a, ok := source.Lookup("nordvpn")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(p, config.SourceSpec{
		ID: "nordvpn", Kind: "nordvpn_json",
		Flags: config.Flags{IsAnonymousVPN: true, IsAnonymous: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("got %d %#v", len(blocks), blocks)
	}
}

func TestRipeAnnounced(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ripe.json")
	if err := os.WriteFile(p, []byte(`{"data":{"prefixes":[{"prefix":"15.0.0.0/8"}]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	a, ok := source.Lookup("ovh_ranges")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(p, config.SourceSpec{
		ID: "ovh_ranges", Kind: "ripe_announced",
		Flags: config.Flags{IsHostingProvider: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].Prefix.String() != "15.0.0.0/8" || !blocks[0].Rec.Traits.IsHostingProvider {
		t.Fatalf("got %#v", blocks)
	}
}

func TestOverlayTorAndRelay(t *testing.T) {
	dir := t.TempDir()
	torPath := filepath.Join(dir, "tor.txt")
	if err := os.WriteFile(torPath, []byte("185.220.101.1\n# comment\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	a, ok := source.Lookup("tor_exits")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(torPath, config.SourceSpec{
		ID: "tor_exits", Kind: "lines",
		Flags: config.Flags{IsTorExitNode: true, IsAnonymous: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || !blocks[0].Rec.Traits.IsTorExitNode || blocks[0].Prefix.String() != "185.220.101.1/32" {
		t.Fatalf("got %#v", blocks)
	}

	csvPath := filepath.Join(dir, "relay.csv")
	if err := os.WriteFile(csvPath, []byte("egress_ip,country\n2.2.2.0/24,US\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	a, ok = source.Lookup("icloud_relay")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err = a.Load(csvPath, config.SourceSpec{
		ID: "icloud_relay", Kind: "csv",
		Flags: config.Flags{IsRelay: true, IsAnonymous: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || !blocks[0].Rec.Traits.IsRelay || blocks[0].Prefix.String() != "2.2.2.0/24" {
		t.Fatalf("got %#v", blocks)
	}
}

func TestGeoLiteASN(t *testing.T) {
	path := filepath.Join(t.TempDir(), "geolite.mmdb")
	writeTestMMDB(t, path, "1.1.1.0/24", mmdbtype.Map{
		mmdbtype.String("autonomous_system_number"):       mmdbtype.Uint32(13335),
		mmdbtype.String("autonomous_system_organization"): mmdbtype.String("CLOUDFLARENET"),
	})
	a, ok := source.Lookup("geolite2_asn")
	if !ok {
		t.Fatal("missing adapter")
	}
	blocks, err := a.Load(path, config.SourceSpec{ID: "geolite2_asn"})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].Rec.ASN.Number != 13335 || blocks[0].Rec.ASN.Organization != "CLOUDFLARENET" {
		t.Fatalf("got %#v", blocks)
	}
}
