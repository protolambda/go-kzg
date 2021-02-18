package bls

import "testing"

// These are sanity tests, to see if whatever bignum library that is being
// used actually handles dst/arg overlaps well.

func TestInplaceAdd(t *testing.T) {
	aVal := RandomFr()
	bVal := RandomFr()
	aPlusB := new(Fr)
	AddModFr(aPlusB, aVal, bVal)
	twoA := new(Fr)
	MulModFr(twoA, aVal, &TWO)

	check := func(name string, fn func(a, b *Fr) bool) {
		t.Run(name, func(t *testing.T) {
			var a, b Fr
			CopyFr(&a, aVal)
			CopyFr(&b, bVal)
			if !fn(&a, &b) {
				t.Error("fail")
			}
		})
	}
	check("dst equals lhs", func(a *Fr, b *Fr) bool {
		AddModFr(a, a, b)
		return EqualFr(a, aPlusB)
	})
	check("dst equals rhs", func(a *Fr, b *Fr) bool {
		AddModFr(b, a, b)
		return EqualFr(b, aPlusB)
	})
	check("dst equals lhs and rhs", func(a *Fr, b *Fr) bool {
		AddModFr(a, a, a)
		return EqualFr(a, twoA)
	})
}

func TestInplaceMul(t *testing.T) {
	aVal := RandomFr()
	bVal := RandomFr()
	aMulB := new(Fr)
	MulModFr(aMulB, aVal, bVal)
	squareA := new(Fr)
	MulModFr(squareA, aVal, aVal)

	check := func(name string, fn func(a, b *Fr) bool) {
		t.Run(name, func(t *testing.T) {
			var a, b Fr
			CopyFr(&a, aVal)
			CopyFr(&b, bVal)
			if !fn(&a, &b) {
				t.Error("fail")
			}
		})
	}
	check("dst equals lhs", func(a *Fr, b *Fr) bool {
		MulModFr(a, a, b)
		return EqualFr(a, aMulB)
	})
	check("dst equals rhs", func(a *Fr, b *Fr) bool {
		MulModFr(b, a, b)
		return EqualFr(b, aMulB)
	})
	check("dst equals lhs and rhs", func(a *Fr, b *Fr) bool {
		MulModFr(a, a, a)
		return EqualFr(a, squareA)
	})
}
