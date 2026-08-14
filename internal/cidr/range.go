package cidr

import (
	"net/netip"

	"github.com/shafqat-a/iplegence/internal/schema"
	"go4.org/netipx"
)

type Block struct {
	Prefix netip.Prefix
	Rec    schema.Record
	Source string
}

type Row struct {
	Prefix netip.Prefix
	Rec    schema.Record
}

// endMaxSentinel is an invalid Addr used as exclusive end after the last
// address of a family (255.255.255.255 or ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff).
// netip.Addr.Next() on those addresses returns the zero Addr; we keep that
// sentinel instead of overflowing.
var endMaxSentinel netip.Addr

func PrefixToHalfOpen(p netip.Prefix) (start, endExclusive netip.Addr, ok bool) {
	if !p.IsValid() {
		return netip.Addr{}, netip.Addr{}, false
	}
	p = p.Masked()
	addr := p.Addr()
	if addr.Is4In6() {
		return netip.Addr{}, netip.Addr{}, false
	}
	r := netipx.RangeOfPrefix(p)
	start = r.From()
	last := r.To()
	endExclusive = last.Next()
	if !endExclusive.IsValid() {
		endExclusive = endMaxSentinel
	}
	return start, endExclusive, true
}

func lastInclusive(start, endExclusive netip.Addr) (netip.Addr, bool) {
	if !start.IsValid() {
		return netip.Addr{}, false
	}
	if !endExclusive.IsValid() {
		if start.Is4() {
			return netip.AddrFrom4([4]byte{255, 255, 255, 255}), true
		}
		var all [16]byte
		for i := range all {
			all[i] = 0xff
		}
		return netip.AddrFrom16(all), true
	}
	prev, ok := prevAddr(endExclusive)
	return prev, ok
}

func prevAddr(a netip.Addr) (netip.Addr, bool) {
	if !a.IsValid() {
		return netip.Addr{}, false
	}
	if a.Is4() {
		n := a.As4()
		for i := 3; i >= 0; i-- {
			if n[i] > 0 {
				n[i]--
				return netip.AddrFrom4(n), true
			}
			n[i] = 255
		}
		return netip.Addr{}, false
	}
	n := a.As16()
	for i := 15; i >= 0; i-- {
		if n[i] > 0 {
			n[i]--
			return netip.AddrFrom16(n), true
		}
		n[i] = 255
	}
	return netip.Addr{}, false
}

func addrLess(a, b netip.Addr) bool {
	if !a.IsValid() {
		return false
	}
	if !b.IsValid() {
		return true
	}
	return a.Less(b)
}

func addrCmp(a, b netip.Addr) int {
	aValid, bValid := a.IsValid(), b.IsValid()
	if !aValid && !bValid {
		return 0
	}
	if !aValid {
		return 1
	}
	if !bValid {
		return -1
	}
	if a.Less(b) {
		return -1
	}
	if b.Less(a) {
		return 1
	}
	return 0
}
