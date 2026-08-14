package source

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/schema"
)

type overlayAdapter struct{ id string }

func (a overlayAdapter) ID() string { return a.id }

func flagsRecord(f config.Flags) schema.Record {
	return schema.Record{Traits: schema.Traits{
		IsHostingProvider: f.IsHostingProvider,
		IsCDN:             f.IsCDN,
		IsAnonymousVPN:    f.IsAnonymousVPN,
		IsTorExitNode:     f.IsTorExitNode,
		IsRelay:           f.IsRelay,
		IsAnonymous:       f.IsAnonymous,
		IsPublicProxy:     f.IsPublicProxy,
	}}
}

func (a overlayAdapter) Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	rec := flagsRecord(spec.Flags)
	if rec.IsEmpty() {
		return nil, nil
	}
	var prefixes []string
	var err error
	switch spec.Kind {
	case "csv":
		prefixes, err = loadCSVFirstColumn(path)
	case "ripe_announced":
		prefixes, err = loadRipeAnnounced(path)
	case "nordvpn_json":
		prefixes, err = loadNordVPN(path)
	default:
		prefixes, err = loadLines(path)
	}
	if err != nil {
		return nil, err
	}
	var blocks []cidr.Block
	for _, s := range prefixes {
		pfx, err := prefixFromString(s)
		if err != nil || !pfx.IsValid() || pfx.Addr().Is4In6() {
			continue
		}
		blocks = append(blocks, cidr.Block{Prefix: pfx, Rec: rec, Source: spec.ID})
	}
	return blocks, nil
}

func loadCSVFirstColumn(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.ReuseRecord = true
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	var out []string
	first := true
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(rec) == 0 {
			continue
		}
		cell := firstPrefixToken(rec[0])
		if first {
			first = false
			if cell == "" {
				continue // header
			}
		}
		if cell != "" {
			out = append(out, cell)
		}
	}
	return out, nil
}

func firstPrefixToken(line string) string {
	fields := strings.FieldsFunc(strings.TrimSpace(line), func(r rune) bool {
		return r == '|' || r == ',' || unicode.IsSpace(r)
	})
	for _, f := range fields {
		f = strings.Trim(f, `"'`)
		if _, err := prefixFromString(f); err == nil {
			return f
		}
	}
	return ""
}

func loadRipeAnnounced(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Data struct {
			Prefixes []struct {
				Prefix string `json:"prefix"`
			} `json:"prefixes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	var out []string
	for _, p := range doc.Data.Prefixes {
		if p.Prefix != "" {
			out = append(out, p.Prefix)
		}
	}
	return out, nil
}

func loadNordVPN(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var servers []struct {
		Station string `json:"station"`
		IPs     []struct {
			IP struct {
				IP string `json:"ip"`
			} `json:"ip"`
		} `json:"ips"`
	}
	if err := json.Unmarshal(b, &servers); err != nil {
		return nil, err
	}
	var out []string
	for _, s := range servers {
		if s.Station != "" {
			out = append(out, s.Station)
		}
		for _, ip := range s.IPs {
			if ip.IP.IP != "" {
				out = append(out, ip.IP.IP)
			}
		}
	}
	return out, nil
}
