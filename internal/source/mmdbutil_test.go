package source_test

import (
	"net"
	"os"
	"testing"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

func writeTestMMDB(t *testing.T, path string, pfx string, rec mmdbtype.DataType) {
	t.Helper()
	tree, err := mmdbwriter.New(mmdbwriter.Options{IPVersion: 6, RecordSize: 28})
	if err != nil {
		t.Fatal(err)
	}
	_, ipnet, err := net.ParseCIDR(pfx)
	if err != nil {
		t.Fatal(err)
	}
	if err := tree.Insert(ipnet, rec); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tree.WriteTo(f); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
