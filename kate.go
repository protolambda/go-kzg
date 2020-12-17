package kate

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
}

func NewKateSettings(fs *FFTSettings, secretG1 *G1, secretG2 *G2) *KateSettings {
	// TODO
	// TODO init secret power circle from generators (in width/2)

	// TODO init extended secrets (i.e. 1st half would be the secret vals, 2nd half would be zero points), necessary for Toeplitz trickery
	// That would be:
	//xext = x + [b.Z1 for a in x]
	//xext_hat = fft(xext, MODULUS, ROOT_OF_UNITY2, inv=False)

	// TODO init zeroing points

	return &KateSettings{
		FFTSettings:      fs,
		secretG1:         nil,
		extendedSecretG1: nil,
		secretG2:         nil,
		zeroG1:           nil,
		zeroG2:           nil,
	}
}
