// +build !bignum_pure,!bignum_hol256

package kate

import (
	"bytes"
	"math/rand"
	"testing"
)

// setup:
// alloc random application data
// change to reverse bit order
// extend data
// compute commitment over extended data
func integrationTestSetup(scale uint8, seed int64) (data []byte, extended []Big, extendedAsPoly []Big, commit *G1, ks *KateSettings) {
	points := 1 << scale
	size := points * 31
	data = make([]byte, size, size)
	rng := rand.New(rand.NewSource(seed))
	rng.Read(data)
	for i := 0; i < 100; i++ {
		data[i] = 0
	}
	evenPoints := make([]Big, points, points)
	// big nums are set from little-endian ints. The upper byte is always zero for input data.
	// 5/8 top bits are unused, other 3 out of range for modulus.
	var tmp [32]byte
	for i := 0; i < points; i++ {
		copy(tmp[:31], data[i*31:(i+1)*31])
		BigNumFrom32(&evenPoints[i], tmp)
	}
	reverseBitOrderBig(evenPoints)
	oddPoints := make([]Big, points, points)
	for i := 0; i < points; i++ {
		CopyBigNum(&oddPoints[i], &evenPoints[i])
	}
	// scale is 1 bigger here, since extended data is twice as big
	fs := NewFFTSettings(scale + 1)
	// convert even points (previous contents of array) to odd points
	fs.DASFFTExtension(oddPoints)
	extended = make([]Big, points*2, points*2)
	for i := 0; i < len(extended); i += 2 {
		CopyBigNum(&extended[i], &evenPoints[i/2])
		CopyBigNum(&extended[i+1], &oddPoints[i/2])
	}
	s1, s2 := generateSetup("1927409816240961209460912649124", uint64(len(extended)))
	ks = NewKateSettings(fs, s1, s2)
	// get coefficient form (half of this is zeroes, but ok)
	coeffs, err := ks.FFT(extended, true)
	if err != nil {
		panic(err)
	}
	debugBigs("poly", coeffs)
	extendedAsPoly = coeffs
	// the 2nd half is all zeroes, can ignore it for faster commitment.
	commit = ks.CommitToPoly(coeffs[:points])
	return
}

type sample struct {
	proof *G1
	sub   []Big
}

func TestFullDAS(t *testing.T) {
	data, extended, extendedAsPoly, commit, ks := integrationTestSetup(10, 1234)
	// undo the bit-reverse ordering of the extended data (which was prepared after reverse-bit ordering the input data)
	reverseBitOrderBig(extended)
	debugBigs("extended data (reordered to original)", extended)

	cosetWidth := uint64(128)
	fk := NewFK20MultiSettings(ks, ks.maxWidth, cosetWidth)
	// compute proofs for cosets
	proofs := fk.FK20MultiDAOptimized(extendedAsPoly)

	// package data of cosets with respective proofs
	sampleCount := uint64(len(extended)) / cosetWidth
	samples := make([]sample, sampleCount, sampleCount)
	for i := uint64(0); i < sampleCount; i++ {
		sample := &samples[i]

		// we can just select it from the original points
		sample.sub = make([]Big, cosetWidth, cosetWidth)
		for j := uint64(0); j < cosetWidth; j++ {
			CopyBigNum(&sample.sub[j], &extended[i*cosetWidth+j])
		}
		debugBigs("sample pre-order", sample.sub)
		// Now redo the reverse bit ordering: from a continuous range to the full domain with large stride.
		// (stride equals sample length here)
		// That's what the proofs are built for, and much more viable to do efficiently.
		reverseBitOrderBig(sample.sub)

		// construct that same coset from the polynomial form, to make sure we have the correct points.
		domainPos := reverseBitsLimited(uint32(sampleCount), uint32(i))

		sample.proof = &proofs[domainPos]

		var x Big
		CopyBigNum(&x, &ks.expandedRootsOfUnity[uint64(domainPos)*(fk.maxWidth/uint64(len(extendedAsPoly)))])
		ys2 := make([]Big, cosetWidth, cosetWidth)
		// don't recompute the subgroup domain, just select it from the bigger domain by applying a stride
		stride := ks.maxWidth / cosetWidth
		coset := make([]Big, cosetWidth, cosetWidth)
		for j := uint64(0); j < cosetWidth; j++ {
			var z Big // a value of the coset list
			mulModBig(&z, &x, &ks.expandedRootsOfUnity[j*stride])
			CopyBigNum(&coset[j], &z)
			EvalPolyAt(&ys2[j], extendedAsPoly, &z)
		}
		for i := uint64(0); i < cosetWidth; i++ {
			if !equalBig(&sample.sub[i], &ys2[i]) {
				debugBigs("sample", sample.sub)
				debugBigs("ys2", ys2)
				t.Fatal("failed to reproduce matching y values for subgroup")
			}
		}
	}
	// skip sample serialization/deserialization, no network to transfer data here.

	// verify cosets individually
	extSize := sampleCount * cosetWidth
	domainStride := ks.maxWidth / extSize
	for i, sample := range samples {
		var x Big
		domainPos := uint64(reverseBitsLimited(uint32(sampleCount), uint32(i)))
		CopyBigNum(&x, &ks.expandedRootsOfUnity[domainPos*domainStride])
		if !ks.CheckProofMulti(commit, sample.proof, &x, sample.sub) {
			t.Fatalf("failed to verify proof of sample %d", i)
		}
	}

	// make some samples go missing
	partialReconstructed := make([]*Big, extSize, extSize)
	rng := rand.New(rand.NewSource(42))
	missing := 1000                  // TODO: set to 0, enable recovery once fixed.
	for i, sample := range samples { // samples are already ordered in original data order
		// make a random subset (but <= 1/2) go missing.
		if rng.Int31n(2) == 0 && missing < len(samples)/2 {
			t.Logf("not using sample %d", i)
			missing++
			continue
		}

		offset := uint64(i) * cosetWidth
		for j := uint64(0); j < cosetWidth; j++ {
			// sample contents are still reverse-bit-ordered, undo that
			partialReconstructed[offset+j] = &sample.sub[reverseBitsLimited(uint32(cosetWidth), uint32(j))]
		}
	}
	// TODO: is second half of IFFT(partialReconstructed) all zeroes? Need to apply bit-reverse ordering? Below recovery is broken.
	// recover missing data
	recovered, err := ks.ErasureCodeRecover(partialReconstructed)
	if err != nil {
		t.Fatal(err)
	}
	debugBigs("recovered", recovered)

	for i := 0; i < len(recovered); i++ {
		if !equalBig(&extended[i], &recovered[i]) {
			t.Errorf("diff %d: %s <> %s", i, bigStr(&extended[i]), bigStr(&recovered[i]))
		}
	}
	// take first half, convert back to bytes
	size := extSize / 2
	reconstructedData := make([]byte, size*31, size*31)
	for i := uint64(0); i < size; i++ {
		p := BigNumTo32(&recovered[i])
		copy(reconstructedData[i*31:(i+1)*31], p[:31])
	}

	// check that data matches original
	if !bytes.Equal(data, reconstructedData) {
		t.Fatal("failed to reconstruct original data")
	}
}

func TestFullUser(t *testing.T) {
	// setup:
	// alloc random application data
	// change to reverse bit order
	// extend data
	// compute commitment over extended data

	// construct application-layer proof for some random points
	// verify application-layer proof
}
