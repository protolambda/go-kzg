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

	xExtFFT []G1 // TODO: sizing this. Maybe pre-compute all possible sizes within maxWidth?

	// TODO: maybe refactor chunks into separate settings?
	chunkLen uint64
	// chunkLen files, each of size maxWidth
	xExtFFTFiles [][]G1
}

func NewKateSettings(fs *FFTSettings, chunkLen uint64, secretG1 []G1, secretG2 []G2) *KateSettings {
	if len(secretG1) != len(secretG2) {
		panic("secret list lengths don't match")
	}
	if uint64(len(secretG1)) < fs.maxWidth {
		panic(fmt.Errorf("expected more values for secrets, maxWidth: %d, got: %d", fs.maxWidth, len(secretG1)))
	}

	ks := &KateSettings{
		FFTSettings:      fs,
		secretG1:         secretG1,
		extendedSecretG1: nil,
		secretG2:         secretG2,
	}

	//x = setup[0][n - 2::-1] + [b.Z1]
	//xext_fft = toeplitz_part1(x)
	xExtFFTPrecompute := func(offset uint64, stride uint64) []G1 {
		n := ks.maxWidth / 2
		x := make([]G1, n, n)
		for i, j := uint64(0), n-2*stride; i < n-1; i, j = i+1, j-stride {
			CopyG1(&x[i], &ks.secretG1[j])
		}
		CopyG1(&x[n-1], &zeroG1)
		return ks.toeplitzPart1(x)
	}

	ks.xExtFFT = xExtFFTPrecompute(0, 1)
	ks.xExtFFTFiles = make([][]G1, chunkLen, chunkLen)
	for i := uint64(0); i < chunkLen; i++ {
		ks.xExtFFTFiles[i] = xExtFFTPrecompute(i, chunkLen)
	}

	// TODO init zeroing points

	return ks
}
