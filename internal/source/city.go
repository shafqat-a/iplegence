package source

import (
	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/schema"
)

type cityAdapter struct{ id string }

func (a cityAdapter) ID() string { return a.id }

type cityNames struct {
	En string `maxminddb:"en"`
}

type cityRow struct {
	Country struct {
		ISOCode   string    `maxminddb:"iso_code"`
		GeonameID uint32    `maxminddb:"geoname_id"`
		Names     cityNames `maxminddb:"names"`
	} `maxminddb:"country"`
	Continent struct {
		Code      string    `maxminddb:"code"`
		GeonameID uint32    `maxminddb:"geoname_id"`
		Names     cityNames `maxminddb:"names"`
	} `maxminddb:"continent"`
	City struct {
		GeonameID uint32    `maxminddb:"geoname_id"`
		Names     cityNames `maxminddb:"names"`
	} `maxminddb:"city"`
	Location struct {
		Latitude       float64 `maxminddb:"latitude"`
		Longitude      float64 `maxminddb:"longitude"`
		AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
		TimeZone       string  `maxminddb:"time_zone"`
	} `maxminddb:"location"`
	Subdivisions []struct {
		GeonameID uint32    `maxminddb:"geoname_id"`
		ISOCode   string    `maxminddb:"iso_code"`
		Names     cityNames `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
	Postal struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"postal"`
}

func (a cityAdapter) Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	return loadMMDB(path, spec.ID, func(row cityRow) schema.Record {
		rec := schema.Record{}
		if iso := schema.NormalizeISO(row.Country.ISOCode); iso != "" {
			rec.Country = schema.Country{
				ISOCode:   iso,
				GeonameID: row.Country.GeonameID,
				Names:     schema.Names{En: firstNonEmpty(row.Country.Names.En, schema.CountryName(iso))},
			}
		}
		if row.Continent.Code != "" {
			code := row.Continent.Code
			rec.Continent = schema.Continent{
				Code:      code,
				GeonameID: row.Continent.GeonameID,
				Names:     schema.Names{En: firstNonEmpty(row.Continent.Names.En, schema.ContinentName(code))},
			}
		}
		if row.City.Names.En != "" || row.City.GeonameID != 0 {
			rec.City = schema.City{GeonameID: row.City.GeonameID, Names: schema.Names{En: row.City.Names.En}}
		}
		if row.Location.Latitude != 0 || row.Location.Longitude != 0 || row.Location.TimeZone != "" {
			rec.Location = schema.Location{
				Latitude:       row.Location.Latitude,
				Longitude:      row.Location.Longitude,
				AccuracyRadius: row.Location.AccuracyRadius,
				TimeZone:       row.Location.TimeZone,
				HasCoordinates: row.Location.Latitude != 0 || row.Location.Longitude != 0,
			}
		}
		for _, s := range row.Subdivisions {
			if s.ISOCode == "" && s.Names.En == "" && s.GeonameID == 0 {
				continue
			}
			rec.Subdivisions = append(rec.Subdivisions, schema.Subdivision{
				GeonameID: s.GeonameID,
				ISOCode:   s.ISOCode,
				Names:     schema.Names{En: s.Names.En},
			})
		}
		rec.Postal.Code = row.Postal.Code
		return rec
	})
}
