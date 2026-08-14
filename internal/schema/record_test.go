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
