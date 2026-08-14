package attrib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shafqat-a/iplegence/internal/config"
)

func Render(cfg config.File, used []string, dest string) error {
	usedSet := map[string]bool{}
	for _, id := range used {
		usedSet[id] = true
	}
	var b strings.Builder
	b.WriteString("# Attribution\n\n")
	b.WriteString("This product merges publicly available IP intelligence sources.\n")
	b.WriteString("It does not claim commercial-grade accuracy.\n\n")
	b.WriteString("## Sources used\n")

	byID := map[string]config.SourceSpec{}
	for _, s := range cfg.Sources {
		byID[s.ID] = s
	}
	for _, id := range used {
		s, ok := byID[id]
		if !ok {
			fmt.Fprintf(&b, "\n### %s\n\nUnknown source.\n", id)
			continue
		}
		writeSource(&b, s)
	}

	var unused []config.SourceSpec
	for _, s := range cfg.Sources {
		if !usedSet[s.ID] {
			unused = append(unused, s)
		}
	}
	if len(unused) > 0 {
		b.WriteString("\n## Configured but unused\n")
		for _, s := range unused {
			writeSource(&b, s)
		}
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dest)
}

func writeSource(b *strings.Builder, s config.SourceSpec) {
	fmt.Fprintf(b, "\n### %s\n\n", s.ID)
	fmt.Fprintf(b, "- License: %s\n", s.License)
	if s.Attribution != "" {
		fmt.Fprintf(b, "- %s\n", s.Attribution)
	}
}
