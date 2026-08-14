package source

import (
	"net/netip"
	"strconv"
	"strings"

	"github.com/oschwald/maxminddb-golang/v2"
	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/schema"
)

type ipinfoAdapter struct{}

func (ipinfoAdapter) ID() string { return "ipinfo_lite" }

type ipinfoRow struct {
	Country       string `maxminddb:"country"`
	CountryCode   string `maxminddb:"country_code"`
	Continent     string `maxminddb:"continent"`
	ContinentCode string `maxminddb:"continent_code"`
	ASN           string `maxminddb:"asn"`
	ASName        string `maxminddb:"as_name"`
	ASDomain      string `maxminddb:"as_domain"`
}

func (ipinfoAdapter) Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	return loadMMDB(path, spec.ID, func(row ipinfoRow) schema.Record {
		iso := firstNonEmpty(row.CountryCode, row.Country)
		cont := firstNonEmpty(row.ContinentCode, row.Continent)
		rec := schema.Record{
			Country:   schema.CountryFromISO(iso),
			Continent: schema.ContinentFromCode(cont),
			ASN: schema.ASN{
				Number:       parseASN(row.ASN),
				Organization: row.ASName,
				Domain:       row.ASDomain,
			},
		}
		return rec
	})
}

func loadMMDB[T any](path, sourceID string, mapFn func(T) schema.Record) ([]cidr.Block, error) {
	db, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var blocks []cidr.Block
	for res := range db.Networks() {
		if err := res.Err(); err != nil {
			return nil, err
		}
		var row T
		pfx := res.Prefix()
		if err := res.Decode(&row); err != nil {
			return nil, err
		}
		if !pfx.IsValid() || pfx.Addr().Is4In6() {
			continue
		}
		rec := mapFn(row)
		if rec.IsEmpty() {
			continue
		}
		blocks = append(blocks, cidr.Block{Prefix: pfx, Rec: rec, Source: sourceID})
	}
	return blocks, nil
}

func parseASN(s string) uint32 {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(strings.ToUpper(s), "AS")
	if s == "" {
		return 0
	}
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(n)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func prefixFromString(s string) (netip.Prefix, error) {
	if p, err := netip.ParsePrefix(s); err == nil {
		return p.Masked(), nil
	}
	if ip, err := netip.ParseAddr(s); err == nil {
		bits := 32
		if ip.Is6() {
			bits = 128
		}
		return netip.PrefixFrom(ip, bits), nil
	}
	return netip.Prefix{}, strconv.ErrSyntax
}
