package kate

import "fmt"

// FK20 Method to compute all proofs
// Toeplitz multiplication via http://www.netlib.org/utk/people/JackDongarra/etemplates/node384.html
// Single proof method

// A Toeplitz matrix is of the form
//
// t_0     t_(-1) t_(-2) ... t_(1-n)
// t_1     t_0    t_(-1) ... t_(2-n)
// t_2     t_1               .
// .              .          .
// .                 .       .
// .                    .    t(-1)
// t_(n-1)   ...       t_1   t_0
//
// The vector [t_0, t_1, ..., t_(n-2), t_(n-1), 0, t_(1-n), t_(2-n), ..., t_(-2), t_(-1)]
// completely determines the Toeplitz matrix and is called the "toeplitz_coefficients" below

// The composition toeplitz_part3(toeplitz_part2(toeplitz_coefficients, toeplitz_part1(x)))
// compute the matrix-vector multiplication T * x
//
// The algorithm here is written under the assumption x = G1 elements, T scalars
//
// For clarity, vectors in "Fourier space" are written with _fft. So for example, the vector
// xext is the extended x vector (padded with zero), and xext_fft is its Fourier transform.

// Performs the first part of the Toeplitz matrix multiplication algorithm, which is a Fourier
// transform of the vector x extended
func (ks *KateSettings) toeplitzPart1(x []G1) []G1 {
	n := uint64(len(x))
	n2 := n * 2
	// Extend x with zeros (neutral element of G1)
	xExt := make([]G1, n2, n2)
	for i := uint64(0); i < n; i++ {
		CopyG1(&xExt[i], &x[i])
	}
	for i := n; i < n2; i++ {
		CopyG1(&xExt[i], &zeroG1)
	}
	xExtFFT, err := ks.FFTG1(xExt, false)
	if err != nil {
		panic(fmt.Errorf("FFT G1 failed in toeplitz part 1: %v", err))
	}
	return xExtFFT
}

// Performs the second part of the Toeplitz matrix multiplication algorithm
func (ks *KateSettings) ToeplitzPart2(toeplitzCoeffs []Big, xExtFFT []G1) (hExtFFT []G1) {
	if uint64(len(toeplitzCoeffs)) != uint64(len(xExtFFT)) {
		panic("expected toeplitz coeffs to match xExtFFT length")
	}
	toeplitzCoeffsFFT, err := ks.FFT(toeplitzCoeffs, false)
	if err != nil {
		panic(fmt.Errorf("FFT failed in toeplitz part 2: %v", err))
	}
	//debugBigs("focus toeplitzCoeffsFFT", toeplitzCoeffsFFT)
	//debugG1s("xExtFFT", xExtFFT)
	n := uint64(len(toeplitzCoeffsFFT))
	//print("mul n: ", n)
	hExtFFT = make([]G1, n, n)
	for i := uint64(0); i < n; i++ {
		mulG1(&hExtFFT[i], &xExtFFT[i], &toeplitzCoeffsFFT[i])
	}
	//debugG1s("hExtFFT", hExtFFT)
	return hExtFFT
}

// Transform back and return the first half of the vector
func (ks *KateSettings) ToeplitzPart3(hExtFFT []G1) []G1 {
	out, err := ks.FFTG1(hExtFFT, true)
	if err != nil {
		panic(fmt.Errorf("toeplitz part 3 err: %v", err))
	}
	// Only the top half is the Toeplitz product, the rest is padding
	return out[:len(out)/2]
}

func (ks *KateSettings) toeplitzCoeffsStepStrided(polynomial []Big, offset uint64, stride uint64) []Big {
	n := uint64(len(polynomial))
	k := n / stride
	k2 := k * 2
	// [last poly item] + [0]*(n+1) + [poly items except first and last]
	toeplitzCoeffs := make([]Big, k2, k2)
	CopyBigNum(&toeplitzCoeffs[0], &polynomial[n-1-offset-stride])
	for i := uint64(1); i <= k+1; i++ {
		CopyBigNum(&toeplitzCoeffs[i], &ZERO)
	}
	for i, j := k+2, 2*stride-offset-1; i < k2; i, j = i+1, j+stride {
		CopyBigNum(&toeplitzCoeffs[i], &polynomial[j])
	}
	return toeplitzCoeffs
}

