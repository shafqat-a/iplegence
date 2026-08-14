package schema

// LookupRecord is the maxminddb decode DTO for CLI and validator lookups.
type LookupRecord struct {
	Country struct {
		ISOCode   string            `maxminddb:"iso_code" json:"iso_code,omitempty"`
		GeonameID uint32            `maxminddb:"geoname_id" json:"geoname_id,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"country" json:"country,omitempty"`
	Continent struct {
		Code      string            `maxminddb:"code" json:"code,omitempty"`
		GeonameID uint32            `maxminddb:"geoname_id" json:"geoname_id,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"continent" json:"continent,omitempty"`
	City struct {
		GeonameID uint32            `maxminddb:"geoname_id" json:"geoname_id,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"city" json:"city,omitempty"`
	Location struct {
		Latitude       float64 `maxminddb:"latitude" json:"latitude,omitempty"`
		Longitude      float64 `maxminddb:"longitude" json:"longitude,omitempty"`
		AccuracyRadius uint16  `maxminddb:"accuracy_radius" json:"accuracy_radius,omitempty"`
		TimeZone       string  `maxminddb:"time_zone" json:"time_zone,omitempty"`
	} `maxminddb:"location" json:"location,omitempty"`
	Subdivisions []struct {
		GeonameID uint32            `maxminddb:"geoname_id" json:"geoname_id,omitempty"`
		ISOCode   string            `maxminddb:"iso_code" json:"iso_code,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"subdivisions" json:"subdivisions,omitempty"`
	Postal struct {
		Code string `maxminddb:"code" json:"code,omitempty"`
	} `maxminddb:"postal" json:"postal,omitempty"`
	ASN struct {
		Number uint32 `maxminddb:"autonomous_system_number" json:"autonomous_system_number,omitempty"`
		Org    string `maxminddb:"autonomous_system_organization" json:"autonomous_system_organization,omitempty"`
		Domain string `maxminddb:"as_domain" json:"as_domain,omitempty"`
	} `maxminddb:"asn" json:"asn,omitempty"`
	Traits struct {
		IsAnonymous       bool `maxminddb:"is_anonymous" json:"is_anonymous,omitempty"`
		IsAnonymousVPN    bool `maxminddb:"is_anonymous_vpn" json:"is_anonymous_vpn,omitempty"`
		IsHostingProvider bool `maxminddb:"is_hosting_provider" json:"is_hosting_provider,omitempty"`
		IsPublicProxy     bool `maxminddb:"is_public_proxy" json:"is_public_proxy,omitempty"`
		IsTorExitNode     bool `maxminddb:"is_tor_exit_node" json:"is_tor_exit_node,omitempty"`
		IsCDN             bool `maxminddb:"is_cdn" json:"is_cdn,omitempty"`
		IsRelay           bool `maxminddb:"is_relay" json:"is_relay,omitempty"`
	} `maxminddb:"traits" json:"traits,omitempty"`
}

func (r LookupRecord) Found() bool {
	return r.Country.ISOCode != "" ||
		r.Continent.Code != "" ||
		r.City.Names != nil ||
		r.Location.TimeZone != "" ||
		r.Location.Latitude != 0 ||
		r.Location.Longitude != 0 ||
		len(r.Subdivisions) > 0 ||
		r.Postal.Code != "" ||
		r.ASN.Number != 0 ||
		r.ASN.Org != "" ||
		r.ASN.Domain != "" ||
		r.Traits.IsAnonymous ||
		r.Traits.IsAnonymousVPN ||
		r.Traits.IsHostingProvider ||
		r.Traits.IsPublicProxy ||
		r.Traits.IsTorExitNode ||
		r.Traits.IsCDN ||
		r.Traits.IsRelay
}
