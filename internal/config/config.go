package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Flags struct {
	IsHostingProvider bool `yaml:"is_hosting_provider"`
	IsCDN             bool `yaml:"is_cdn"`
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
		f.MaxBytes = 188743680
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
	return os.ExpandEnv(s.URL)
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
