// +build !bignum_pure,!bignum_hol256

package kate

import "testing"

func TestKateSettings_DAUsingFK20Multi(t *testing.T) {
	fs := NewFFTSettings(4 + 5 + 1)
	chunkLen := uint64(16)
	chunkCount := uint64(32)
	n := chunkLen * chunkCount
	s1, s2 := generateSetup("1927409816240961209460912649124", chunkLen*chunkCount*2)
	ks := NewKateSettings(fs, s1, s2)
	fk := NewFK20MultiSettings(ks, n*2, chunkLen)

	// replicate same polynomial as in python test
	polynomial := make([]Big, n, n)
	var tmp134 Big
	asBig(&tmp134, 134)
	for i := uint64(0); i < chunkCount; i++ {
		for j, v := range []uint64{1, 2, 3, 4, 7, 8, 9, 10, 13, 14, 1, 15, 0, 1000, 0, 33} {
			asBig(&polynomial[i*chunkLen+uint64(j)], v)
		}
		subModBig(&polynomial[i*chunkLen+12], &ZERO, &ONE)    // "MODULUS - 1"
		subModBig(&polynomial[i*chunkLen+14], &ZERO, &tmp134) // "MODULUS - 134"
	}

	commitment := ks.CommitToPoly(polynomial)
	t.Log("commitment\n", strG1(commitment))

	allProofs := fk.DAUsingFK20Multi(polynomial)
	t.Log("All KZG proofs computed for data availability (supersampled by factor 2)")
	for i := 0; i < len(allProofs); i++ {
		t.Logf("%d: %s", i, strG1(&allProofs[i]))
	}

	// We have the data in polynomial form already,
	// no need to use the DAS FFT (which extends data directly, not coeffs).
	extendedCoeffs := make([]Big, n*2, n*2)
	for i := uint64(0); i < n; i++ {
		CopyBigNum(&extendedCoeffs[i], &polynomial[i])
	}
	for i := n; i < n*2; i++ {
		CopyBigNum(&extendedCoeffs[i], &ZERO)
	}
	extendedData, err := ks.FFT(extendedCoeffs, false)
	if err != nil {
		t.Fatal(err)
	}
	reverseBitOrderBig(extendedData)
	debugBigs("extended_data", extendedData)

	n2 := n * 2
	domainStride := fk.maxWidth / n2
	for pos := uint64(0); pos < 2*chunkCount; pos++ {
		domainPos := reverseBitsLimited(uint32(2*chunkCount), uint32(pos))
		var x Big
		CopyBigNum(&x, &ks.expandedRootsOfUnity[uint64(domainPos)*domainStride])
		ys := extendedData[chunkLen*pos : chunkLen*(pos+1)]
		// ys, but constructed by evaluating the polynomial in the sub-domain range
		ys2 := make([]Big, chunkLen, chunkLen)
		// don't recompute the subgroup domain, just select it from the bigger domain by applying a stride
		stride := ks.maxWidth / chunkLen
		coset := make([]Big, chunkLen, chunkLen)
		for i := uint64(0); i < chunkLen; i++ {
			var z Big // a value of the coset list
			mulModBig(&z, &x, &ks.expandedRootsOfUnity[i*stride])
			CopyBigNum(&coset[i], &z)
			EvalPolyAt(&ys2[i], polynomial, &z)
		}
		// permanently change order of ys values
		reverseBitOrderBig(ys)
		for i := 0; i < len(ys); i++ {
			if !equalBig(&ys[i], &ys2[i]) {
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
