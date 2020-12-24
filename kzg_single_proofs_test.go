package kate

import "testing"

func TestKateSettings_CommitToPoly(t *testing.T) {
	// TODO
}

func TestKateSettings_ComputeProofSingle(t *testing.T) {
	// TODO
}

func TestKateSettings_CheckProofSingle(t *testing.T) {
	fs := NewFFTSettings(4)
	s1, s2 := generateSetup("1927409816240961209460912649124", 16+1)
	ks := NewKateSettings(fs, s1, s2)
	for i := 0; i < len(ks.secretG1); i++ {
		t.Logf("secret g1 %d: %s", i, strG1(&ks.secretG1[i]))
	}

	polynomial := testPoly(1, 2, 3, 4, 7, 7, 7, 7, 13, 13, 13, 13, 13, 13, 13, 13)
	for i := 0; i < len(polynomial); i++ {
		t.Logf("poly coeff %d: %s", i, bigStr(&polynomial[i]))
	}

	commitment := ks.CommitToPoly(polynomial)
	t.Log("commitment\n", strG1(commitment))

	proof := ks.ComputeProofSingle(polynomial, 17)
	t.Log("proof\n", strG1(proof))

	var x Big
	asBig(&x, 17)
	var value Big
	EvalPolyAt(&value, polynomial, &x)
	t.Log("value\n", bigStr(&value))

	if !ks.CheckProofSingle(commitment, proof, &x, &value) {
		t.Fatal("could not verify proof")
	}
}

func testPoly(polynomial ...uint64) []Big {
	n := len(polynomial)
	polynomialBig := make([]Big, n, n)
	for i := 0; i < n; i++ {
		asBig(&polynomialBig[i], polynomial[i])
	}
	return polynomialBig
}

func generateSetup(secret string, n uint64) ([]G1, []G2) {
	var s Big
	bigNum(&s, secret)

	var sPow Big
	CopyBigNum(&sPow, &ONE)

	s1Out := make([]G1, n, n)
	s2Out := make([]G2, n, n)
	for i := uint64(0); i < n; i++ {
		mulG1(&s1Out[i], &genG1, &sPow)
		mulG2(&s2Out[i], &genG2, &sPow)
		var tmp Big
		CopyBigNum(&tmp, &sPow)
		mulModBig(&sPow, &tmp, &s)
	}
	return s1Out, s2Out
}
