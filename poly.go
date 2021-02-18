package kzg

import "github.com/protolambda/go-kzg/bls"

// invert the divisor, then multiply
func polyFactorDiv(dst *bls.Big, a *bls.Big, b *bls.Big) {
	// TODO: use divmod instead.
	var tmp bls.Big
	bls.InvModBig(&tmp, b)
	bls.MulModBig(dst, &tmp, a)
}

// Long polynomial division for two polynomials in coefficient form
func polyLongDiv(dividend []bls.Big, divisor []bls.Big) []bls.Big {
	a := make([]bls.Big, len(dividend), len(dividend))
	for i := 0; i < len(a); i++ {
		bls.CopyBigNum(&a[i], &dividend[i])
	}
	aPos := len(a) - 1
	bPos := len(divisor) - 1
	diff := aPos - bPos
	out := make([]bls.Big, diff+1, diff+1)
	for diff >= 0 {
		quot := &out[diff]
		polyFactorDiv(quot, &a[aPos], &divisor[bPos])
		var tmp, tmp2 bls.Big
		for i := bPos; i >= 0; i-- {
			// In steps: a[diff + i] -= b[i] * quot
			// tmp =  b[i] * quot
			bls.MulModBig(&tmp, quot, &divisor[i])
			// tmp2 = a[diff + i] - tmp
			bls.SubModBig(&tmp2, &a[diff+i], &tmp)
			// a[diff + i] = tmp2
			bls.CopyBigNum(&a[diff+i], &tmp2)
		}
		aPos -= 1
		diff -= 1
	}
	return out
}
