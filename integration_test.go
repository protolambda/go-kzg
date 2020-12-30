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
	cosetWidth := uint64(128)
	fk := NewFK20MultiSettings(ks, ks.maxWidth, cosetWidth)
	// compute proofs for cosets
	proofs := fk.FK20MultiDAOptimized(extendedAsPoly)
	reverseBitOrderG1(proofs)

	// package data of cosets with respective proofs
	sampleCount := uint64(len(extended)) / cosetWidth
	samples := make([]sample, sampleCount, sampleCount)
	for i := uint64(0); i < sampleCount; i++ {
		sample := &samples[i]
		sample.sub = extended[i*cosetWidth : (i+1)*cosetWidth]
		sample.proof = &proofs[i]
	}
	// skip sample serialization/deserialization, no network to transfer data here.

	// make some cosets go missing
	// recover missing cosets

	// verify cosets individually
	extSize := sampleCount * cosetWidth
	domainStride := ks.maxWidth / sampleCount
	for i, sample := range samples {
		x := &ks.expandedRootsOfUnity[uint64(i)*domainStride]
		reverseBitOrderBig(sample.sub)
		if !ks.CheckProofMulti(commit, sample.proof, x, sample.sub) {
			t.Fatalf("failed to verify proof of sample %d", i)
		}
		reverseBitOrderBig(sample.sub) // undo the in-place order change
	}
	reconstructed := make([]Big, extSize, extSize)
	for i, sample := range samples {
		for j := uint64(0); j < cosetWidth; j++ {
			CopyBigNum(&reconstructed[uint64(i)*cosetWidth+j], &sample.sub[j])
		}
	}
	// undo reverse bit order
	reverseBitOrderBig(reconstructed)

	// take first half, convert back to bytes
	size := extSize / 2
	reconstructedData := make([]byte, size*31, size*31)
	for i := uint64(0); i < size; i++ {
		p := BigNumTo32(&reconstructed[i])
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
