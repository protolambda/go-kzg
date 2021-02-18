package bls

import "testing"

// These are sanity tests, to see if whatever bignum library that is being
// used actually handles dst/arg overlaps well.

func TestInplaceAdd(t *testing.T) {
	aVal := RandomBig()
	bVal := RandomBig()
	aPlusB := new(Big)
	AddModBig(aPlusB, aVal, bVal)
	twoA := new(Big)
	MulModBig(twoA, aVal, &TWO)

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
		AddModBig(a, a, b)
		return EqualBig(a, aPlusB)
	})
	check("dst equals rhs", func(a *Big, b *Big) bool {
		AddModBig(b, a, b)
		return EqualBig(b, aPlusB)
	})
	check("dst equals lhs and rhs", func(a *Big, b *Big) bool {
		AddModBig(a, a, a)
		return EqualBig(a, twoA)
	})
}

func TestInplaceMul(t *testing.T) {
	aVal := RandomBig()
	bVal := RandomBig()
	aMulB := new(Big)
	MulModBig(aMulB, aVal, bVal)
	squareA := new(Big)
	MulModBig(squareA, aVal, aVal)

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
		MulModBig(a, a, b)
		return EqualBig(a, aMulB)
	})
	check("dst equals rhs", func(a *Big, b *Big) bool {
		MulModBig(b, a, b)
		return EqualBig(b, aMulB)
	})
	check("dst equals lhs and rhs", func(a *Big, b *Big) bool {
		MulModBig(a, a, a)
		return EqualBig(a, squareA)
	})
}
