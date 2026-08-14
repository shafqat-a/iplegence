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
