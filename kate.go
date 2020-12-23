package kate

import "fmt"

type KateSettings struct {
	*FFTSettings

	// setup values
	// [b.multiply(b.G1, pow(s, i, MODULUS)) for i in range(WIDTH+1)],
	secretG1         []G1
	extendedSecretG1 []G1
	// [b.multiply(b.G2, pow(s, i, MODULUS)) for i in range(WIDTH+1)],
	secretG2 []G2
	// [b.multiply(b.G1, field.eval_poly_at(l, s)) for l in LAGRANGE_POLYS],
	zeroG1 []G1
	// [b.multiply(b.G2, field.eval_poly_at(l, s)) for l in LAGRANGE_POLYS],
	zeroG2 []G2

	// size of width
	xExtFFT []G1
}

func NewKateSettings(fs *FFTSettings, secretG1 []G1, secretG2 []G2) *KateSettings {
	if len(secretG1) != len(secretG2) {
		panic("secret list lengths don't match")
	}
	if uint64(len(secretG1)) != fs.width+1 {
		panic(fmt.Errorf("expected different width secrets: %d, got: %d", fs.width+1, len(secretG1)))
	}

	// TODO init extended secrets (i.e. 1st half would be the secret vals, 2nd half would be zero points), necessary for Toeplitz trickery
	// That would be:

	ks := &KateSettings{
		FFTSettings:      fs,
		secretG1:         secretG1,
		extendedSecretG1: nil,
		secretG2:         secretG2,
		zeroG1:           nil,
		zeroG2:           nil,
	}

	//x = setup[0][n - 2::-1] + [b.Z1]
	//xext_fft = toeplitz_part1(x)
	n := ks.width / 2
	x := make([]G1, n, n)
	for i := uint64(0); i < n-1; i++ {
		CopyG1(&x[i], &ks.secretG1[n-1-i])
	}
	CopyG1(&x[n-1], &zeroG1)
	ks.xExtFFT = ks.toeplitzPart1(x)

	// TODO init zeroing points

	return ks
}
