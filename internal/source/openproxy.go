package source

import (
	"encoding/csv"
	"io"
	"net/netip"
	"os"
	"strings"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/schema"
	"go4.org/netipx"
)

type openproxyAdapter struct{}

func (openproxyAdapter) ID() string { return "openproxydb" }

func (openproxyAdapter) Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.ReuseRecord = true
	r.LazyQuotes = true
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	idx := map[string]int{}
	for i, h := range header {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}
	col := func(name string) int {
		i, ok := idx[name]
		if !ok {
			return -1
		}
		return i
	}
	iNet, iCIDR := col("network"), col("cidr")
	iStart, iEnd := col("start"), col("end")
	iProxy, iVPN := col("is_proxy"), col("is_vpn")
	iTor, iHost := col("is_tor"), col("is_hosting")
	iCDN, iAnon := col("is_cdn"), col("is_anonymous")
	iRelay := col("is_relay")
	get := func(rec []string, i int) string {
		if i < 0 || i >= len(rec) {
			return ""
		}
		return rec[i]
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
		traits := schema.Traits{
			IsPublicProxy:     parseBool(get(rec, iProxy)),
			IsAnonymousVPN:    parseBool(get(rec, iVPN)),
			IsTorExitNode:     parseBool(get(rec, iTor)),
			IsHostingProvider: parseBool(get(rec, iHost)),
			IsCDN:             parseBool(get(rec, iCDN)),
			IsAnonymous:       parseBool(get(rec, iAnon)),
			IsRelay:           parseBool(get(rec, iRelay)),
		}
		row := schema.Record{Traits: traits}
		if row.IsEmpty() {
			continue
		}
		var prefs []netip.Prefix
		if p := firstNonEmpty(get(rec, iNet), get(rec, iCIDR)); p != "" {
			pfx, err := prefixFromString(p)
			if err != nil {
				continue
			}
			prefs = []netip.Prefix{pfx}
		} else if s, e := get(rec, iStart), get(rec, iEnd); s != "" && e != "" {
			from, err1 := netip.ParseAddr(s)
			to, err2 := netip.ParseAddr(e)
			if err1 != nil || err2 != nil {
				continue
			}
			ir := netipx.IPRangeFrom(from, to)
			if !ir.IsValid() {
				continue
			}
			prefs = ir.Prefixes()
		} else {
			continue
		}
		for _, pfx := range prefs {
			blocks = append(blocks, cidr.Block{Prefix: pfx, Rec: row, Source: spec.ID})
		}
	}
	return blocks, nil
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y", "t":
		return true
	default:
		return false
	}
}
