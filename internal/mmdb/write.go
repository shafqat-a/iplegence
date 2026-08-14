package mmdb

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/maxmind/mmdbwriter"
	"github.com/shafqat-a/iplegence/internal/cidr"
)

type WriteResult struct {
	Path   string
	Bytes  int64
	SHA256 string
}

var ErrTooLarge = errors.New("mmdb exceeds max_bytes")

func Write(rows []cidr.Row, dest string, maxBytes int64) (WriteResult, error) {
	tree, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType:            "iplegence-Superior-IP",
		Description:             map[string]string{"en": "Merged free/open IP intelligence (country, ASN, traits)"},
		IPVersion:               6,
		RecordSize:              28,
		DisableIPv4Aliasing:     false,
		IncludeReservedNetworks: true,
	})
	if err != nil {
		return WriteResult{}, err
	}
	for _, row := range rows {
		m := row.Rec.ToMMDBType()
		if m == nil {
			continue
		}
		_, ipnet, err := net.ParseCIDR(row.Prefix.String())
		if err != nil {
			return WriteResult{}, fmt.Errorf("parse %s: %w", row.Prefix, err)
		}
		if err := tree.Insert(ipnet, m); err != nil {
			return WriteResult{}, fmt.Errorf("insert %s: %w", row.Prefix, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return WriteResult{}, err
	}
	tmp := dest + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return WriteResult{}, err
	}
	bw := bufio.NewWriterSize(f, 4<<20)
	if _, err := tree.WriteTo(bw); err != nil {
		f.Close()
		os.Remove(tmp)
		return WriteResult{}, err
	}
	if err := bw.Flush(); err != nil {
		f.Close()
		os.Remove(tmp)
		return WriteResult{}, err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return WriteResult{}, err
	}
	info, err := os.Stat(tmp)
	if err != nil {
		os.Remove(tmp)
		return WriteResult{}, err
	}
	if maxBytes > 0 && info.Size() > maxBytes {
		os.Remove(tmp)
		os.Remove(dest)
		return WriteResult{}, fmt.Errorf("%w: %d > %d", ErrTooLarge, info.Size(), maxBytes)
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return WriteResult{}, err
	}
	sum, err := fileSHA256(dest)
	if err != nil {
		return WriteResult{}, err
	}
	sumPath := dest + ".sha256"
	line := sum + "  " + filepath.Base(dest) + "\n"
	if err := os.WriteFile(sumPath, []byte(line), 0o644); err != nil {
		return WriteResult{}, err
	}
	return WriteResult{Path: dest, Bytes: info.Size(), SHA256: sum}, nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
