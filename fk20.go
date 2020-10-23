package go_verkle

func semiToeplitzFFT(toeplitzCoeffs []Big, x []Big) []Big {
	if len(toeplitzCoeffs) == 0 {
		panic("no coeffs")
	}

	// TODO
	//var xextHat []Point
	//if len(x) != len(xextHat) {
	//	panic("width must have changed")
	//}
	//xext := make([]Big, 2*len(x), 2*len(x))
	//
	//for i := 0; i < len(xext); i++ {
	//	xext[len(x)+i] = ZERO
	//}
	//xextHat = fft(xext, MODULUS, ROOT_OF_UNITY, false)

	text := make([]Big, 1+len(toeplitzCoeffs)+len(toeplitzCoeffs)-1)
	text[0] = toeplitzCoeffs[0]
	for i := 1; i <= len(toeplitzCoeffs); i++ {
		text[i] = ZERO
	}
	copy(text[len(toeplitzCoeffs)+1:], toeplitzCoeffs[1:])

	textHat := FFT(text, MODULUS, ROOT_OF_UNITY2, false)
	yextHat := make([]Big, 2*len(x))
	for i := 0; i < len(x); i++ {
		//if type(xext_hat[0]) == tuple:
		//	yext_hat[i] = b.multiply(xext_hat[i], text_hat[i])
		//else:
		yextHat[i] = mulModBig(yextHat[i], textHat[i], MODULUS)
	}
	return FFT(yextHat, MODULUS, ROOT_OF_UNITY2, true)[:len(x)]
}

// TODO
type Point struct{}

type Setup struct {
	seriesG1   []Point
	seriesG2   []Point
	lagrangeG1 []Point
	lagrangeG2 []Point
}

// Get polynomial coefficients using IFFT
func generateAllProofs(values []Big, setup *Setup) []Big {
	if len(values) != WIDTH {
		panic("bad values")
	}
	// Generate polynomial coefficients
	coeffs := FFT(values, MODULUS, ROOT_OF_UNITY, true)

	// Toeplitz matrix multiplication
	var h []Big
	{
		coeffs = append(coeffs, ZERO)
		x := make([]Point, len(values)-2, len(values)-2)
		setupG1 := setup.seriesG1
		for i := 0; i < len(setupG1)-2; i++ {
			x[len(x)-1-i] = setupG1[i]
		}

		// TODO not with big ints
		//h = semiToeplitzFFT(coeffs, x)
	}

	// final FFT
	return FFT(h, MODULUS, ROOT_OF_UNITY, false)
}

// TODO
//# Generates the data and commitent tree for a piece of data
//# as well as the precomputed proofs
//def generate_tree(data, setup):
