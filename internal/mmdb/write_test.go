package mmdb_test

import (
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"testing"

	"github.com/oschwald/maxminddb-golang/v2"
	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/mmdb"
	"github.com/shafqat-a/iplegence/internal/schema"
)

func mustP(s string) netip.Prefix {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		panic(err)
	}
	return p
}

func TestWriteRoundTrip(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "Superior-IP.mmdb")
	rows := []cidr.Row{
		{
			Prefix: mustP("8.8.8.0/24"),
			Rec: schema.Record{
				Country: schema.Country{ISOCode: "US", Names: schema.Names{En: "United States"}},
				ASN:     schema.ASN{Number: 15169, Organization: "Google LLC", Domain: "google.com"},
				Traits:  schema.Traits{IsHostingProvider: true},
			},
		},
		{
			Prefix: mustP("1.1.1.0/24"),
			Rec: schema.Record{
				Country: schema.Country{ISOCode: "AU", Names: schema.Names{En: "Australia"}},
				ASN:     schema.ASN{Number: 13335, Organization: "CLOUDFLARENET"},
				Traits:  schema.Traits{IsCDN: true, IsHostingProvider: true},
			},
		},
	}
	res, err := mmdb.Write(rows, dest, 10<<20)
	if err != nil {
		t.Fatal(err)
	}
	if res.Bytes == 0 || res.SHA256 == "" {
		t.Fatalf("result %#v", res)
	}
	if _, err := os.Stat(dest + ".sha256"); err != nil {
		t.Fatal(err)
	}
	db, err := maxminddb.Open(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var rec8 schema.LookupRecord
	ip8 := netip.MustParseAddr("8.8.8.8")
	if err := db.Lookup(ip8).Decode(&rec8); err != nil {
		t.Fatal(err)
	}
	if rec8.Country.ISOCode != "US" || rec8.ASN.Number != 15169 || !rec8.Traits.IsHostingProvider {
		t.Fatalf("8.8.8.8 got %+v", rec8)
	}

	var rec1 schema.LookupRecord
	ip1 := netip.MustParseAddr("1.1.1.1")
	if err := db.Lookup(ip1).Decode(&rec1); err != nil {
		t.Fatal(err)
	}
	if rec1.Country.ISOCode != "AU" || !rec1.Traits.IsCDN || !rec1.Traits.IsHostingProvider {
		t.Fatalf("1.1.1.1 got %+v", rec1)
	}
}

func TestWriteTooLarge(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "Superior-IP.mmdb")
	rows := []cidr.Row{{
		Prefix: mustP("8.8.8.0/24"),
		Rec:    schema.Record{Country: schema.Country{ISOCode: "US"}},
	}}
	_, err := mmdb.Write(rows, dest, 1)
	if !errors.Is(err, mmdb.ErrTooLarge) {
		t.Fatalf("want ErrTooLarge, got %v", err)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("dest must be removed on size abort")
	}
}
