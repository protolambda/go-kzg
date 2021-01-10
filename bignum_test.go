package kate

import "testing"

// These are sanity tests, to see if whatever bignum library that is being
// used actually handles dst/arg overlaps well.

func TestInplaceAdd(t *testing.T) {
	aVal := randomBig()
	bVal := randomBig()
	aPlusB := new(Big)
	addModBig(aPlusB, aVal, bVal)
	twoA := new(Big)
	mulModBig(twoA, aVal, &TWO)

	check := func(name string, fn func(a, b *Big) bool) {
		t.Run(name, func(t *testing.T) {
			var a, b Big
			CopyBigNum(&a, aVal)
			CopyBigNum(&b, bVal)
			if !fn(&a, &b) {
				t.Error("fail")
			}
		})
	}
	check("dst equals lhs", func(a *Big, b *Big) bool {
		addModBig(a, a, b)
		return equalBig(a, aPlusB)
	})
	check("dst equals rhs", func(a *Big, b *Big) bool {
		addModBig(b, a, b)
		return equalBig(b, aPlusB)
	})
	check("dst equals lhs and rhs", func(a *Big, b *Big) bool {
		addModBig(a, a, a)
		return equalBig(a, twoA)
	})
}

func TestInplaceMul(t *testing.T) {
	aVal := randomBig()
	bVal := randomBig()
	aMulB := new(Big)
	mulModBig(aMulB, aVal, bVal)
	squareA := new(Big)
	mulModBig(squareA, aVal, aVal)

	check := func(name string, fn func(a, b *Big) bool) {
		t.Run(name, func(t *testing.T) {
			var a, b Big
			CopyBigNum(&a, aVal)
			CopyBigNum(&b, bVal)
			if !fn(&a, &b) {
				t.Error("fail")
			}
		})
	}
	check("dst equals lhs", func(a *Big, b *Big) bool {
		mulModBig(a, a, b)
		return equalBig(a, aMulB)
	})
	check("dst equals rhs", func(a *Big, b *Big) bool {
		mulModBig(b, a, b)
		return equalBig(b, aMulB)
	})
	check("dst equals lhs and rhs", func(a *Big, b *Big) bool {
		mulModBig(a, a, a)
		return equalBig(a, squareA)
	})
}
