package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Flags struct {
	IsHostingProvider bool `yaml:"is_hosting_provider"`
	IsCDN             bool `yaml:"is_cdn"`
	IsAnonymousVPN    bool `yaml:"is_anonymous_vpn"`
	IsTorExitNode     bool `yaml:"is_tor_exit_node"`
	IsRelay           bool `yaml:"is_relay"`
	IsAnonymous       bool `yaml:"is_anonymous"`
	IsPublicProxy     bool `yaml:"is_public_proxy"`
}

type SourceSpec struct {
	ID          string   `yaml:"id"`
	Enabled     bool     `yaml:"enabled"`
	Kind        string   `yaml:"kind"`
	URL         string   `yaml:"url"`
	Path        string   `yaml:"path"`
	License     string   `yaml:"license"`
	Attribution string   `yaml:"attribution"`
	RequiresEnv []string `yaml:"requires_env"`
	Required    bool     `yaml:"required"`
	ExtractGlob string   `yaml:"extract_glob"`
	ExtractTo   string   `yaml:"extract_to"`
	Flags       Flags    `yaml:"flags"`
}

type File struct {
	DownloadDir            string       `yaml:"download_dir"`
	OutputDir              string       `yaml:"output_dir"`
	OutputName             string       `yaml:"output_name"`
	MaxBytes               int64        `yaml:"max_bytes"`
	DownloadTimeoutSeconds int          `yaml:"download_timeout_seconds"`
	DownloadRetries        int          `yaml:"download_retries"`
	DownloadConcurrency    int          `yaml:"download_concurrency"`
	Sources                []SourceSpec `yaml:"sources"`
}

func Load(path string) (File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	var f File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return File{}, err
	}
	if f.DownloadDir == "" {
		f.DownloadDir = "download"
	}
	if f.OutputDir == "" {
		f.OutputDir = "dist"
	}
	if f.OutputName == "" {
		f.OutputName = "Superior-IP.mmdb"
	}
	if f.MaxBytes == 0 {
		f.MaxBytes = 1073741824
	}
	if f.DownloadTimeoutSeconds == 0 {
		f.DownloadTimeoutSeconds = 300
	}
	if f.DownloadRetries == 0 {
		f.DownloadRetries = 3
	}
	if f.DownloadConcurrency == 0 {
		f.DownloadConcurrency = 6
	}
	return f, nil
}

func (s SourceSpec) ExpandURL() string {
	return s.ExpandURLAt(time.Now().UTC())
}

func (s SourceSpec) ExpandURLAt(t time.Time) string {
	u := strings.ReplaceAll(s.URL, "${YYYYMM}", t.UTC().Format("2006-01"))
	return os.ExpandEnv(u)
}

func (s SourceSpec) LoadPath() string {
	if s.ExtractTo != "" && (s.Kind == "maxmind_tar_gz" || s.Kind == "gzip" || s.Kind == "zip") {
		return s.ExtractTo
	}
	return s.Path
}

func (s SourceSpec) MissingEnv() []string {
	var missing []string
	for _, name := range s.RequiresEnv {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			missing = append(missing, name)
		}
	}
	return missing
}

func (s SourceSpec) String() string {
	return fmt.Sprintf("source %s", s.ID)
}
