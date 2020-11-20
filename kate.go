package go_verkle

import "fmt"

func MakeSetup() *G1 {
	// TODO: setup: for secret point to evaluate polynomials at
	// setup values are defined as [g * s**i for i in range(m)]
	return nil
}

// TODO: input data as points

// TODO: input data to polynomial coeffs
func GetPoly(data []Big) []Big {
	// TODO
	return nil
}

func (fs *FFTSettings) MakeKateCommitment(data []Big) (*G1, error) {
	// TODO: inverse FFT of input data: get polynomial to evaluate for data
	// TODO: commitment: "C = [p(s)]_1 = [Sum(p_i * s**i)]_1 = Sum(p_i * [s**i]_1)"
	//    - get coefficients of data -> polynomial representation
	//    - evaluate at shared secret point -> commitment is "p(s)G", an elliptic curve point,
	// Attacker can't construct another polynomial that has the same commitment.
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


	fs.secretG1

	h, err := fs.semiToeplitzFFT(coeffs, )
	if err != nil {
		return nil, fmt.Errorf("could not compute h for building Kate commitment: %v", err)
	}

	// r is a list of commitments, one for each value
	r, err := fs.FFT(h, false)
	if err != nil {
		return nil, fmt.Errorf("could not compute commitment from h: %v", err)
	}
	// TODO either combine them all, or combine a subset for comitting to a subset?
	return nil, nil
}

func VerifyCommitment(setup *G1, commitment G1, values map[uint64]*Big) bool {
	// Credits to Dankrad for describing Kate proofs here:
	// https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html
	// This is just how I approach and understand it, for implementation purposes.
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

// TODO
func MakeProof(setup *G1, values []Big, indices []uint64) *G1 {
	// TODO: 
	return nil
}
