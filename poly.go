package kzg

// invert the divisor, then multiply
func polyFactorDiv(dst *Big, a *Big, b *Big) {
	// TODO: use divmod instead.
	var tmp Big
	invModBig(&tmp, b)
	mulModBig(dst, &tmp, a)
}

// Long polynomial division for two polynomials in coefficient form
func polyLongDiv(dividend []Big, divisor []Big) []Big {
	a := make([]Big, len(dividend), len(dividend))
	for i := 0; i < len(a); i++ {
		CopyBigNum(&a[i], &dividend[i])
	}
	aPos := len(a) - 1
	bPos := len(divisor) - 1
	diff := aPos - bPos
	out := make([]Big, diff+1, diff+1)
	for diff >= 0 {
		quot := &out[diff]
		polyFactorDiv(quot, &a[aPos], &divisor[bPos])
		var tmp, tmp2 Big
		for i := bPos; i >= 0; i-- {
			// In steps: a[diff + i] -= b[i] * quot
			// tmp =  b[i] * quot
			mulModBig(&tmp, quot, &divisor[i])
			// tmp2 = a[diff + i] - tmp
			subModBig(&tmp2, &a[diff+i], &tmp)
			// a[diff + i] = tmp2
			CopyBigNum(&a[diff+i], &tmp2)
		}
		aPos -= 1
		diff -= 1
	}
	return out
}
