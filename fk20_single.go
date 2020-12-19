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
func (ks *KateSettings) ToeplitzPart1(x []G1) []G1 {
	half := ks.width / 2
	if uint64(len(x)) != half {
		panic(fmt.Errorf("expected width %d (half of settings), got %d", half, len(x)))
	}
	// Extend x with zeros (neutral element of G1)
	xExt := make([]G1, ks.width, ks.width)
	for i := uint64(0); i < half; i++ {
		CopyG1(&xExt[i], &x[i])
	}
	for i := half; i < ks.width; i++ {
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
	if uint64(len(toeplitzCoeffs)) != ks.width {
		panic("expected toeplitz coeffs to match the width")
	}
	if uint64(len(xExtFFT)) != ks.width {
		panic("expected xExtFFT to match width")
	}
	toeplitzCoeffsFFT, err := ks.FFT(toeplitzCoeffs, false)
	if err != nil {
		panic(fmt.Errorf("FFT failed in toeplitz part 2: %v", err))
	}
	hExtFFT = make([]G1, ks.width, ks.width)
	for i := uint64(0); i < ks.width; i++ {
		mulG1(&hExtFFT[i], &xExtFFT[i], &toeplitzCoeffsFFT[i])
	}
	return hExtFFT
}

// Transform back and return the first half of the vector
func (ks *KateSettings) ToeplitzPart3(hExtFFT []G1) []G1 {
	if uint64(len(hExtFFT)) != ks.width {
		panic("expected hExtFFT to match the width")
	}
	out, err := ks.FFTG1(hExtFFT, true)
	if err != nil {
		panic(fmt.Errorf("toeplitz part 3 err: %v", err))
	}
	// Only the top half is the Toeplitz product, the rest is padding
	return out[:ks.width/2]
}

// Compute all n (single) proofs according to FK20 method
func (ks *KateSettings) FK20Single(polynomial []Big) []G1 {
	half := ks.width / 2
	if uint64(len(polynomial)) != half {
		panic(fmt.Errorf(
			"expected input of length %d to match half of precomputed settings length %d",
			len(polynomial), ks.width))
	}
	// the inverse domain, but last entry zero
	x := make([]G1, ks.width, ks.width)
	for i := uint64(0); i < ks.width-1; i++ {
		CopyG1(&x[i], &ks.secretG1[ks.width-1-i])
	}
	CopyG1(&x[ks.width-1], &zeroG1)

	xExtFFT := ks.ToeplitzPart1(x)

	// [last poly item] + [0]*(half+1) + [poly items except first]
	toeplitzCoeffs := make([]Big, ks.width+1, ks.width+1)

	// Compute the vector h from the paper using a Toeplitz matrix multiplication
	hExtFFT := ks.ToeplitzPart2(toeplitzCoeffs, xExtFFT)
	h := ks.ToeplitzPart3(hExtFFT)

	out, err := ks.FFTG1(h, false)
	if err != nil {
		panic(err)
	}
	return out
}

func (ks *KateSettings) FK20SingleDAOptimized(polynomial []Big) []G1 {
	// TODO
	return nil
}

func (ks *KateSettings) DAUsingFK20(polynomial []Big) []G1 {
	// TODO
	return nil
}
