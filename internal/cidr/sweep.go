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

type sweepEvent struct {
	pos netip.Addr
	end bool
	idx int
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
	evs := make([]sweepEvent, 0, len(blocks)*2)
	for i, b := range blocks {
		evs = append(evs, sweepEvent{pos: b.start, end: false, idx: i})
		evs = append(evs, sweepEvent{pos: b.end, end: true, idx: i})
	}
	slices.SortFunc(evs, func(a, b sweepEvent) int {
		if c := addrCmp(a.pos, b.pos); c != 0 {
			return c
		}
		if a.end != b.end {
			if a.end {
				return -1
			}
			return 1
		}
		return 0
	})

	active := make(map[int]struct{}, 8)
	var last netip.Addr
	haveLast := false
	var out []mergedSpan

	i := 0
	for i < len(evs) {
		pos := evs[i].pos
		if haveLast && addrCmp(last, pos) != 0 && len(active) > 0 {
			var dst schema.Record
			winners := map[string]string{}
			for idx := range active {
				b := blocks[idx]
				merge.Merge(&dst, b.rec, b.source, p, winners)
			}
			if !dst.IsEmpty() {
				out = append(out, mergedSpan{start: last, end: pos, rec: dst})
			}
		}
		for i < len(evs) && addrCmp(evs[i].pos, pos) == 0 {
			if evs[i].end {
				delete(active, evs[i].idx)
			} else {
				active[evs[i].idx] = struct{}{}
			}
			i++
		}
		last = pos
		haveLast = true
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
