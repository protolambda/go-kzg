// Original: https://github.com/ethereum/research/blob/master/kzg_data_availability/kzg_proofs.py

//go:build !bignum_pure && !bignum_hol256
// +build !bignum_pure,!bignum_hol256

package kzg

import "github.com/protolambda/go-kzg/bls"

// KZG commitment to polynomial in evaluation form, i.e. eval = FFT(coeffs).
// The eval length must match the prepared KZG settings width.
func CommitToEvalPoly(secretG1IFFT []bls.G1Point, eval []bls.Fr) *bls.G1Point {
	return bls.LinCombG1(secretG1IFFT, eval)
}

// KZG commitment to polynomial in coefficient form
func (ks *KZGSettings) CommitToPoly(coeffs []bls.Fr) *bls.G1Point {
	return bls.LinCombG1(ks.SecretG1[:len(coeffs)], coeffs)
}

// KZG commitment to polynomial in coefficient form, unoptimized version
func (ks *KZGSettings) CommitToPolyUnoptimized(coeffs []bls.Fr) *bls.G1Point {
	// Do so by computing the linear combination with the shared secret.
	var out bls.G1Point
	bls.ClearG1(&out)
	var tmp, tmp2 bls.G1Point
	for i := 0; i < len(coeffs); i++ {
		bls.MulG1(&tmp, &ks.SecretG1[i], &coeffs[i])
		bls.AddG1(&tmp2, &out, &tmp)
		bls.CopyG1(&out, &tmp2)
	}
	return &out
}

// Compute KZG proof for polynomial in coefficient form at position x
func (ks *KZGSettings) ComputeProofSingle(poly []bls.Fr, x uint64) *bls.G1Point {
	// divisor = [-x, 1]
	divisor := [2]bls.Fr{}
	var tmp bls.Fr
	bls.AsFr(&tmp, x)
	bls.SubModFr(&divisor[0], &bls.ZERO, &tmp)
	bls.CopyFr(&divisor[1], &bls.ONE)
	//for i := 0; i < 2; i++ {
	//	fmt.Printf("div poly %d: %s\n", i, FrStr(&divisor[i]))
	//}
	// quot = poly / divisor
	quotientPolynomial := polyLongDiv(poly, divisor[:])
	//for i := 0; i < len(quotientPolynomial); i++ {
	//	fmt.Printf("quot poly %d: %s\n", i, FrStr(&quotientPolynomial[i]))
	//}

	// evaluate quotient poly at shared secret, in G1
	return bls.LinCombG1(ks.SecretG1[:len(quotientPolynomial)], quotientPolynomial)
}

// Check a proof for a KZG commitment for an evaluation f(x) = y
func (ks *KZGSettings) CheckProofSingle(commitment *bls.G1Point, proof *bls.G1Point, x *bls.Fr, y *bls.Fr) bool {
	// Verify the pairing equation
	var xG2 bls.G2Point
	bls.MulG2(&xG2, &bls.GenG2, x)
	var sMinuxX bls.G2Point
	bls.SubG2(&sMinuxX, &ks.SecretG2[1], &xG2)
	var yG1 bls.G1Point
	bls.MulG1(&yG1, &bls.GenG1, y)
	var commitmentMinusY bls.G1Point
	bls.SubG1(&commitmentMinusY, commitment, &yG1)

	// This trick may be applied in the BLS-lib specific code:
	//
	// e([commitment - y], [1]) = e([proof],  [s - x])
	//    equivalent to
	// e([commitment - y]^(-1), [1]) * e([proof],  [s - x]) = 1_T
	//
	return bls.PairingsVerify(&commitmentMinusY, &bls.GenG2, proof, &sMinuxX)
}
