package engine

import (
	"strconv"
	"strings"
	"unicode"
)

// NaturalLess compares names like Windows Explorer for the common case where
// digit runs should sort numerically: 1, 2, 10 instead of 1, 10, 2.
func NaturalLess(a, b string) bool {
	ra, rb := []rune(strings.ToLower(a)), []rune(strings.ToLower(b))
	for i, j := 0, 0; i < len(ra) || j < len(rb); {
		if i >= len(ra) {
			return true
		}
		if j >= len(rb) {
			return false
		}

		if unicode.IsDigit(ra[i]) && unicode.IsDigit(rb[j]) {
			ii, jj := i, j
			for ii < len(ra) && unicode.IsDigit(ra[ii]) {
				ii++
			}
			for jj < len(rb) && unicode.IsDigit(rb[jj]) {
				jj++
			}
			as, bs := string(ra[i:ii]), string(rb[j:jj])
			an, _ := strconv.ParseUint(strings.TrimLeft(as, "0"), 10, 64)
			bn, _ := strconv.ParseUint(strings.TrimLeft(bs, "0"), 10, 64)
			if strings.Trim(as, "0") == "" {
				an = 0
			}
			if strings.Trim(bs, "0") == "" {
				bn = 0
			}
			if an != bn {
				return an < bn
			}
			if len(as) != len(bs) {
				return len(as) < len(bs)
			}
			i, j = ii, jj
			continue
		}

		if ra[i] != rb[j] {
			return ra[i] < rb[j]
		}
		i++
		j++
	}
	return strings.ToLower(a) < strings.ToLower(b)
}
