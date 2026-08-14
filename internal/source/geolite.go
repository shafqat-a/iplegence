package source

import (
	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/config"
	"github.com/shafqat-a/iplegence/internal/schema"
)

type geoliteAdapter struct{}

func (geoliteAdapter) ID() string { return "geolite2_asn" }

type geoliteASNRow struct {
	Number uint32 `maxminddb:"autonomous_system_number"`
	Org    string `maxminddb:"autonomous_system_organization"`
}

func (geoliteAdapter) Load(path string, spec config.SourceSpec) ([]cidr.Block, error) {
	return loadMMDB(path, spec.ID, func(row geoliteASNRow) schema.Record {
		return schema.Record{ASN: schema.ASN{Number: row.Number, Organization: row.Org}}
	})
}
