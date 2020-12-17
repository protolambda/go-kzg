package kate

// Compute Kate proof for polynomial in coefficient form at positions x * w^y where w is
// an n-th root of unity (this is the proof for one data availability sample, which consists
// of several polynomial evaluations)
func (ks *KateSettings) ComputeProofMulti(poly []Big, x uint, n uint) *G1 {
	// TODO
	return nil
}

// Check a proof for a Kate commitment for an evaluation f(x w^i) = y_i
func (ks *KateSettings) CheckProofMulti(commitment *G1, proof *G1, x uint, ys []Big) bool {
	// TODO
	return false
}
