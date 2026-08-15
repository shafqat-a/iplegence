package usage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/usage"
)

func TestLoadCompactAndRaw(t *testing.T) {
	dir := t.TempDir()
	compact := filepath.Join(dir, "compact.json")
	if err := os.WriteFile(compact, []byte(`{"asns":{"18":["Educational/Research"],"65000":["Government"]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cat, err := usage.Load(compact)
	if err != nil {
		t.Fatal(err)
	}
	if got := cat[18]; len(got) != 1 || got[0] != "Educational/Research" {
		t.Fatalf("compact 18: %v", got)
	}
	raw := filepath.Join(dir, "raw.json")
	if err := os.WriteFile(raw, []byte(`{"data":[{"asn":15169,"info_type":"Content","info_types":["Content"]}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cat, err = usage.Load(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := cat[15169]; len(got) != 1 || got[0] != "Content" {
		t.Fatalf("raw 15169: %v", got)
	}
}

func TestCompactJSONRoundTrip(t *testing.T) {
	b, err := usage.CompactJSON(usage.Catalog{18: {"Educational/Research"}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "out.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	cat, err := usage.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cat[18][0] != "Educational/Research" {
		t.Fatalf("got %v", cat[18])
	}
}

func TestIsCatalog(t *testing.T) {
	if !usage.IsCatalog(config.SourceSpec{Kind: "peeringdb_api"}) {
		t.Fatal("api")
	}
	if usage.IsCatalog(config.SourceSpec{Kind: "mmdb", ID: "ipinfo_lite"}) {
		t.Fatal("other")
	}
}
