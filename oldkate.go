package kate

import "fmt"

func (fs *FFTSettings) semiToeplitzFFT(toeplitzCoeffs []Big) ([]G1, error) {
	if uint64(len(toeplitzCoeffs)) != fs.width/2 {
		return nil, fmt.Errorf("unexpected toeplitzCoeffs count: %d, expected half FFT width: %d",
			len(toeplitzCoeffs), fs.width)
	}

	// h = A * [S]
	// In A, the diagonal and top right of diagonal is filled with values of interest, staggered.
	// The bottom left of diagonal is filled with zeroes.
	// To create this shape in a Toeplitz matrix, pad the input as:
	//  values:  [a_0, 0, 0, 0, ...,   0, a_n-1, a_n-2, ..., a_n-(n-1)]
	//  indices: [  0, 1, 2, 3, ..., n/2, n/2+1, n/2+2, ..., n-1      ]
	//
	// This creates a circulant matrix C (half of the row filled, staggered each row),
	//  diagonalized by the FFT of the padded a values.
	//
	// I.e. C = IFFT(diagonal(FFT(pad(a)))
	//  And now to multiply C with x, it can be done efficiently by doing so in between the FFT steps.
	//
	// input:   a: t
	//     pad(a): text
	//          s: x
	//     pad(s): xext
	//      xext^: FFT(xext)
	//      text^: FFT(text)
	//      yext^: xext^ * text^
	//       yext: IFFT(yext^)
	//
	//  x* = FFT(a)
	//
	//  IFFT(y)
	text := make([]Big, fs.width, fs.width)
	// first coefficient stays first
	CopyBigNum(&text[0], &toeplitzCoeffs[0])
	halfWidth := fs.width / 2
	for i := uint64(1); i <= halfWidth; i++ {
		CopyBigNum(&text[i], &ZERO)
	}
	// then, in reverse order, the rest follows
	for i, dst := halfWidth-1, halfWidth+1; i > 0; i, dst = i-1, dst+1 {
		CopyBigNum(&text[dst], &toeplitzCoeffs[i])
	}

	textHat, err := fs.FFT(text, false)
	if err != nil {
		return nil, fmt.Errorf("failed to compute FFT of extended toeplitz coeffs")
	}

	// now multiply all toeplitz values corresponding secret domain values
	yextHat := make([]G1, fs.width, fs.width)
	for i := uint64(0); i < fs.width; i++ {
		mulG1(&yextHat[i], &fs.extendedSecretG1[i], &textHat[i])
	}
	// now take the IFFT  // TODO
	out, err := fs.FFTG1(yextHat, true)
	if err != nil {
		return nil, fmt.Errorf("failed to compute FFT of yextHat: %v", err)
	}
	return out, nil
}

// TODO: input data to polynomial coeffs (with IFFT)
func (fs *FFTSettings) GetPolyCoeffs(data []Big) []Big {
	// TODO
	return nil
}

// Commit simply to all values, unoptimized.
// Do so by evaluating the polynomial at the shared secret.
func (fs *FFTSettings) Commit(coeffs []Big) *G1 {
	var out G1
	ClearG1(&out)
	var tmp, tmp2 G1
	for i := 0; i < len(coeffs); i++ {
		mulG1(&tmp, &fs.secretG1[i], &coeffs[i])
		addG1(&tmp2, &out, &tmp)
		CopyG1(&out, &tmp2)
	}
	return &out
}

// TODO: depending on BLS library, there are optimized bindings for this function available.
// Try Herumi hbls.FrEvaluatePolynomial()
func evalPolyAt(dst *Big, coeffs []Big, x *Big) {
	if len(coeffs) == 0 {
		CopyBigNum(dst, &ZERO)
		return
	}
	if equalZero(x) {
		CopyBigNum(dst, &coeffs[0])
		return
	}
	// Horner's method: work backwards, avoid doing more than N multiplications
	// https://en.wikipedia.org/wiki/Horner%27s_method
	var last Big
	CopyBigNum(&last, &coeffs[len(coeffs)-1])
	var tmp Big
	for i := len(coeffs) - 2; i >= 0; i-- {
		mulModBig(&tmp, &last, x)
		addModBig(&last, &tmp, &coeffs[i])
	}
	CopyBigNum(dst, &last)
}

func polyLongDiv(dst []Big, dividend []Big, divisor []Big) {
	// TODO
}

func (fs *FFTSettings) MakeProof(coeffs []Big, index uint64) *G1 {
	var x Big
	asBig(&x, index)

	// evaluation
	var y Big
	evalPolyAt(&y, coeffs, &x)

	// dividend = poly - y
	dividend := make([]Big, len(coeffs), len(coeffs))
	subModBig(&dividend[0], &coeffs[0], &y)
	for i := 1; i < len(coeffs); i++ {
		CopyBigNum(&dividend[i], &coeffs[i])
	}

	// divisor = x - index = coeffs [-index, 1]
	divisor := make([]Big, 2, 2)
	subModBig(&divisor[0], &ZERO, &x)
	CopyBigNum(&divisor[1], &ONE)

	// witness polynomial
	witnessPoly := make([]Big, len(coeffs), len(coeffs))
	polyLongDiv(witnessPoly, dividend, divisor)

	// commit evaluation
	return fs.Commit(witnessPoly)
}

