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

// Sanity check the mod div function, some libraries do regular integer div.
func TestDivModFr(t *testing.T) {
	var aVal Fr
	SetFr(&aVal, "26444158170683616486493062254748234829545368615823006596610545696213139843950")
	var bVal Fr
	SetFr(&bVal, "44429412392042760961177795624245903017634520927909329890010277843232676752272")
	var div Fr
	DivModFr(&div, &aVal, &bVal)

	var invB Fr
	InvModFr(&invB, &bVal)
	var mulInv Fr
	MulModFr(&mulInv, &aVal, &invB)

	if !EqualFr(&div, &mulInv) {
		t.Fatalf("div num not equal to mul inv: %s, %s", FrStr(&div), FrStr(&mulInv))
	}
}

func TestValidFr(t *testing.T) {
	data := FrTo32(&MODULUS_MINUS1)
	if !ValidFr(data) {
		t.Fatal("expected mod-1 to be valid")
	}
	var tmp [32]byte
	for i := 0; i < 32; i++ {
		if data[i] == 0xff {
			continue
		}
		tmp = data
		tmp[i] += 1
		if ValidFr(tmp) {
			t.Fatal("expected anything larger than mod-1 to be invalid")
		}
	}
	v := RandomFr()
	data = FrTo32(v)
	if !ValidFr(data) {
		t.Fatalf("expected generated Fr %s to be valid", v)
	}
	data = [32]byte{}
	if !ValidFr(data) {
		t.Fatal("expected zero to be valid")
	}
}
