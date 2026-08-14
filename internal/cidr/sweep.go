package cidr

import (
	"fmt"
	"net/netip"
	"slices"

	"github.com/shafqat-a/iplegence/internal/merge"
	"github.com/shafqat-a/iplegence/internal/schema"
	"go4.org/netipx"
)

type intervalBlock struct {
	start, end netip.Addr
	rec        schema.Record
	source     string
}

type mergedSpan struct {
	start, end netip.Addr
	rec        schema.Record
}

func MergeBlocks(blocks []Block, p merge.Priority) ([]Row, error) {
	var v4, v6 []intervalBlock
	for _, b := range blocks {
		start, end, ok := PrefixToHalfOpen(b.Prefix)
		if !ok {
			return nil, fmt.Errorf("invalid or IPv4-mapped prefix %s", b.Prefix)
		}
		ib := intervalBlock{start: start, end: end, rec: b.Rec, source: b.Source}
		if start.Is4() {
			v4 = append(v4, ib)
		} else {
			v6 = append(v6, ib)
		}
	}
	spans := append(sweepFamily(v4, p), sweepFamily(v6, p)...)
	spans = collapseSpans(spans)
	return spansToRows(spans)
}

func sweepFamily(blocks []intervalBlock, p merge.Priority) []mergedSpan {
	if len(blocks) == 0 {
		return nil
	}
	points := make([]netip.Addr, 0, len(blocks)*2)
	for _, b := range blocks {
		points = append(points, b.start, b.end)
	}
	slices.SortFunc(points, addrCmp)
	points = slices.CompactFunc(points, func(a, b netip.Addr) bool {
		return addrCmp(a, b) == 0
	})

	var out []mergedSpan
	for i := 0; i+1 < len(points); i++ {
		lo, hi := points[i], points[i+1]
		if addrCmp(lo, hi) == 0 {
			continue
		}
		var dst schema.Record
		winners := map[string]string{}
		any := false
		for _, b := range blocks {
			if addrCmp(b.start, lo) <= 0 && addrCmp(hi, b.end) <= 0 {
				merge.Merge(&dst, b.rec, b.source, p, winners)
				any = true
			}
		}
		if !any || dst.IsEmpty() {
			continue
		}
		out = append(out, mergedSpan{start: lo, end: hi, rec: dst})
	}
	return out
}

func spansToRows(spans []mergedSpan) ([]Row, error) {
	var rows []Row
	for _, s := range spans {
		last, ok := lastInclusive(s.start, s.end)
		if !ok {
			return nil, fmt.Errorf("empty span starting at %s", s.start)
		}
		ir := netipx.IPRangeFrom(s.start, last)
		if !ir.IsValid() {
			return nil, fmt.Errorf("invalid range %s-%s", s.start, last)
		}
		for _, pfx := range ir.Prefixes() {
			rows = append(rows, Row{Prefix: pfx, Rec: s.rec})
		}
	}
	return rows, nil
}
