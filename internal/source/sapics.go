package source

import (
	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/schema"
)

type sapicsASNAdapter struct{}

func (sapicsASNAdapter) ID() string { return "sapics_origin_asn" }

type sapicsASNRow struct {
	ASNNumber uint32 `maxminddb:"autonomous_system_number"`
	ASN       uint32 `maxminddb:"asn"`
	ASNStr    string `maxminddb:"as_number"`
	Org       string `maxminddb:"autonomous_system_organization"`
}

func (sapicsASNAdapter) Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	return loadMMDB(path, spec.ID, func(row sapicsASNRow) schema.Record {
		n := row.ASNNumber
		if n == 0 {
			n = row.ASN
		}
		if n == 0 {
			n = parseASN(row.ASNStr)
		}
		return schema.Record{ASN: schema.ASN{Number: n, Organization: row.Org}}
	})
}

type sapicsCountryAdapter struct{}

func (sapicsCountryAdapter) ID() string { return "sapics_country" }

type sapicsCountryRow struct {
	CountryCode string `maxminddb:"country_code"`
	Country     string `maxminddb:"country"`
}

func (sapicsCountryAdapter) Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	return loadMMDB(path, spec.ID, func(row sapicsCountryRow) schema.Record {
		iso := firstNonEmpty(row.CountryCode, row.Country)
		return schema.Record{Country: schema.CountryFromISO(iso)}
	})
}
