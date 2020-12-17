package kate

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
	// TODO
	return nil
}

// Performs the second part of the Toeplitz matrix multiplication algorithm
func (ks *KateSettings) ToeplitzPart2(toeplitzCoeffs []Big, xExtFFT []Big) (hExtFFT []G1) {
	// TODO
	return nil
}

// Transform back and return the first half of the vector
// Only the top half is the Toeplitz product, the rest is padding
func (ks *KateSettings) ToeplitzPart3(hExtFFT []G1) []G1 {
	// TODO
	return nil
}

// Compute all n (single) proofs according to FK20 method
func (ks *KateSettings) FK20Single(polynomial []Big) []G1 {
	// TODO
	return nil
}

func (ks *KateSettings) FK20SingleDAOptimized(polynomial []Big) []G1 {
	// TODO
	return nil
}

func (ks *KateSettings) DAUsingFK20(polynomial []Big) []G1 {
	// TODO
	return nil
}
