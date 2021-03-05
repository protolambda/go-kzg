// +build !bignum_pure,!bignum_hol256

package kzg

import (
	"github.com/protolambda/go-kzg/bls"
	"testing"
)

func TestKZGSettings_DAUsingFK20Multi(t *testing.T) {
	fs := NewFFTSettings(4 + 5 + 1)
	chunkLen := uint64(16)
	chunkCount := uint64(32)
	n := chunkLen * chunkCount
	s1, s2 := GenerateTestingSetup("1927409816240961209460912649124", chunkLen*chunkCount*2)
	ks := NewKZGSettings(fs, s1, s2)
	fk := NewFK20MultiSettings(ks, n*2, chunkLen)

	// replicate same polynomial as in python test
	polynomial := make([]bls.Fr, n, n)
	var tmp134 bls.Fr
	bls.AsFr(&tmp134, 134)
	for i := uint64(0); i < chunkCount; i++ {
		// Note: different contents from older python test, make each section different,
		// to cover toeplitz coefficient edge cases better.
		for j, v := range []uint64{1, 2, 3, 4 + i, 7, 8 + i*i, 9, 10, 13, 14, 1, 15, 0, 1000, 0, 33} {
			bls.AsFr(&polynomial[i*chunkLen+uint64(j)], v)
		}
		bls.SubModFr(&polynomial[i*chunkLen+12], &bls.ZERO, &bls.ONE) // "MODULUS - 1"
		bls.SubModFr(&polynomial[i*chunkLen+14], &bls.ZERO, &tmp134)  // "MODULUS - 134"
	}

	commitment := ks.CommitToPoly(polynomial)
	t.Log("commitment\n", bls.StrG1(commitment))

	allProofs := fk.DAUsingFK20Multi(polynomial)
	t.Log("All KZG proofs computed for data availability (supersampled by factor 2)")
	for i := 0; i < len(allProofs); i++ {
		t.Logf("%d: %s", i, bls.StrG1(&allProofs[i]))
	}

	// We have the data in polynomial form already,
	// no need to use the DAS FFT (which extends data directly, not coeffs).
	extendedCoeffs := make([]bls.Fr, n*2, n*2)
	for i := uint64(0); i < n; i++ {
		bls.CopyFr(&extendedCoeffs[i], &polynomial[i])
	}
	for i := n; i < n*2; i++ {
		bls.CopyFr(&extendedCoeffs[i], &bls.ZERO)
	}
	extendedData, err := ks.FFT(extendedCoeffs, false)
	if err != nil {
		t.Fatal(err)
	}
	reverseBitOrderFr(extendedData)
	debugFrs("extended_data", extendedData)

	n2 := n * 2
	domainStride := fk.MaxWidth / n2
	for pos := uint64(0); pos < 2*chunkCount; pos++ {
		domainPos := reverseBitsLimited(uint32(2*chunkCount), uint32(pos))
		var x bls.Fr
		bls.CopyFr(&x, &ks.ExpandedRootsOfUnity[uint64(domainPos)*domainStride])
		ys := extendedData[chunkLen*pos : chunkLen*(pos+1)]
		// ys, but constructed by evaluating the polynomial in the sub-domain range
		ys2 := make([]bls.Fr, chunkLen, chunkLen)
		// don't recompute the subgroup domain, just select it from the bigger domain by applying a stride
		stride := ks.MaxWidth / chunkLen
		coset := make([]bls.Fr, chunkLen, chunkLen)
		for i := uint64(0); i < chunkLen; i++ {
			var z bls.Fr // a value of the coset list
			bls.MulModFr(&z, &x, &ks.ExpandedRootsOfUnity[i*stride])
			bls.CopyFr(&coset[i], &z)
			bls.EvalPolyAt(&ys2[i], polynomial, &z)
		}
		// permanently change order of ys values
		reverseBitOrderFr(ys)
		for i := 0; i < len(ys); i++ {
			if !bls.EqualFr(&ys[i], &ys2[i]) {
				t.Fatal("failed to reproduce matching y values for subgroup")
			}
		}

		proof := &allProofs[pos]
		if !ks.CheckProofMulti(commitment, proof, &x, ys) {
			t.Fatal("could not verify proof")
		}
		t.Logf("Data availability check %d passed", pos)
	}
}
