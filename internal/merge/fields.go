package merge

import "github.com/shafqat-a/iplegence/internal/schema"

// Merge copies non-empty fields from src into dst when srcID ranks
// higher than the current winner for that field. winners is updated in place.
// Trait bools are OR'd regardless of rank.
func Merge(dst *schema.Record, src schema.Record, srcID string, p Priority, winners map[string]string) {
	if dst == nil {
		return
	}
	if src.Country.ISOCode != "" && better(p.Country, srcID, winners["country"]) {
		dst.Country = src.Country
		winners["country"] = srcID
	}
	if src.Continent.Code != "" && better(p.Continent, srcID, winners["continent"]) {
		dst.Continent = src.Continent
		winners["continent"] = srcID
	}
	if src.ASN.Number != 0 && better(p.ASNNumber, srcID, winners["asn_number"]) {
		dst.ASN.Number = src.ASN.Number
		winners["asn_number"] = srcID
	}
	if src.ASN.Organization != "" && better(p.ASNOrg, srcID, winners["asn_org"]) {
		dst.ASN.Organization = src.ASN.Organization
		winners["asn_org"] = srcID
	}
	if src.ASN.Domain != "" && better(p.ASNDomain, srcID, winners["asn_domain"]) {
		dst.ASN.Domain = src.ASN.Domain
		winners["asn_domain"] = srcID
	}
	if (src.City.Names.En != "" || src.City.GeonameID != 0) && better(p.City, srcID, winners["city"]) {
		dst.City = src.City
		winners["city"] = srcID
	}
	if (src.Location.HasCoordinates || src.Location.TimeZone != "") && better(p.Location, srcID, winners["location"]) {
		dst.Location = src.Location
		winners["location"] = srcID
	}
	if len(src.Subdivisions) > 0 && better(p.Subdivisions, srcID, winners["subdivisions"]) {
		dst.Subdivisions = append([]schema.Subdivision(nil), src.Subdivisions...)
		winners["subdivisions"] = srcID
	}
	if src.Postal.Code != "" && better(p.Postal, srcID, winners["postal"]) {
		dst.Postal = src.Postal
		winners["postal"] = srcID
	}
	dst.Traits.IsAnonymous = dst.Traits.IsAnonymous || src.Traits.IsAnonymous
	dst.Traits.IsAnonymousVPN = dst.Traits.IsAnonymousVPN || src.Traits.IsAnonymousVPN
	dst.Traits.IsHostingProvider = dst.Traits.IsHostingProvider || src.Traits.IsHostingProvider
	dst.Traits.IsPublicProxy = dst.Traits.IsPublicProxy || src.Traits.IsPublicProxy
	dst.Traits.IsTorExitNode = dst.Traits.IsTorExitNode || src.Traits.IsTorExitNode
	dst.Traits.IsCDN = dst.Traits.IsCDN || src.Traits.IsCDN
	dst.Traits.IsRelay = dst.Traits.IsRelay || src.Traits.IsRelay
}
