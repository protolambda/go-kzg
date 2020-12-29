// Original: https://github.com/ethereum/research/blob/master/kzg_data_availability/kzg_proofs.py

package kate

// Kate commitment to polynomial in coefficient form
func (ks *KateSettings) CommitToPoly(coeffs []Big) *G1 {
	return LinCombG1(ks.secretG1[:len(coeffs)], coeffs)
}

// Kate commitment to polynomial in coefficient form, unoptimized version
func (ks *KateSettings) CommitToPolyUnoptimized(coeffs []Big) *G1 {
	// Do so by computing the linear combination with the shared secret.
	var out G1
	ClearG1(&out)
	var tmp, tmp2 G1
	for i := 0; i < len(coeffs); i++ {
		mulG1(&tmp, &ks.secretG1[i], &coeffs[i])
		addG1(&tmp2, &out, &tmp)
		CopyG1(&out, &tmp2)
	}
	return &out
}

// Compute Kate proof for polynomial in coefficient form at position x
func (ks *KateSettings) ComputeProofSingle(poly []Big, x uint64) *G1 {
	// divisor = [-x, 1]
	divisor := [2]Big{}
	var tmp Big
	asBig(&tmp, x)
	subModBig(&divisor[0], &ZERO, &tmp)
	CopyBigNum(&divisor[1], &ONE)
	//for i := 0; i < 2; i++ {
	//	fmt.Printf("div poly %d: %s\n", i, bigStr(&divisor[i]))
	//}
	// quot = poly / divisor
	quotientPolynomial := polyLongDiv(poly, divisor[:])
	//for i := 0; i < len(quotientPolynomial); i++ {
	//	fmt.Printf("quot poly %d: %s\n", i, bigStr(&quotientPolynomial[i]))
	//}

	// evaluate quotient poly at shared secret, in G1
	return LinCombG1(ks.secretG1[:len(quotientPolynomial)], quotientPolynomial)
}

// Check a proof for a Kate commitment for an evaluation f(x) = y
func (ks *KateSettings) CheckProofSingle(commitment *G1, proof *G1, x *Big, y *Big) bool {
	// Verify the pairing equation
	var xG2 G2
	mulG2(&xG2, &genG2, x)
	var sMinuxX G2
	subG2(&sMinuxX, &ks.secretG2[1], &xG2)
	var yG1 G1
	mulG1(&yG1, &genG1, y)
	var commitmentMinusY G1
	subG1(&commitmentMinusY, commitment, &yG1)

	// This trick may be applied in the BLS-lib specific code:
	//
	// e([commitment - y], [1]) = e([proof],  [s - x])
	//    equivalent to
	// e([commitment - y]^(-1), [1]) * e([proof],  [s - x]) = 1_T
	//
	return PairingsVerify(&commitmentMinusY, &genG2, proof, &sMinuxX)
}
