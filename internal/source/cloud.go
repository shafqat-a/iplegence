package source

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/schema"
)

type cloudAdapter struct{ id string }

func (a cloudAdapter) ID() string { return a.id }

func (a cloudAdapter) Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	rec := schema.Record{Traits: schema.Traits{
		IsHostingProvider: spec.Flags.IsHostingProvider,
		IsCDN:             spec.Flags.IsCDN,
	}}
	var prefixes []string
	var err error
	switch spec.ID {
	case "aws_ranges":
		prefixes, err = loadAWS(path)
	case "gcp_ranges":
		prefixes, err = loadGCP(path)
	case "azure_ranges":
		prefixes, err = loadAzure(path)
	case "cloudflare_v4", "cloudflare_v6":
		prefixes, err = loadLines(path)
	default:
		return nil, fmt.Errorf("unknown cloud source %q", spec.ID)
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

func loadAWS(path string) ([]string, error) {
	var doc struct {
		Prefixes []struct {
			IPPrefix string `json:"ip_prefix"`
		} `json:"prefixes"`
		IPv6Prefixes []struct {
			IPv6Prefix string `json:"ipv6_prefix"`
		} `json:"ipv6_prefixes"`
	}
	if err := readJSON(path, &doc); err != nil {
		return nil, err
	}
	var out []string
	for _, p := range doc.Prefixes {
		if p.IPPrefix != "" {
			out = append(out, p.IPPrefix)
		}
	}
	for _, p := range doc.IPv6Prefixes {
		if p.IPv6Prefix != "" {
			out = append(out, p.IPv6Prefix)
		}
	}
	return out, nil
}

func loadGCP(path string) ([]string, error) {
	var doc struct {
		Prefixes []struct {
			IPv4Prefix string `json:"ipv4Prefix"`
			IPv6Prefix string `json:"ipv6Prefix"`
		} `json:"prefixes"`
	}
	if err := readJSON(path, &doc); err != nil {
		return nil, err
	}
	var out []string
	for _, p := range doc.Prefixes {
		if p.IPv4Prefix != "" {
			out = append(out, p.IPv4Prefix)
		}
		if p.IPv6Prefix != "" {
			out = append(out, p.IPv6Prefix)
		}
	}
	return out, nil
}

func loadAzure(path string) ([]string, error) {
	var doc struct {
		Values []struct {
			Properties struct {
				AddressPrefixes []string `json:"addressPrefixes"`
			} `json:"properties"`
		} `json:"values"`
	}
	if err := readJSON(path, &doc); err != nil {
		return nil, err
	}
	var out []string
	for _, v := range doc.Values {
		out = append(out, v.Properties.AddressPrefixes...)
	}
	return out, nil
}

func loadLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if p := firstPrefixToken(line); p != "" {
			out = append(out, p)
		}
	}
	return out, sc.Err()
}

func readJSON(path string, dest any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}
