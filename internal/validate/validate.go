package validate

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/shafqat-a/iplegence/internal/lookup"
)

type Expect struct {
	IP         string `json:"ip"`
	CountryISO string `json:"country_iso"`
	ASN        uint32 `json:"asn"`
	Hosting    *bool  `json:"is_hosting_provider"`
	CDN        *bool  `json:"is_cdn"`
}

func LoadCases(path string) ([]Expect, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cases []Expect
	if err := json.Unmarshal(b, &cases); err != nil {
		return nil, err
	}
	return cases, nil
}

func Check(dbPath string, cases []Expect) []error {
	db, err := lookup.Open(dbPath)
	if err != nil {
		return []error{err}
	}
	defer db.Close()
	var errs []error
	for _, c := range cases {
		rec, ok, err := db.Lookup(c.IP)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", c.IP, err))
			continue
		}
		if !ok {
			errs = append(errs, fmt.Errorf("%s: not found", c.IP))
			continue
		}
		if c.CountryISO != "" && rec.Country.ISOCode != c.CountryISO {
			errs = append(errs, fmt.Errorf("%s: country %s want %s", c.IP, rec.Country.ISOCode, c.CountryISO))
		}
		if c.ASN != 0 && rec.ASN.Number != c.ASN {
			errs = append(errs, fmt.Errorf("%s: asn %d want %d", c.IP, rec.ASN.Number, c.ASN))
		}
		if c.Hosting != nil && rec.Traits.IsHostingProvider != *c.Hosting {
			errs = append(errs, fmt.Errorf("%s: is_hosting_provider %v want %v", c.IP, rec.Traits.IsHostingProvider, *c.Hosting))
		}
		if c.CDN != nil && rec.Traits.IsCDN != *c.CDN {
			errs = append(errs, fmt.Errorf("%s: is_cdn %v want %v", c.IP, rec.Traits.IsCDN, *c.CDN))
		}
	}
	return errs
}
