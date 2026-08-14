package lookup_test

import (
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/lookup"
	"github.com/shafqat-a/iplegence/internal/mmdb"
	"github.com/shafqat-a/iplegence/internal/schema"
)

func TestOpenLookup(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "Superior-IP.mmdb")
	_, err := mmdb.Write([]cidr.Row{{
		Prefix: netip.MustParsePrefix("8.8.8.0/24"),
		Rec: schema.Record{
			Country: schema.Country{ISOCode: "US"},
			ASN:     schema.ASN{Number: 15169},
			Traits:  schema.Traits{IsHostingProvider: true},
		},
	}}, dest, 10<<20)
	if err != nil {
		t.Fatal(err)
	}
	db, err := lookup.Open(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rec, ok, err := db.Lookup("8.8.8.8")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if rec.Country.ISOCode != "US" || rec.ASN.Number != 15169 {
		t.Fatalf("got %+v", rec)
	}
	_, ok, err = db.Lookup("9.9.9.9")
	if err != nil || ok {
		t.Fatalf("unknown should be not found, ok=%v err=%v", ok, err)
	}
	_, _, err = db.Lookup("not-an-ip")
	if err == nil {
		t.Fatal("bad ip should error")
	}
}
