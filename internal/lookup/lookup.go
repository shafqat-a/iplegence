package lookup

import (
	"fmt"
	"net/netip"

	"github.com/oschwald/maxminddb-golang/v2"
	"github.com/shafqat-a/iplegence/internal/schema"
)

type DB struct {
	r *maxminddb.Reader
}

func Open(path string) (*DB, error) {
	r, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}
	return &DB{r: r}, nil
}

func (d *DB) Close() error {
	if d == nil || d.r == nil {
		return nil
	}
	return d.r.Close()
}

func (d *DB) Lookup(ip string) (schema.LookupRecord, bool, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return schema.LookupRecord{}, false, fmt.Errorf("bad ip: %w", err)
	}
	var rec schema.LookupRecord
	res := d.r.Lookup(addr)
	if err := res.Decode(&rec); err != nil {
		return schema.LookupRecord{}, false, err
	}
	if !rec.Found() {
		return schema.LookupRecord{}, false, nil
	}
	return rec, true, nil
}
