package pipeline_test

import (
	"context"
	"net/netip"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/oschwald/maxminddb-golang/v2"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/pipeline"
	"github.com/shafqat-a/iplegence/internal/schema"
)

func TestRunSkipDownload(t *testing.T) {
	root := repoRoot(t)
	dir := t.TempDir()
	cfg := config.File{
		DownloadDir: dir,
		OutputDir:   dir,
		OutputName:  "Superior-IP.mmdb",
		MaxBytes:    10 << 20,
		Sources: []config.SourceSpec{
			{ID: "iptoasn", Enabled: true, Kind: "tsv_gz", Path: filepath.Join(root, "testdata", "iptoasn.tsv"), Required: true},
			{ID: "openproxydb", Enabled: true, Kind: "csv", Path: filepath.Join(root, "testdata", "openproxy.csv"), Required: true},
			{ID: "aws_ranges", Enabled: true, Kind: "json", Path: filepath.Join(root, "testdata", "aws.json"), Flags: config.Flags{IsHostingProvider: true}, Required: true},
			{ID: "cloudflare_v4", Enabled: true, Kind: "lines", Path: filepath.Join(root, "testdata", "cloudflare-v4.txt"), Flags: config.Flags{IsCDN: true, IsHostingProvider: true}, Required: true},
		},
	}
	cfgPath := filepath.Join(dir, "sources.yaml")
	b, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, b, 0o644); err != nil {
		t.Fatal(err)
	}
	pri := filepath.Join(root, "configs", "priority.yaml")
	if err := pipeline.Run(context.Background(), pipeline.Options{
		ConfigPath:   cfgPath,
		PriorityPath: pri,
		SkipDownload: true,
	}); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "Superior-IP.mmdb")
	if _, err := os.Stat(dest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dest + ".sha256"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "ATTRIBUTION.md")); err != nil {
		t.Fatal(err)
	}
	db, err := maxminddb.Open(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var rec schema.LookupRecord
	if err := db.Lookup(netip.MustParseAddr("8.8.8.8")).Decode(&rec); err != nil {
		t.Fatal(err)
	}
	if rec.Country.ISOCode != "US" || rec.ASN.Number != 15169 {
		t.Fatalf("8.8.8.8 got %+v", rec)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("go.mod not found")
	return ""
}
