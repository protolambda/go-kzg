package go_verkle

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
	halfWidth := fs.width/2
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