func (fs *FFTSettings) Verify(commit *G1, proof *G1, value Big, index uint64, sG2 *G2) bool {
	// verify: e([q(s)]_1, [s - z]_2) == e([p(s)]_1 - [y]_1, H)
	// proof = [q(s)]_1
	// [s-z]_2 = sG2 - (index * g_2[0])
	// commit = [p(s)]_1
	// [y]_1 = value * g_1[0]

	var left, right hbls.GT
	hbls.Pairing(&left, (*hbls.G1)(proof), nil)   // TODO
	hbls.Pairing(&right, (*hbls.G1)(commit), nil) // TODO
	return left.IsEqual(&right)
}

func (fs *FFTSettings) VerifyMultiProof(commit *G1, proof *G1, values []Big, indices []uint64) bool {
	return false // TODO
}

func (fs *FFTSettings) MakeKateProofs(data []Big) ([]G1, error) {
	// TODO: inverse FFT of input data: get polynomial to evaluate for data
	// TODO: commitment: "C = [p(s)]_1 = [Sum(p_i * s**i)]_1 = Sum(p_i * [s**i]_1)"
	//    - get coefficients of data -> polynomial representation
	//    - evaluate at shared secret point -> commitment is "p(s)G", an elliptic curve point,
	// Attacker can't construct another polynomial that has the same commitment.

	// TODO: skip coeffs step, data can already be extended outside here?

	// TODO: this should use another fft settings with half the width
	coeffs, err := fs.FFT(data, true)
	if err != nil {
		return nil, fmt.Errorf("could not get polynomial coeffs of data: %v", err)
	}

	// reverse the coeffs
	var tmp Big
	for i, j := 0, len(coeffs)-1; i < j; i, j = i+1, j-1 {
		CopyBigNum(&tmp, &coeffs[i])
		CopyBigNum(&coeffs[i], &coeffs[j])
		CopyBigNum(&coeffs[j], &tmp)
	}
	// then zero out the last
	CopyBigNum(&coeffs[len(coeffs)-1], &ZERO)

	h, err := fs.semiToeplitzFFT(coeffs)
	if err != nil {
		return nil, fmt.Errorf("could not compute h for building Kate commitment: %v", err)
	}

	// r is a list of commitments, one for each value
	r, err := fs.FFTG1(h, false)
	if err != nil {
		return nil, fmt.Errorf("could not compute commitment from h: %v", err)
	}
	// TODO API design: either combine them all, or combine a subset for comitting to a subset? Or return all for later use?
	return r, nil
}

func VerifyCommitment(setup *G1, commitment G1, values map[uint64]*Big) bool {
	// Credits to Dankrad for describing Kate proofs for relative beginners here:
	// https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html
	// This is comment is a summary of the above for implementation purposes.
	//
	// Open the commitment, without sending the whole polynomial, so that we can verify some data is correct
	//
	// Ingredients:
	//   - combine commitments (on same curve) by adding them
	//   - can't multiply, but can do a pairing
	//
	// Goal: proof that p(z) = y
	//
	// I.e. proof that at z, the polynomial that was committed to evaluates to y.
	// Effectively: proof a piece of original input data exists
	//
	// A polynomial is divisible by (X - z) if it has a zero at z.
	// The converse is also true: zero at z if divisible by (X - z).
	//
	// To get p(z) = 0, factor out (X - z), and get polynomial q(X), showing it is divisible:
	// p(X) = (X - z) * q(X)
	//
	// Now instead of showing it is zero at z, we want p(z) = y.
	// So simply subtract y from the polynomial to make that work.
	// (We use addition on the commitment)
	//
	// I.e. to proof p(z) = y we need to show : q(X) * (X - z) = (p(X) - y)
	//
	// And although we can't multiply polynomials, we can verify the pairing:
	//
	// e([q(s)]_1, [s - z]_2) == e([p(s)]_1 - [y]_1, H)
	//
	// [p(s)]_1 is the commitment
	// [q(s)]_1 here is the proof data
	// [s - z]_2 is for the (X - z) part at s, on the other curve
	// H is some helper
	//
	// Correctness: nobody knows s, so it can safely be used for X
	// Soundness:
	// - work back from above approach: there's only a single y value for which the p(z) - y = 0
	// - work back from commitment and secret: see better description by Dankrad, depends on q-strong SDH assumption
	//
	// Multi-proofs:
	// Instead of just a single (z, y), show it for a series of (z_i, y_i) pairs
	//
	// Now to repeat the p(z) = 0, an "interpolation polynomial" I(X) is used:
	// a polynomial to subtract from P(X) so that the resulting polynomial is zero at all z values of interest.
	//
	// And again, factor out (X - z), but for all z values, to show divisibility for each.
	// This combined (X - z) terms will be called the zero polynomial Z(X)
	//
	// And to proof p(z) = z for all z now: q(X) * Z(X) = p(x) - I(X)
	// And then the pairing verification becomes:
	// e([q(s)]_1, [Z(s)]_2) == e([p(s)]_1 - [I(s)]_1, H)
	//
	// Note that the format for the commitment and proof didn't change:
	// any amount of evaluations can be proven with a single group element.
	//
	return false
}
