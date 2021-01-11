// +build !bignum_pure,!bignum_hol256

package kate

import (
	"fmt"
	"testing"
)

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
	cosetScale := uint8(3)
	coset := make([]Big, 1<<cosetScale, 1<<cosetScale)
	s1, s2 = generateSetup("1927409816240961209460912649124", 8+1)
	ks = NewKateSettings(NewFFTSettings(cosetScale), s1, s2)
	for i := 0; i < len(coset); i++ {
		fmt.Printf("rootz %d: %s\n", i, bigStr(&ks.expandedRootsOfUnity[i]))
		mulModBig(&coset[i], &xBig, &ks.expandedRootsOfUnity[i])
		fmt.Printf("coset %d: %s\n", i, bigStr(&coset[i]))
	}
	ys := make([]Big, len(coset), len(coset))
	for i := 0; i < len(coset); i++ {
		EvalPolyAt(&ys[i], polynomial, &coset[i])
		fmt.Printf("ys %d: %s\n", i, bigStr(&ys[i]))
	}

	proof := ks.ComputeProofMulti(polynomial, x, uint64(len(coset)))
	fmt.Printf("proof: %s\n", strG1(proof))
	if !ks.CheckProofMulti(commitment, proof, &xBig, ys) {
		t.Fatal("could not verify proof")
	}
}
