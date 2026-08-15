package usage

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shafqat-a/iplegence/internal/config"
)

type Catalog map[uint32][]string

type Network struct {
	ASN       uint32   `json:"asn"`
	InfoType  string   `json:"info_type"`
	InfoTypes []string `json:"info_types"`
}

func (n Network) Types() []string {
	seen := map[string]bool{}
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, t := range n.InfoTypes {
		add(t)
	}
	add(n.InfoType)
	return out
}

type fileShape struct {
	ASNs map[string][]string `json:"asns"`
	Data []Network           `json:"data"`
}

func IsCatalog(spec config.SourceSpec) bool {
	return spec.Kind == "peeringdb_api" || spec.Kind == "peeringdb_json"
}

func Load(path string) (Catalog, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc fileShape
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("peeringdb: %w", err)
	}
	out := Catalog{}
	for key, types := range doc.ASNs {
		asn, err := strconv.ParseUint(key, 10, 32)
		if err != nil || asn == 0 {
			continue
		}
		out.add(uint32(asn), types)
	}
	for _, n := range doc.Data {
		out.add(n.ASN, n.Types())
	}
	return out, nil
}

func CompactJSON(cat Catalog) ([]byte, error) {
	asns := make(map[string][]string, len(cat))
	for asn, types := range cat {
		if asn == 0 || len(types) == 0 {
			continue
		}
		asns[strconv.FormatUint(uint64(asn), 10)] = types
	}
	return json.Marshal(fileShape{ASNs: asns})
}

func (c Catalog) AddNetwork(n Network) {
	c.add(n.ASN, n.Types())
}

func (c Catalog) add(asn uint32, types []string) {
	if asn == 0 {
		return
	}
	if len(types) == 0 {
		return
	}
	c[asn] = mergeTypes(c[asn], types)
}

func mergeTypes(dst, src []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range dst {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	for _, t := range src {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}
