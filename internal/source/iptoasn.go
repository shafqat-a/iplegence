package source

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strings"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/schema"
	"go4.org/netipx"
)

type iptoasnAdapter struct{}

func (iptoasnAdapter) ID() string { return "iptoasn" }

func (iptoasnAdapter) Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var r io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		r = gz
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var blocks []cidr.Block
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 5 {
			parts = strings.Fields(line)
		}
		if len(parts) < 5 {
			continue
		}
		from, err1 := netip.ParseAddr(parts[0])
		to, err2 := netip.ParseAddr(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}
		asn := parseASN(parts[2])
		if asn == 0 {
			continue
		}
		rec := schema.Record{
			Country: schema.CountryFromISO(parts[3]),
			ASN:     schema.ASN{Number: asn, Organization: parts[4]},
		}
		ir := netipx.IPRangeFrom(from, to)
		if !ir.IsValid() {
			return nil, fmt.Errorf("iptoasn invalid range %s-%s", from, to)
		}
		for _, pfx := range ir.Prefixes() {
			blocks = append(blocks, cidr.Block{Prefix: pfx, Rec: rec, Source: spec.ID})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return blocks, nil
}
