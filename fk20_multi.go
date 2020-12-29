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
func (ks *FK20MultiSettings) FK20Multi(polynomial []Big) []G1 {
	n := uint64(len(polynomial))
	n2 := n * 2
	if ks.maxWidth < n2 {
		panic(fmt.Errorf("KateSettings are set to maxWidth %d but got half polynomial of length %d",
			ks.maxWidth, n))
	}

	hExtFFT := make([]G1, n2, n2)
	for i := uint64(0); i < n2; i++ {
		CopyG1(&hExtFFT[i], &zeroG1)
	}

	var tmp G1
	for i := uint64(0); i < ks.chunkLen; i++ {
		toeplitzCoeffs := ks.toeplitzCoeffsStepStrided(polynomial, i, ks.chunkLen)
		hExtFFTFile := ks.ToeplitzPart2(toeplitzCoeffs, ks.xExtFFTFiles[i])
		for j := uint64(0); j < n2; j++ {
			addG1(&tmp, &hExtFFT[j], &hExtFFTFile[j])
			CopyG1(&hExtFFT[j], &tmp)
		}
	}
	h := ks.ToeplitzPart3(hExtFFT)

	out, err := ks.FFTG1(h, false)
	if err != nil {
		panic(err)
	}
	return out
}

// FK20 multi-proof method, optimized for dava availability where the top half of polynomial
// coefficients == 0
func (ks *FK20MultiSettings) FK20MultiDAOptimized(polynomial []Big) []G1 {
	n2 := uint64(len(polynomial))
	if ks.maxWidth < n2 {
		panic(fmt.Errorf("KateSettings are set to maxWidth %d but got polynomial of length %d",
			ks.maxWidth, n2))
	}
	n := n2 / 2
	for i := n; i < n2; i++ {
		if !equalZero(&polynomial[i]) {
			panic("bad input, second half should be zeroed")
		}
	}

	k := n / ks.chunkLen
	k2 := k * 2
	hExtFFT := make([]G1, k2, k2)
	for i := uint64(0); i < k2; i++ {
		CopyG1(&hExtFFT[i], &zeroG1)
	}

	reducedPoly := polynomial[:n]
	var tmp G1
	for i := uint64(0); i < ks.chunkLen; i++ {
		toeplitzCoeffs := ks.toeplitzCoeffsStepStrided(reducedPoly, i, ks.chunkLen)
		//debugBigs(fmt.Sprintf("toeplitz coeffs %d", i), toeplitzCoeffs)
		//debugG1s(fmt.Sprintf("xExtFFTFile %d", i), ks.xExtFFTFiles[i])
		hExtFFTFile := ks.ToeplitzPart2(toeplitzCoeffs, ks.xExtFFTFiles[i])
		//debugG1s(fmt.Sprintf("hExtFFTFile %d", i), hExtFFTFile)
		for j := uint64(0); j < k2; j++ {
			addG1(&tmp, &hExtFFT[j], &hExtFFTFile[j])
			CopyG1(&hExtFFT[j], &tmp)
		}
	}
	h := ks.ToeplitzPart3(hExtFFT)

	// TODO: maybe use a G1 version of the DAS extension FFT to perform the h -> output conversion?

	// Now redo the padding before final step.
	// Instead of copying h into a new extended array, just reuse the old capacity.
	h = h[:k2]
	for i := k; i < k2; i++ {
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
func (ks *FK20MultiSettings) DAUsingFK20Multi(polynomial []Big) []G1 {
	n := uint64(len(polynomial))
	if n > ks.maxWidth/2 {
		panic("expected poly contents not bigger than half the size of the FK20-multi settings")
	}
	if !isPowerOfTwo(n) {
		panic("expected poly length to be power of two")
	}
	n2 := n * 2
	extendedPolynomial := make([]Big, n2, n2)
	for i := uint64(0); i < n; i++ {
		CopyBigNum(&extendedPolynomial[i], &polynomial[i])
	}
	for i := n; i < n2; i++ {
		CopyBigNum(&extendedPolynomial[i], &ZERO)
	}
	allProofs := ks.FK20MultiDAOptimized(extendedPolynomial)
	// change to reverse bit order.
	reverseBitOrderG1(allProofs)
	return allProofs
}
