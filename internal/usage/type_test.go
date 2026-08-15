package usage_test

import (
	"net/netip"
	"testing"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/schema"
	"github.com/shafqat-a/iplegence/internal/usage"
)

func TestInferPeeringDBEducationBeatsHostingFlag(t *testing.T) {
	rec := schema.Record{
		ASN:    schema.ASN{Number: 18, Organization: "University of Texas"},
		Traits: schema.Traits{IsHostingProvider: true},
	}
	typ, src := usage.Infer(rec, []string{"Educational/Research"})
	if typ != usage.TypeEducation || src != usage.SourcePeeringDB {
		t.Fatalf("got %s/%s", typ, src)
	}
}

func TestInferPrefixHostingBeatsISPName(t *testing.T) {
	rec := schema.Record{
		ASN:    schema.ASN{Number: 16509, Organization: "Amazon.com, Inc."},
		Traits: schema.Traits{IsHostingProvider: true},
	}
	typ, src := usage.Infer(rec, []string{"Cable/DSL/ISP"})
	if typ != usage.TypeHosting || src != usage.SourcePrefixFlag {
		t.Fatalf("got %s/%s", typ, src)
	}
}

func TestInferNameEducationAndGovernment(t *testing.T) {
	typ, src := usage.Infer(schema.Record{ASN: schema.ASN{Organization: "Example University Network"}}, nil)
	if typ != usage.TypeEducation || src != usage.SourceASNName {
		t.Fatalf("university: %s/%s", typ, src)
	}
	typ, src = usage.Infer(schema.Record{ASN: schema.ASN{Organization: "Ministry of Health", Domain: "moh.gov.bd"}}, nil)
	if typ != usage.TypeGovernment || src != usage.SourceASNName {
		t.Fatalf("gov: %s/%s", typ, src)
	}
}

func TestInferMobileAndResidentialNames(t *testing.T) {
	typ, src := usage.Infer(schema.Record{ASN: schema.ASN{Organization: "Grameenphone Mobile"}}, nil)
	if typ != usage.TypeMobile || src != usage.SourceASNName {
		t.Fatalf("mobile: %s/%s", typ, src)
	}
	typ, src = usage.Infer(schema.Record{ASN: schema.ASN{Organization: "Example Broadband Telecom"}}, nil)
	if typ != usage.TypeResidential || src != usage.SourceASNName {
		t.Fatalf("isp: %s/%s", typ, src)
	}
}

func TestInferPeeringDBEnterpriseAndContent(t *testing.T) {
	typ, src := usage.Infer(schema.Record{ASN: schema.ASN{Organization: "Some Corp"}}, []string{"Enterprise"})
	if typ != usage.TypeBusiness || src != usage.SourcePeeringDB {
		t.Fatalf("enterprise: %s/%s", typ, src)
	}
	typ, src = usage.Infer(schema.Record{ASN: schema.ASN{Organization: "Google LLC"}}, []string{"Content"})
	if typ != usage.TypeHosting || src != usage.SourcePeeringDB {
		t.Fatalf("content: %s/%s", typ, src)
	}
}

func TestInferDoesNotGuessUnknownOrCloudflareToken(t *testing.T) {
	typ, src := usage.Infer(schema.Record{ASN: schema.ASN{Organization: "GOOGLE", Domain: "google.com"}}, nil)
	if typ != "" || src != "" {
		t.Fatalf("google name must stay unknown, got %s/%s", typ, src)
	}
	typ, src = usage.Infer(schema.Record{ASN: schema.ASN{Organization: "CLOUDFLARENET"}}, nil)
	if typ != "" {
		t.Fatalf("cloudflare token must not match cloud word, got %s", typ)
	}
}

func TestInferIgnoresNSPAndNotDisclosed(t *testing.T) {
	typ, _ := usage.Infer(schema.Record{ASN: schema.ASN{Organization: "Transit Co"}}, []string{"NSP", "Not Disclosed"})
	if typ != "" {
		t.Fatalf("got %s", typ)
	}
}

func TestApplyWritesTraits(t *testing.T) {
	rows := []cidr.Row{{
		Prefix: netip.MustParsePrefix("129.114.0.0/16"),
		Rec:    schema.Record{ASN: schema.ASN{Number: 18, Organization: "UT"}},
	}}
	usage.Apply(rows, usage.Catalog{18: {"Educational/Research"}})
	if rows[0].Rec.Traits.UsageType != usage.TypeEducation {
		t.Fatalf("got %#v", rows[0].Rec.Traits)
	}
	if rows[0].Rec.Traits.UsageTypeSource != usage.SourcePeeringDB {
		t.Fatalf("got %#v", rows[0].Rec.Traits)
	}
}
