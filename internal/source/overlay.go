package source

import (
	"encoding/csv"
	"io"
	"os"

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
		cell := rec[0]
		if first {
			first = false
			if _, err := prefixFromString(cell); err != nil {
				continue // header
			}
		}
		out = append(out, cell)
	}
	return out, nil
}
