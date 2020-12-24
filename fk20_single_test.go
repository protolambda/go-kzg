package kate

import "testing"

func TestKateSettings_DAUsingFK20(t *testing.T) {
	fs := NewFFTSettings(5)
	s1, s2 := generateSetup("1927409816240961209460912649124", 32+1)
	ks := NewKateSettings(fs, s1, s2)

	polynomial := testPoly(1, 2, 3, 4, 7, 7, 7, 7, 13, 13, 13, 13, 13, 13, 13, 13)

	commitment := ks.CommitToPoly(polynomial)
	t.Log("commitment\n", strG1(commitment))

	allProofs := ks.DAUsingFK20(polynomial)
	t.Log("All KZG proofs computed")
	for i := 0; i < len(allProofs); i++ {
		t.Logf("%d: %s", i, strG1(&allProofs[i]))
	}

	// Now check a random position
	pos := uint64(9)
	var posBig Big
	asBig(&posBig, pos)
	var x Big
	CopyBigNum(&x, &ks.expandedRootsOfUnity[pos])
	t.Log("x:\n", bigStr(&x))
	var y Big
	EvalPolyAt(&y, polynomial, &x)
	t.Log("y:\n", bigStr(&y))

	proof := &allProofs[reverseBitsLimited(uint32(2*16), uint32(pos))]

	if !ks.CheckProofSingle(commitment, proof, &x, &y) {
		t.Fatal("could not verify proof")
	}
}
