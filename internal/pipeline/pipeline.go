package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/shafqat-a/iplegence/internal/attrib"
	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/download"
	"github.com/shafqat-a/iplegence/internal/merge"
	"github.com/shafqat-a/iplegence/internal/mmdb"
	"github.com/shafqat-a/iplegence/internal/source"
	"github.com/shafqat-a/iplegence/internal/usage"
)

type Options struct {
	ConfigPath   string
	SkipDownload bool
	PriorityPath string
}

func Run(ctx context.Context, opt Options) error {
	start := time.Now()
	cfg, err := config.Load(opt.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pri, err := merge.LoadPriority(opt.PriorityPath)
	if err != nil {
		return fmt.Errorf("load priority: %w", err)
	}
	if !opt.SkipDownload {
		results := download.FetchAll(ctx, cfg)
		if err := download.RequiredError(results); err != nil {
			return fmt.Errorf("download: %w", err)
		}
		for _, r := range results {
			if r.Err != nil {
				log.Printf("skip %s: %v", r.Source.ID, r.Err)
			}
		}
	}

	var blocks []cidr.Block
	var used []string
	catalog := usage.Catalog{}
	for _, spec := range cfg.Sources {
		if !spec.Enabled {
			continue
		}
		path := spec.LoadPath()
		if _, err := os.Stat(path); err != nil {
			if spec.Required && !opt.SkipDownload {
				return fmt.Errorf("required source %s missing file %s", spec.ID, path)
			}
			log.Printf("source %s: no file, skipping", spec.ID)
			continue
		}
		if usage.IsCatalog(spec) {
			cat, err := usage.Load(path)
			if err != nil {
				if spec.Required {
					return fmt.Errorf("load %s: %w", spec.ID, err)
				}
				log.Printf("source %s: %v", spec.ID, err)
				continue
			}
			log.Printf("source %s: %d asns", spec.ID, len(cat))
			for asn, types := range cat {
				catalog[asn] = types
			}
			used = append(used, spec.ID)
			continue
		}
		part, err := source.Load(path, spec)
		if err != nil {
			if spec.Required {
				return fmt.Errorf("load %s: %w", spec.ID, err)
			}
			log.Printf("source %s: %v", spec.ID, err)
			continue
		}
		log.Printf("source %s: %d blocks", spec.ID, len(part))
		if len(part) > 0 {
			used = append(used, spec.ID)
			blocks = append(blocks, part...)
		}
	}

	rows, err := cidr.MergeBlocks(blocks, pri)
	if err != nil {
		return fmt.Errorf("merge: %w", err)
	}
	usage.Apply(rows, catalog)
	dest := filepath.Join(cfg.OutputDir, cfg.OutputName)
	res, err := mmdb.Write(rows, dest, cfg.MaxBytes)
	if err != nil {
		return fmt.Errorf("write mmdb: %w", err)
	}
	attr := filepath.Join(cfg.OutputDir, "ATTRIBUTION.md")
	if err := attrib.Render(cfg, used, attr); err != nil {
		return fmt.Errorf("attribution: %w", err)
	}
	log.Printf("wrote %s (%.2f MB, %d rows, sha256=%s) in %s",
		res.Path, float64(res.Bytes)/(1<<20), len(rows), res.SHA256, time.Since(start).Round(time.Millisecond))
	return nil
}
