package cidr

func collapseSpans(spans []mergedSpan) []mergedSpan {
	if len(spans) == 0 {
		return nil
	}
	out := []mergedSpan{spans[0]}
	for _, s := range spans[1:] {
		prev := &out[len(out)-1]
		sameFamily := prev.start.Is4() == s.start.Is4()
		if sameFamily && addrCmp(prev.end, s.start) == 0 && prev.rec.Equal(s.rec) {
			prev.end = s.end
			continue
		}
		out = append(out, s)
	}
	return out
}
