package schema

import "github.com/maxmind/mmdbwriter/mmdbtype"

type Names struct {
	En string
}

type Country struct {
	ISOCode   string
	GeonameID uint32
	Names     Names
}

type Continent struct {
	Code      string
	GeonameID uint32
	Names     Names
}

type City struct {
	GeonameID uint32
	Names     Names
}

type Location struct {
	Latitude       float64
	Longitude      float64
	AccuracyRadius uint16
	TimeZone       string
	HasCoordinates bool
}

type Subdivision struct {
	GeonameID uint32
	ISOCode   string
	Names     Names
}

type Postal struct {
	Code string
}

type ASN struct {
	Number       uint32
	Organization string
	Domain       string
}

type Traits struct {
	IsAnonymous       bool
	IsAnonymousVPN    bool
	IsHostingProvider bool
	IsPublicProxy     bool
	IsTorExitNode     bool
	IsCDN             bool
	IsRelay           bool
}

type Record struct {
	Country      Country
	Continent    Continent
	City         City
	Location     Location
	Subdivisions []Subdivision
	Postal       Postal
	ASN          ASN
	Traits       Traits
}

func (r Record) IsEmpty() bool {
	return r.Country == (Country{}) &&
		r.Continent == (Continent{}) &&
		r.City == (City{}) &&
		r.Location == (Location{}) &&
		len(r.Subdivisions) == 0 &&
		r.Postal == (Postal{}) &&
		r.ASN == (ASN{}) &&
		r.Traits == (Traits{})
}

func (r Record) Equal(other Record) bool {
	if r.Country != other.Country ||
		r.Continent != other.Continent ||
		r.City != other.City ||
		r.Location != other.Location ||
		r.Postal != other.Postal ||
		r.ASN != other.ASN ||
		r.Traits != other.Traits {
		return false
	}
	if len(r.Subdivisions) != len(other.Subdivisions) {
		return false
	}
	for i := range r.Subdivisions {
		if r.Subdivisions[i] != other.Subdivisions[i] {
			return false
		}
	}
	return true
}

func (r Record) ToMMDBType() mmdbtype.Map {
	m := mmdbtype.Map{}
	if c := countryMap(r.Country); c != nil {
		m[mmdbtype.String("country")] = c
	}
	if c := continentMap(r.Continent); c != nil {
		m[mmdbtype.String("continent")] = c
	}
	if c := cityMap(r.City); c != nil {
		m[mmdbtype.String("city")] = c
	}
	if loc := locationMap(r.Location); loc != nil {
		m[mmdbtype.String("location")] = loc
	}
	if subs := subdivisionsMap(r.Subdivisions); subs != nil {
		m[mmdbtype.String("subdivisions")] = subs
	}
	if r.Postal.Code != "" {
		m[mmdbtype.String("postal")] = mmdbtype.Map{
			mmdbtype.String("code"): mmdbtype.String(r.Postal.Code),
		}
	}
	if a := asnMap(r.ASN); a != nil {
		m[mmdbtype.String("asn")] = a
	}
	if t := traitsMap(r.Traits); t != nil {
		m[mmdbtype.String("traits")] = t
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func namesMap(n Names) mmdbtype.Map {
	if n.En == "" {
		return nil
	}
	return mmdbtype.Map{mmdbtype.String("en"): mmdbtype.String(n.En)}
}

func countryMap(c Country) mmdbtype.Map {
	m := mmdbtype.Map{}
	if c.ISOCode != "" {
		m[mmdbtype.String("iso_code")] = mmdbtype.String(c.ISOCode)
	}
	if c.GeonameID != 0 {
		m[mmdbtype.String("geoname_id")] = mmdbtype.Uint32(c.GeonameID)
	}
	if n := namesMap(c.Names); n != nil {
		m[mmdbtype.String("names")] = n
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func continentMap(c Continent) mmdbtype.Map {
	m := mmdbtype.Map{}
	if c.Code != "" {
		m[mmdbtype.String("code")] = mmdbtype.String(c.Code)
	}
	if c.GeonameID != 0 {
		m[mmdbtype.String("geoname_id")] = mmdbtype.Uint32(c.GeonameID)
	}
	if n := namesMap(c.Names); n != nil {
		m[mmdbtype.String("names")] = n
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func cityMap(c City) mmdbtype.Map {
	m := mmdbtype.Map{}
	if c.GeonameID != 0 {
		m[mmdbtype.String("geoname_id")] = mmdbtype.Uint32(c.GeonameID)
	}
	if n := namesMap(c.Names); n != nil {
		m[mmdbtype.String("names")] = n
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func locationMap(l Location) mmdbtype.Map {
	m := mmdbtype.Map{}
	if l.HasCoordinates {
		m[mmdbtype.String("latitude")] = mmdbtype.Float64(l.Latitude)
		m[mmdbtype.String("longitude")] = mmdbtype.Float64(l.Longitude)
	}
	if l.AccuracyRadius != 0 {
		m[mmdbtype.String("accuracy_radius")] = mmdbtype.Uint16(l.AccuracyRadius)
	}
	if l.TimeZone != "" {
		m[mmdbtype.String("time_zone")] = mmdbtype.String(l.TimeZone)
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func subdivisionsMap(subs []Subdivision) mmdbtype.Slice {
	if len(subs) == 0 {
		return nil
	}
	out := make(mmdbtype.Slice, 0, len(subs))
	for _, s := range subs {
		m := mmdbtype.Map{}
		if s.GeonameID != 0 {
			m[mmdbtype.String("geoname_id")] = mmdbtype.Uint32(s.GeonameID)
		}
		if s.ISOCode != "" {
			m[mmdbtype.String("iso_code")] = mmdbtype.String(s.ISOCode)
		}
		if n := namesMap(s.Names); n != nil {
			m[mmdbtype.String("names")] = n
		}
		if len(m) == 0 {
			continue
		}
		out = append(out, m)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func asnMap(a ASN) mmdbtype.Map {
	m := mmdbtype.Map{}
	if a.Number != 0 {
		m[mmdbtype.String("autonomous_system_number")] = mmdbtype.Uint32(a.Number)
	}
	if a.Organization != "" {
		m[mmdbtype.String("autonomous_system_organization")] = mmdbtype.String(a.Organization)
	}
	if a.Domain != "" {
		m[mmdbtype.String("as_domain")] = mmdbtype.String(a.Domain)
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func traitsMap(t Traits) mmdbtype.Map {
	m := mmdbtype.Map{}
	if t.IsAnonymous {
		m[mmdbtype.String("is_anonymous")] = mmdbtype.Bool(true)
	}
	if t.IsAnonymousVPN {
		m[mmdbtype.String("is_anonymous_vpn")] = mmdbtype.Bool(true)
	}
	if t.IsHostingProvider {
		m[mmdbtype.String("is_hosting_provider")] = mmdbtype.Bool(true)
	}
	if t.IsPublicProxy {
		m[mmdbtype.String("is_public_proxy")] = mmdbtype.Bool(true)
	}
	if t.IsTorExitNode {
		m[mmdbtype.String("is_tor_exit_node")] = mmdbtype.Bool(true)
	}
	if t.IsCDN {
		m[mmdbtype.String("is_cdn")] = mmdbtype.Bool(true)
	}
	if t.IsRelay {
		m[mmdbtype.String("is_relay")] = mmdbtype.Bool(true)
	}
	if len(m) == 0 {
		return nil
	}
	return m
}
