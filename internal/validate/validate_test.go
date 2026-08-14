package validate_test

import (
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/mmdb"
	"github.com/shafqat-a/iplegence/internal/schema"
	"github.com/shafqat-a/iplegence/internal/validate"
)

func TestCheckOKAndFail(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "Superior-IP.mmdb")
	hosting := true
	_, err := mmdb.Write([]cidr.Row{{
		Prefix: netip.MustParsePrefix("8.8.8.0/24"),
		Rec: schema.Record{
			Country: schema.Country{ISOCode: "US"},
			City:    schema.City{Names: schema.Names{En: "Mountain View"}},
			ASN:     schema.ASN{Number: 15169},
			Traits:  schema.Traits{IsHostingProvider: true},
		},
	}}, dest, 10<<20)
	if err != nil {
		t.Fatal(err)
	}
	errs := validate.Check(dest, []validate.Expect{{
		IP: "8.8.8.8", CountryISO: "US", City: "Mountain View", ASN: 15169, Hosting: &hosting,
	}})
	if len(errs) != 0 {
		t.Fatalf("expected pass, got %v", errs)
	}
	errs = validate.Check(dest, []validate.Expect{{
		IP: "8.8.8.8", CountryISO: "US", ASN: 1,
	}})
	if len(errs) == 0 {
		t.Fatal("expected ASN mismatch")
	}
}
