package source_test

import (
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
