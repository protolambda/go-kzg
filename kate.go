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

	// size of width
	xExtFFT  []G1
	chunkLen uint64
	// chunkLen files, each of size width
	xExtFFTFiles [][]G1
}

func NewKateSettings(fs *FFTSettings, chunkLen uint64, secretG1 []G1, secretG2 []G2) *KateSettings {
	if len(secretG1) != len(secretG2) {
		panic("secret list lengths don't match")
	}
	if uint64(len(secretG1)) < fs.width {
		panic(fmt.Errorf("expected more values for secrets, width: %d, got: %d", fs.width, len(secretG1)))
	}
	if min := fs.width * chunkLen; min < uint64(len(secretG1)) {
		panic(fmt.Errorf("not enough secret values to cover width * chunklen: %d * %d = %d. Got: %d", fs.width, chunkLen, min, len(secretG1)))
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
		n := ks.width / 2
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
