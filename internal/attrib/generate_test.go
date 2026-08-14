package attrib_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shafqat-a/iplegence/internal/attrib"
	"github.com/shafqat-a/iplegence/internal/config"
)

func TestRenderIncludesIPinfoAndDisclaimer(t *testing.T) {
	cfg := config.File{
		Sources: []config.SourceSpec{
			{
				ID:          "ipinfo_lite",
				License:     "CC BY-SA 4.0",
				Attribution: "This product includes IP data from IPinfo (https://ipinfo.io).",
			},
			{
				ID:          "sapics_origin_asn",
				License:     "PDDL",
				Attribution: "ASN data from sapics/ip-location-db origin-asn (https://github.com/sapics/ip-location-db).",
			},
			{
				ID:          "geolite2_asn",
				License:     "CC BY-SA 4.0 + MaxMind GeoLite2 EULA",
				Attribution: "This product includes GeoLite2 Data created by MaxMind, available from https://www.maxmind.com.",
			},
		},
	}
	dest := filepath.Join(t.TempDir(), "ATTRIBUTION.md")
	if err := attrib.Render(cfg, []string{"ipinfo_lite", "sapics_origin_asn"}, dest); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	for _, want := range []string{"IPinfo", "sapics/ip-location-db", "commercial-grade", "## Configured but unused", "MaxMind"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
}
