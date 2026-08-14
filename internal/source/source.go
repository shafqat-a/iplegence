package source

import (
	"fmt"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
)

type Block = cidr.Block

type Adapter interface {
	ID() string
	Load(path string, spec config.SourceSpec) ([]cidr.Block, error)
}

func Lookup(id string) (Adapter, bool) {
	a, ok := adapters[id]
	return a, ok
}

func Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	a, ok := Lookup(spec.ID)
	if !ok {
		return nil, fmt.Errorf("no adapter for source %q", spec.ID)
	}
	return a.Load(path, spec)
}

var adapters = map[string]Adapter{
	"ipinfo_lite":       ipinfoAdapter{},
	"sapics_origin_asn": sapicsASNAdapter{},
	"sapics_country":    sapicsCountryAdapter{},
	"geolite2_asn":      geoliteAdapter{},
	"geolite2_city":     cityAdapter{id: "geolite2_city"},
	"dbip_city":         cityAdapter{id: "dbip_city"},
	"ipapi_city_v4":     ipapiAdapter{id: "ipapi_city_v4"},
	"ipapi_city_v6":     ipapiAdapter{id: "ipapi_city_v6"},
	"iptoasn":           iptoasnAdapter{},
	"openproxydb":       openproxyAdapter{},
	"aws_ranges":        cloudAdapter{id: "aws_ranges"},
	"gcp_ranges":        cloudAdapter{id: "gcp_ranges"},
	"azure_ranges":      cloudAdapter{id: "azure_ranges"},
	"cloudflare_v4":     cloudAdapter{id: "cloudflare_v4"},
	"cloudflare_v6":     cloudAdapter{id: "cloudflare_v6"},
	"tor_exits":         overlayAdapter{id: "tor_exits"},
	"icloud_relay":        overlayAdapter{id: "icloud_relay"},
	"x4b_vpn":             overlayAdapter{id: "x4b_vpn"},
	"nordvpn":             overlayAdapter{id: "nordvpn"},
	"digitalocean_ranges": overlayAdapter{id: "digitalocean_ranges"},
	"ovh_ranges":          overlayAdapter{id: "ovh_ranges"},
	"hetzner_ranges":      overlayAdapter{id: "hetzner_ranges"},
}
