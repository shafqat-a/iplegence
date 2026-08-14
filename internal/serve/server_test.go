package serve_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/shafqat-a/iplegence/internal/cidr"
	"github.com/shafqat-a/iplegence/internal/mmdb"
	"github.com/shafqat-a/iplegence/internal/schema"
	"github.com/shafqat-a/iplegence/internal/serve"
)

func TestLookupAndHealth(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "Superior-IP.mmdb")
	_, err := mmdb.Write([]cidr.Row{{
		Prefix: netip.MustParsePrefix("8.8.8.0/24"),
		Rec: schema.Record{
			Country: schema.Country{ISOCode: "US"},
			ASN:     schema.ASN{Number: 15169},
		},
	}}, dest, 10<<20)
	if err != nil {
		t.Fatal(err)
	}
	srv, err := serve.New("127.0.0.1:0", dest)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	h := srv.Handler()

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("health %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/lookup/8.8.8.8", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("lookup %d %s", rr.Code, rr.Body.String())
	}
	var rec schema.LookupRecord
	if err := json.Unmarshal(rr.Body.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	if rec.Country.ISOCode != "US" || rec.ASN.Number != 15169 {
		t.Fatalf("got %+v", rec)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/not-an-ip", nil))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("bad ip %d", rr.Code)
	}
}