// TODO: call above with offset=0, stride=1
func (ks *KateSettings) toeplitzCoeffsStep(polynomial []Big) []Big {
	n := uint64(len(polynomial))
	n2 := n * 2
	// [last poly item] + [0]*(n+1) + [poly items except first and last]
	toeplitzCoeffs := make([]Big, n2, n2)
	CopyBigNum(&toeplitzCoeffs[0], &polynomial[n-1])
	for i := uint64(1); i <= n+1; i++ {
		CopyBigNum(&toeplitzCoeffs[i], &ZERO)
	}
	for i, j := n+2, 1; i < n2; i, j = i+1, j+1 {
		CopyBigNum(&toeplitzCoeffs[i], &polynomial[j])
	}
	return toeplitzCoeffs
}

// Compute all n (single) proofs according to FK20 method
func (fk *FK20SingleSettings) FK20Single(polynomial []Big) []G1 {
	toeplitzCoeffs := fk.toeplitzCoeffsStep(polynomial)
	// Compute the vector h from the paper using a Toeplitz matrix multiplication
	hExtFFT := fk.ToeplitzPart2(toeplitzCoeffs, fk.xExtFFT)
	h := fk.ToeplitzPart3(hExtFFT)

	// TODO: correct? It will pad up implicitly again, but
	out, err := fk.FFTG1(h, false)
	if err != nil {
		panic(err)
	}
	return out
}

// Special version of the FK20 for the situation of data availability checks:
// The upper half of the polynomial coefficients is always 0, so we do not need to extend to twice the size
// for Toeplitz matrix multiplication
func (fk *FK20SingleSettings) FK20SingleDAOptimized(polynomial []Big) []G1 {
	if uint64(len(polynomial)) > fk.maxWidth {
		panic(fmt.Errorf(
			"expected input of length %d (incl half of zeroes) to not exceed precomputed settings length %d",
			len(polynomial), fk.maxWidth))
	}
	n2 := uint64(len(polynomial))
	if !isPowerOfTwo(n2) {
		panic(fmt.Errorf("expected input length to be power of two, got %d", n2))
	}
	n := n2 / 2
	for i := n; i < n2; i++ {
		if !equalZero(&polynomial[i]) {
			panic("bad input, second half should be zeroed")
		}
	}
	reducedPoly := polynomial[:n]
	toeplitzCoeffs := fk.toeplitzCoeffsStep(reducedPoly)
	// Compute the vector h from the paper using a Toeplitz matrix multiplication
	hExtFFT := fk.ToeplitzPart2(toeplitzCoeffs, fk.xExtFFT)
	h := fk.ToeplitzPart3(hExtFFT)

	// Now redo the padding before final step.
	// Instead of copying h into a new extended array, just reuse the old capacity.
	h = h[:n2]
	for i := n; i < n2; i++ {
		CopyG1(&h[i], &zeroG1)
	}
	out, err := fk.FFTG1(h, false)
	if err != nil {
		panic(err)
	}
	return out
}

// Computes all the KZG proofs for data availability checks. This involves sampling on the double domain
// and reordering according to reverse bit order
func (fk *FK20SingleSettings) DAUsingFK20(polynomial []Big) []G1 {
	n := uint64(len(polynomial))
	if n > fk.maxWidth/2 {
		panic("expected poly contents not bigger than half the size of the FK20-single settings")
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
	allProofs := fk.FK20SingleDAOptimized(extendedPolynomial)
	// change to reverse bit order.
	reverseBitOrderG1(allProofs)
	return allProofs
}
