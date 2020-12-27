package kate

import "fmt"

// FK20 Method to compute all proofs
// Toeplitz multiplication via http://www.netlib.org/utk/people/JackDongarra/etemplates/node384.html
// Multi proof method

// For a polynomial of size n, let w be a n-th root of unity. Then this method will return
// k=n/l KZG proofs for the points
//
// 	   proof[0]: w^(0*l + 0), w^(0*l + 1), ... w^(0*l + l - 1)
// 	   proof[0]: w^(0*l + 0), w^(0*l + 1), ... w^(0*l + l - 1)
// 	   ...
// 	   proof[i]: w^(i*l + 0), w^(i*l + 1), ... w^(i*l + l - 1)
// 	   ...
func (ks *KateSettings) FK20Multi(polynomial []Big) []G1 {
	chunkCount := uint64(len(polynomial)) * 2 / ks.chunkLen
	if ks.width != chunkCount {
		panic(fmt.Errorf("KateSettings are set to width %d and chunkLen %d,"+
			" but got chunk count %d mismatching the width (polynomial len %d)",
			ks.width, ks.chunkLen, chunkCount, len(polynomial)))
	}

	hExtFFT := make([]G1, ks.width, ks.width)
	for i := uint64(0); i < ks.width; i++ {
		CopyG1(&hExtFFT[i], &zeroG1)
	}

	var tmp G1
	for i := uint64(0); i < ks.chunkLen; i++ {
		toeplitzCoeffs := ks.toeplitzCoeffsStepStrided(polynomial, i, ks.chunkLen)
		hExtFFTFile := ks.ToeplitzPart2(toeplitzCoeffs, ks.xExtFFTFiles[i])
		for j := uint64(0); j < ks.width; j++ {
			addG1(&tmp, &hExtFFT[j], &hExtFFTFile[j])
			CopyG1(&hExtFFT[j], &tmp)
		}
	}
	h := ks.ToeplitzPart3(hExtFFT)

	// TODO: correct? It will pad up implicitly again, but
	out, err := ks.FFTG1(h, false)
	if err != nil {
		panic(err)
	}
	return out
}

// FK20 multi-proof method, optimized for dava availability where the top half of polynomial
// coefficients == 0
func (ks *KateSettings) FK20MultiDAOptimized(polynomial []Big) []G1 {
	chunkCount := uint64(len(polynomial)) / ks.chunkLen
	if ks.width != chunkCount {
		panic(fmt.Errorf("KateSettings are set to width %d and chunkLen %d,"+
			" but got chunk count %d mismatching the width (polynomial len %d)",
			ks.width, ks.chunkLen, chunkCount, len(polynomial)))
	}
	n := ks.width / 2
	for i := n; i < ks.width; i++ {
		if !equalZero(&polynomial[i]) {
			panic("bad input, second half should be zeroed")
		}
	}

	hExtFFT := make([]G1, ks.width, ks.width)
	for i := uint64(0); i < ks.width; i++ {
		CopyG1(&hExtFFT[i], &zeroG1)
	}

	reducedPoly := polynomial[:n]
	var tmp G1
	for i := uint64(0); i < ks.chunkLen; i++ {
		toeplitzCoeffs := ks.toeplitzCoeffsStepStrided(reducedPoly, i, ks.chunkLen)
		hExtFFTFile := ks.ToeplitzPart2(toeplitzCoeffs, ks.xExtFFTFiles[i])
		for j := uint64(0); j < ks.width; j++ {
			addG1(&tmp, &hExtFFT[j], &hExtFFTFile[j])
			CopyG1(&hExtFFT[j], &tmp)
		}
	}
	h := ks.ToeplitzPart3(hExtFFT)

	// Now redo the padding before final step.
	// Instead of copying h into a new extended array, just reuse the old capacity.
	h = h[:ks.width]
	for i := n; i < ks.width; i++ {
		CopyG1(&h[i], &zeroG1)
	}
	out, err := ks.FFTG1(h, false)
	if err != nil {
		panic(err)
	}
	return out
}

// Computes all the KZG proofs for data availability checks. This involves sampling on the double domain
// and reordering according to reverse bit order
func (ks *KateSettings) DAUsingFK20Multi(polynomial []Big) []G1 {
	n := uint64(len(polynomial))
	if n*2 != ks.width {
		panic("expected poly contents half the size of the Kate settings")
	}
	extendedPolynomial := make([]Big, ks.width, ks.width)
	for i := uint64(0); i < n; i++ {
		CopyBigNum(&extendedPolynomial[i], &polynomial[i])
	}
	for i := n; i < ks.width; i++ {
		CopyBigNum(&extendedPolynomial[i], &ZERO)
	}
	allProofs := ks.FK20MultiDAOptimized(extendedPolynomial)
	// change to reverse bit order.
	reverseBitOrderG1(allProofs)
	return allProofs
}
