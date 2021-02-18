// Original: https://github.com/ethereum/research/blob/master/kzg_data_availability/kzg_proofs.py

// +build !bignum_pure,!bignum_hol256

package kzg

import "github.com/protolambda/go-kzg/bls"

// Compute KZG proof for polynomial in coefficient form at positions x * w^y where w is
// an n-th root of unity (this is the proof for one data availability sample, which consists
// of several polynomial evaluations)
func (ks *KZGSettings) ComputeProofMulti(poly []bls.Big, x uint64, n uint64) *bls.G1 {
	// divisor = [-pow(x, n, MODULUS)] + [0] * (n - 1) + [1]
	divisor := make([]bls.Big, n+1, n+1)
	var xBig bls.Big
	bls.AsBig(&xBig, x)
	// TODO: inefficient, could use squaring, or maybe BLS lib offers a power method?
	// TODO: for small ranges, maybe compute pow(x, n, mod) in uint64?
	var xPowN, tmp bls.Big
	for i := uint64(0); i < n; i++ {
		bls.MulModBig(&tmp, &xPowN, &xBig)
		bls.CopyBigNum(&xPowN, &tmp)
	}

	// -pow(x, n, MODULUS)
	bls.SubModBig(&divisor[0], &bls.ZERO, &xPowN)
	// [0] * (n - 1)
	for i := uint64(1); i < n; i++ {
		bls.CopyBigNum(&divisor[i], &bls.ZERO)
	}
	// 1
	bls.CopyBigNum(&divisor[n], &bls.ONE)

	// quot = poly / divisor
	quotientPolynomial := polyLongDiv(poly, divisor[:])
	//for i := 0; i < len(quotientPolynomial); i++ {
	//	fmt.Printf("quot poly %d: %s\n", i, BigStr(&quotientPolynomial[i]))
	//}

	// evaluate quotient poly at shared secret, in G1
	return bls.LinCombG1(ks.secretG1[:len(quotientPolynomial)], quotientPolynomial)
}

// Check a proof for a KZG commitment for an evaluation f(x w^i) = y_i
// The ys must have a power of 2 length
func (ks *KZGSettings) CheckProofMulti(commitment *bls.G1, proof *bls.G1, x *bls.Big, ys []bls.Big) bool {
	// Interpolate at a coset. Note because it is a coset, not the subgroup, we have to multiply the
	// polynomial coefficients by x^i
	interpolationPoly, err := ks.FFT(ys, true)
	if err != nil {
		panic("ys is bad, cannot compute FFT")
	}
	// TODO: can probably be optimized
	// apply div(c, pow(x, i, MODULUS)) to every coeff c in interpolationPoly
	// x^0 at first, then up to x^n
	var xPow bls.Big
	bls.CopyBigNum(&xPow, &bls.ONE)
	var tmp, tmp2 bls.Big
	for i := 0; i < len(interpolationPoly); i++ {
		bls.InvModBig(&tmp, &xPow)
		bls.MulModBig(&tmp2, &interpolationPoly[i], &tmp)
		bls.CopyBigNum(&interpolationPoly[i], &tmp2)
		bls.MulModBig(&tmp, &xPow, x)
		bls.CopyBigNum(&xPow, &tmp)
	}
	// [x^n]_2
	var xn2 bls.G2
	bls.MulG2(&xn2, &bls.GenG2, &xPow)
	// [s^n - x^n]_2
	var xnMinusYn bls.G2
	bls.SubG2(&xnMinusYn, &ks.secretG2[len(ys)], &xn2)

	// [interpolation_polynomial(s)]_1
	is1 := bls.LinCombG1(ks.secretG1[:len(interpolationPoly)], interpolationPoly)
	// [commitment - interpolation_polynomial(s)]_1 = [commit]_1 - [interpolation_polynomial(s)]_1
	var commitMinusInterpolation bls.G1
	bls.SubG1(&commitMinusInterpolation, commitment, is1)

	// Verify the pairing equation
	//
	// e([commitment - interpolation_polynomial(s)], [1]) = e([proof],  [s^n - x^n])
	//    equivalent to
	// e([commitment - interpolation_polynomial]^(-1), [1]) * e([proof],  [s^n - x^n]) = 1_T
	//

	return bls.PairingsVerify(&commitMinusInterpolation, &bls.GenG2, proof, &xnMinusYn)
}
