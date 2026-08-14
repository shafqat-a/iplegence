package source

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/schema"
	"go4.org/netipx"
)

type ipapiAdapter struct{ id string }

func (a ipapiAdapter) ID() string { return a.id }

func (a ipapiAdapter) Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.ReuseRecord = true
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	idx := map[string]int{}
	for i, h := range header {
		idx[normalizeHeader(h)] = i
	}
	col := func(names ...string) int {
		for _, n := range names {
			if i, ok := idx[normalizeHeader(n)]; ok {
				return i
			}
		}
		return -1
	}
	iStart := col("start_ip", "startip")
	iEnd := col("end_ip", "endip")
	iCont := col("continent")
	iCC := col("country_code", "countrycode")
	iCity := col("city")
	iState := col("state")
	iZip := col("zip")
	iTZ := col("timezone")
	iLat := col("latitude")
	iLon := col("longitude")
	if iStart < 0 || iEnd < 0 {
		return nil, fmt.Errorf("ipapi csv missing start/end columns: %v", header)
	}
	get := func(rec []string, i int) string {
		if i < 0 || i >= len(rec) {
			return ""
		}
		return strings.TrimSpace(rec[i])
	}
	var blocks []cidr.Block
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		from, err1 := netip.ParseAddr(get(rec, iStart))
		to, err2 := netip.ParseAddr(get(rec, iEnd))
		if err1 != nil || err2 != nil {
			continue
		}
		iso := schema.NormalizeISO(get(rec, iCC))
		row := schema.Record{
			Country:   schema.CountryFromISO(iso),
			Continent: schema.ContinentFromCode(get(rec, iCont)),
		}
		if city := get(rec, iCity); city != "" {
			row.City = schema.City{Names: schema.Names{En: city}}
		}
		if st := get(rec, iState); st != "" {
			row.Subdivisions = []schema.Subdivision{{Names: schema.Names{En: st}}}
		}
		row.Postal.Code = get(rec, iZip)
		lat := parseFloat(get(rec, iLat))
		lon := parseFloat(get(rec, iLon))
		tz := get(rec, iTZ)
		if lat != 0 || lon != 0 || tz != "" {
			row.Location = schema.Location{
				Latitude:       lat,
				Longitude:      lon,
				TimeZone:       tz,
				HasCoordinates: lat != 0 || lon != 0,
			}
		}
		if row.IsEmpty() {
			continue
		}
		ir := netipx.IPRangeFrom(from, to)
		if !ir.IsValid() {
			continue
		}
		for _, pfx := range ir.Prefixes() {
			blocks = append(blocks, cidr.Block{Prefix: pfx, Rec: row, Source: spec.ID})
		}
	}
	return blocks, nil
}

func normalizeHeader(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
