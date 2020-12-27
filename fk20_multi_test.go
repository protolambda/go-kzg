package kate

import "testing"

func TestKateSettings_DAUsingFK20Multi(t *testing.T) {
	fs := NewFFTSettings(5)
	chunkLen := uint64(16)
	chunkCount := uint64(32)
	n := chunkLen * chunkCount
	s1, s2 := generateSetup("1927409816240961209460912649124", chunkLen*chunkCount)
	ks := NewKateSettings(fs, chunkLen, s1, s2)

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

	allProofs := ks.DAUsingFK20Multi(polynomial)
	t.Log("All KZG proofs computed for data availability (supersampled by factor 2)")
	for i := 0; i < len(allProofs); i++ {
		t.Logf("%d: %s", i, strG1(&allProofs[i]))
	}

	// Now check all positions
	// we don't want to modify the original input, and the inner function would modify it in-place, so make a copy.
	oddData := make([]Big, n, n)
	for i := 0; i < len(oddData); i++ {
		CopyBigNum(&oddData[i], &polynomial[i])
	}
	// get the odd data (input even data)
	fs.DASFFTExtension(oddData)
	// we need both here, merge them into one array
	extendedData := make([]Big, n*2, n*2)
	for i := uint64(0); i < n; i++ {
		CopyBigNum(&extendedData[i*2], &polynomial[i])
		CopyBigNum(&extendedData[i*2+1], &oddData[i])
	}

	for pos := uint64(0); pos < 2*chunkCount; pos++ {
		domainPos := reverseBitsLimited(uint32(2*chunkCount), uint32(pos))
		var x Big
		CopyBigNum(&x, &ks.expandedRootsOfUnity[domainPos])
		ys := extendedData[chunkLen*pos : chunkLen*(pos+1)]
		ys2 := make([]Big, chunkLen, chunkLen)
		// don't recompute the subgroup domain, just select it from the bigger domain by applying a stride
		stride := (uint64(len(ks.expandedRootsOfUnity)) - 1) / chunkLen
		for i := uint64(0); i < chunkLen; i++ {
			var z Big // a value of the coset list
			CopyBigNum(&z, &ks.expandedRootsOfUnity[i*stride])
			EvalPolyAt(&ys2[i], polynomial, &z)
		}
		// permanently change order of ys values
		reverseBitOrderBig(ys)
		for i := 0; i < len(ys); i++ {
			if !equalBig(&ys[i], &ys[2]) {
				t.Fatal("failed to reproduce matching y values for subgroup")
			}
		}
		t.Log("x:\n", bigStr(&x))

		proof := &allProofs[pos]
		if !ks.CheckProofMulti(commitment, proof, &x, ys) {
			t.Fatal("could not verify proof")
		}
	}
}
