package kate

// Kate commitment to polynomial in coefficient form
func (ks *KateSettings) CommitToPoly(poly []Big) *G1 {
	// TODO
	return nil
}

// Compute Kate proof for polynomial in coefficient form at position x
func (ks *KateSettings) ComputeProofSingle(poly []Big, x uint) *G1 {
	// TODO
	return nil
}

// Check a proof for a Kate commitment for an evaluation f(x) = y
func (ks *KateSettings) CheckProofSingle(commitment *G1, proof *G1, x uint, y *Big) bool {
	// TODO
	return false
}
