package config_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shafqat-a/iplegence/internal/config"
)

func TestLoadSourcesYAML(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "..", "..")
	cfg, err := config.Load(filepath.Join(root, "configs", "sources.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OutputName != "Superior-IP.mmdb" {
		t.Fatalf("output name %q", cfg.OutputName)
	}
	if cfg.MaxBytes != 188743680 {
		t.Fatalf("max_bytes %d", cfg.MaxBytes)
	}
	if len(cfg.Sources) < 10 {
		t.Fatalf("expected phase-1 sources, got %d", len(cfg.Sources))
	}
	var sawIPinfo bool
	for _, s := range cfg.Sources {
		if s.ID == "ipinfo_lite" {
			sawIPinfo = true
			if !s.Required {
				t.Fatal("ipinfo_lite must be required")
			}
		}
	}
	if !sawIPinfo {
		t.Fatal("missing ipinfo_lite")
	}
}

func TestMissingEnv(t *testing.T) {
	t.Setenv("IPINFO_TOKEN", "")
	s := config.SourceSpec{RequiresEnv: []string{"IPINFO_TOKEN"}}
	if got := s.MissingEnv(); len(got) != 1 || got[0] != "IPINFO_TOKEN" {
		t.Fatalf("got %v", got)
	}
	t.Setenv("IPINFO_TOKEN", "tok")
	if got := s.MissingEnv(); len(got) != 0 {
		t.Fatalf("expected none, got %v", got)
	}
}
