// Original: https://github.com/ethereum/research/blob/master/kzg_data_availability/kzg_proofs.py

// +build !bignum_pure,!bignum_hol256

package kate

// Compute Kate proof for polynomial in coefficient form at positions x * w^y where w is
// an n-th root of unity (this is the proof for one data availability sample, which consists
// of several polynomial evaluations)
func (ks *KateSettings) ComputeProofMulti(poly []Big, x uint64, n uint64) *G1 {
	// divisor = [-pow(x, n, MODULUS)] + [0] * (n - 1) + [1]
	divisor := make([]Big, n+1, n+1)
	var xBig Big
	asBig(&xBig, x)
	// TODO: inefficient, could use squaring, or maybe BLS lib offers a power method?
	// TODO: for small ranges, maybe compute pow(x, n, mod) in uint64?
	var xPowN, tmp Big
	for i := uint64(0); i < n; i++ {
		mulModBig(&tmp, &xPowN, &xBig)
		CopyBigNum(&xPowN, &tmp)
	}

	// -pow(x, n, MODULUS)
	subModBig(&divisor[0], &ZERO, &xPowN)
	// [0] * (n - 1)
	for i := uint64(1); i < n; i++ {
		CopyBigNum(&divisor[i], &ZERO)
	}
	// 1
	CopyBigNum(&divisor[n], &ONE)

	// quot = poly / divisor
	quotientPolynomial := polyLongDiv(poly, divisor[:])
	//for i := 0; i < len(quotientPolynomial); i++ {
	//	fmt.Printf("quot poly %d: %s\n", i, bigStr(&quotientPolynomial[i]))
	//}

	// evaluate quotient poly at shared secret, in G1
	return LinCombG1(ks.secretG1[:len(quotientPolynomial)], quotientPolynomial)
}

// Check a proof for a Kate commitment for an evaluation f(x w^i) = y_i
// The ys must have a power of 2 length
func (ks *KateSettings) CheckProofMulti(commitment *G1, proof *G1, x *Big, ys []Big) bool {
	// Interpolate at a coset. Note because it is a coset, not the subgroup, we have to multiply the
	// polynomial coefficients by x^i
	interpolationPoly, err := ks.FFT(ys, true)
	if err != nil {
		panic("ys is bad, cannot compute FFT")
	}
	// TODO: can probably be optimized
	// apply div(c, pow(x, i, MODULUS)) to every coeff c in interpolationPoly
	// x^0 at first, then up to x^n
	var xPow Big
	CopyBigNum(&xPow, &ONE)
	var tmp, tmp2 Big
	for i := 0; i < len(interpolationPoly); i++ {
		invModBig(&tmp, &xPow)
		mulModBig(&tmp2, &interpolationPoly[i], &tmp)
		CopyBigNum(&interpolationPoly[i], &tmp2)
		mulModBig(&tmp, &xPow, x)
		CopyBigNum(&xPow, &tmp)
	}
	// [x^n]_2
	var xn2 G2
	mulG2(&xn2, &genG2, norm(&xPow))
	// [s^n - x^n]_2
	var xnMinusYn G2
	subG2(&xnMinusYn, &ks.secretG2[len(ys)], &xn2)

	// [interpolation_polynomial(s)]_1
	is1 := LinCombG1(ks.secretG1[:len(interpolationPoly)], interpolationPoly)
	// [commitment - interpolation_polynomial(s)]_1 = [commit]_1 - [interpolation_polynomial(s)]_1
	var commitMinusInterpolation G1
	subG1(&commitMinusInterpolation, commitment, is1)

	// Verify the pairing equation
	//
	// e([commitment - interpolation_polynomial(s)], [1]) = e([proof],  [s^n - x^n])
	//    equivalent to
	// e([commitment - interpolation_polynomial]^(-1), [1]) * e([proof],  [s^n - x^n]) = 1_T
	//

	return PairingsVerify(&commitMinusInterpolation, &genG2, proof, &xnMinusYn)
}
