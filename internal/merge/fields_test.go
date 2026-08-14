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

func TestCityPrefersGeoLiteOverDBIP(t *testing.T) {
	p := testdataPriority(t)
	var dst schema.Record
	w := map[string]string{}
	merge.Merge(&dst, schema.Record{City: schema.City{Names: schema.Names{En: "Other"}}}, "dbip_city", p, w)
	merge.Merge(&dst, schema.Record{City: schema.City{Names: schema.Names{En: "Mountain View"}}}, "geolite2_city", p, w)
	if dst.City.Names.En != "Mountain View" {
		t.Fatalf("got %s", dst.City.Names.En)
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
