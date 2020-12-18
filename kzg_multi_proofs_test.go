package kate

import "testing"

func TestKateSettings_ComputeProofMulti(t *testing.T) {
	// TODO
}

func TestKateSettings_CheckProofMulti(t *testing.T) {
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

	x := uint64(5431)
	var xBig Big
	asBig(&xBig, x)
	cosetScale := 3
	coset := make([]Big, 1<<cosetScale, 1<<cosetScale)
	root := &scale2RootOfUnity[cosetScale]
	rootz := expandRootOfUnity(root)
	for i := 0; i < len(coset); i++ {
		mulModBig(&coset[i], &xBig, &rootz[i])
	}
	ys := make([]Big, len(coset), len(coset))
	for i := 0; i < len(coset); i++ {
		EvalPolyAt(&ys[i], polynomial, &coset[i])
	}

	proof := ks.ComputeProofMulti(polynomial, x, uint64(len(coset)))
	if !ks.CheckProofMulti(commitment, proof, x, ys) {
		// TODO; test failing
		t.Fatal("could not verify proof")
	}
}
