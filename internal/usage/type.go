package usage

import (
	"strings"
	"unicode"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/schema"
)

const (
	TypeResidential = "residential"
	TypeMobile      = "mobile"
	TypeBusiness    = "business"
	TypeEducation   = "education"
	TypeGovernment  = "government"
	TypeHosting     = "hosting"

	SourcePeeringDB  = "peeringdb"
	SourceASNName    = "asn_name"
	SourcePrefixFlag = "prefix_flag"
)

// Infer assigns a coarse usage type from PeeringDB network types, ASN
// name/domain keywords, and prefix-level hosting/CDN flags. Empty means unknown.
// The result is inferred, not commercial usage_type.
func Infer(rec schema.Record, pdbTypes []string) (typ, src string) {
	org := rec.ASN.Organization
	domain := rec.ASN.Domain

	if t := fromPeeringDB(pdbTypes, TypeEducation, TypeGovernment); t != "" {
		return t, SourcePeeringDB
	}
	if t := fromName(org, domain, TypeEducation, TypeGovernment); t != "" {
		return t, SourceASNName
	}
	if rec.Traits.IsHostingProvider || rec.Traits.IsCDN {
		return TypeHosting, SourcePrefixFlag
	}
	if t := fromName(org, domain, TypeMobile); t != "" {
		return t, SourceASNName
	}
	if t := fromPeeringDB(pdbTypes, TypeResidential, TypeBusiness, TypeHosting); t != "" {
		return t, SourcePeeringDB
	}
	if t := fromName(org, domain, TypeHosting, TypeResidential); t != "" {
		return t, SourceASNName
	}
	return "", ""
}

func Apply(rows []cidr.Row, catalog map[uint32][]string) {
	for i := range rows {
		typ, src := Infer(rows[i].Rec, catalog[rows[i].Rec.ASN.Number])
		rows[i].Rec.Traits.UsageType = typ
		rows[i].Rec.Traits.UsageTypeSource = src
	}
}

func fromPeeringDB(types []string, want ...string) string {
	have := map[string]bool{}
	for _, raw := range types {
		if t := mapPeeringDB(raw); t != "" {
			have[t] = true
		}
	}
	for _, t := range want {
		if have[t] {
			return t
		}
	}
	return ""
}

func mapPeeringDB(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "/")
	switch s {
	case "educational/research", "education", "educational", "research":
		return TypeEducation
	case "government":
		return TypeGovernment
	case "enterprise":
		return TypeBusiness
	case "cable/dsl/isp", "cable/dsl", "isp":
		return TypeResidential
	case "cdn", "content":
		return TypeHosting
	default:
		return ""
	}
}

func fromName(org, domain string, want ...string) string {
	norm := normalize(org)
	have := map[string]bool{}
	if matchEducation(norm, domain) {
		have[TypeEducation] = true
	}
	if matchGovernment(norm, domain) {
		have[TypeGovernment] = true
	}
	if matchMobile(norm) {
		have[TypeMobile] = true
	}
	if matchHosting(norm) {
		have[TypeHosting] = true
	}
	if matchResidential(norm) {
		have[TypeResidential] = true
	}
	for _, t := range want {
		if have[t] {
			return t
		}
	}
	return ""
}

func matchEducation(norm, domain string) bool {
	if domainSuffix(domain, ".edu", ".edu.", ".ac.uk", ".ac.", ".edu.au") {
		return true
	}
	words := []string{
		"university", "universidad", "universite", "universitat", "universiteit",
		"universidade", "college", "campus", "education", "educational",
		"academic", "nren", "polytechnic",
	}
	if hasAnyWord(norm, words...) {
		return true
	}
	return hasAnyPhrase(norm, "school district", "school of", "research network", "research education")
}

func matchGovernment(norm, domain string) bool {
	if domainSuffix(domain, ".gov", ".gov.", ".govt.", ".gob.", ".gouv.") {
		return true
	}
	words := []string{
		"government", "gouvernement", "gobierno", "governo",
		"ministry", "ministere", "ministerio",
		"municipality", "municipal", "govt",
	}
	if hasAnyWord(norm, words...) {
		return true
	}
	return hasAnyPhrase(norm, "armed forces", "public sector", "federal government")
}

func matchMobile(norm string) bool {
	return hasAnyWord(norm, "mobile", "wireless", "cellular", "gsm", "umts", "wcdma", "lte", "5g")
}

func matchHosting(norm string) bool {
	if hasAnyWord(norm, "hosting", "datacenter", "datacentre", "colocation", "vps", "cloud") {
		return true
	}
	return hasAnyPhrase(norm, "data center", "data centre", "dedicated server")
}

func matchResidential(norm string) bool {
	if hasAnyWord(norm, "telecom", "telecommunications", "telecommunication", "broadband", "isp", "dsl", "telekom", "telefonica", "telkom") {
		return true
	}
	return hasAnyPhrase(norm, "internet service", "cable internet")
}

func normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := true
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSpace = false
			continue
		}
		if !prevSpace {
			b.WriteByte(' ')
			prevSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func hasAnyWord(norm string, words ...string) bool {
	for _, w := range words {
		if hasPhrase(norm, w) {
			return true
		}
	}
	return false
}

func hasAnyPhrase(norm string, phrases ...string) bool {
	for _, p := range phrases {
		if hasPhrase(norm, p) {
			return true
		}
	}
	return false
}

func hasPhrase(norm, phrase string) bool {
	if norm == "" || phrase == "" {
		return false
	}
	return strings.Contains(" "+norm+" ", " "+strings.ToLower(phrase)+" ")
}

func domainSuffix(domain string, suffixes ...string) bool {
	d := strings.ToLower(strings.TrimSpace(domain))
	if d == "" {
		return false
	}
	for _, s := range suffixes {
		if strings.HasSuffix(d, strings.TrimSuffix(s, ".")) || strings.Contains(d, s) {
			return true
		}
	}
	return false
}
