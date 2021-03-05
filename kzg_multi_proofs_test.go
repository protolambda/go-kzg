// +build !bignum_pure,!bignum_hol256

package kzg

import (
	"fmt"
	"github.com/protolambda/go-kzg/bls"
	"testing"
)

func TestKZGSettings_CheckProofMulti(t *testing.T) {
	fs := NewFFTSettings(4)
	s1, s2 := GenerateTestingSetup("1927409816240961209460912649124", 16+1)
	ks := NewKZGSettings(fs, s1, s2)
	for i := 0; i < len(ks.SecretG1); i++ {
		t.Logf("secret g1 %d: %s", i, bls.StrG1(&ks.SecretG1[i]))
	}

	polynomial := testPoly(1, 2, 3, 4, 7, 7, 7, 7, 13, 13, 13, 13, 13, 13, 13, 13)
	for i := 0; i < len(polynomial); i++ {
		t.Logf("poly coeff %d: %s", i, bls.FrStr(&polynomial[i]))
	}

	commitment := ks.CommitToPoly(polynomial)
	t.Log("commitment\n", bls.StrG1(commitment))

	x := uint64(5431)
	var xFr bls.Fr
	bls.AsFr(&xFr, x)
	cosetScale := uint8(3)
	coset := make([]bls.Fr, 1<<cosetScale, 1<<cosetScale)
	s1, s2 = GenerateTestingSetup("1927409816240961209460912649124", 8+1)
	ks = NewKZGSettings(NewFFTSettings(cosetScale), s1, s2)
	for i := 0; i < len(coset); i++ {
		fmt.Printf("rootz %d: %s\n", i, bls.FrStr(&ks.ExpandedRootsOfUnity[i]))
		bls.MulModFr(&coset[i], &xFr, &ks.ExpandedRootsOfUnity[i])
		fmt.Printf("coset %d: %s\n", i, bls.FrStr(&coset[i]))
	}
	ys := make([]bls.Fr, len(coset), len(coset))
	for i := 0; i < len(coset); i++ {
		bls.EvalPolyAt(&ys[i], polynomial, &coset[i])
		fmt.Printf("ys %d: %s\n", i, bls.FrStr(&ys[i]))
	}

	proof := ks.ComputeProofMulti(polynomial, x, uint64(len(coset)))
	fmt.Printf("proof: %s\n", bls.StrG1(proof))
	if !ks.CheckProofMulti(commitment, proof, &xFr, ys) {
		t.Fatal("could not verify proof")
	}
}
