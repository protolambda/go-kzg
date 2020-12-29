package kate

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestDASFFTExtension(t *testing.T) {
	fs := NewFFTSettings(4)
	half := fs.maxWidth / 2
	data := make([]Big, half, half)
	for i := uint64(0); i < half; i++ {
		asBig(&data[i], i)
	}
	debugBigs("even data", data)
	fs.DASFFTExtension(data)
	debugBigs("odd data", data)
	bigNumHelper := func(v string) (out Big) {
		bigNum(&out, v)
		return
	}
	expected := []Big{
		bigNumHelper("40848550508281085032507004530576241411780082424652766156356301038276798860159"),
		bigNumHelper("6142039928270026418094108197259568351390689035055085818561263188953927618475"),
		bigNumHelper("11587324666845105443475591151536072107134200545187128887977105192896420361353"),
		bigNumHelper("22364018979440222199939016627179319600179064631238957550218800890988804372329"),
		bigNumHelper("11587324666845105443475591151536072107134200545187128887977105192896420361353"),
		bigNumHelper("6142039928270026418094108197259568351390689035055085818561263188953927618475"),
		bigNumHelper("40848550508281085032507004530576241411780082424652766156356301038276798860159"),
		bigNumHelper("17787776339145915450250797138634814172282648860553994191802836368572645501264"),
	}
	for i := range data {
		if got := &data[i]; !equalBig(got, &expected[i]) {
			t.Errorf("difference: %d: got: %s  expected: %s", i, bigStr(got), bigStr(&expected[i]))
		}
	}
}

func TestParametrizedDASFFTExtension(t *testing.T) {
	testScale := func(seed int64, scale uint8, t *testing.T) {
		fs := NewFFTSettings(scale)
		evenData := make([]Big, fs.maxWidth/2, fs.maxWidth/2)
		rng := rand.New(rand.NewSource(seed))
		for i := uint64(0); i < fs.maxWidth/2; i++ {
			asBig(&evenData[i], rng.Uint64()) // TODO could be a full random F_r instead of uint64
		}
		debugBigs("input data", evenData)
		// we don't want to modify the original input, and the inner function would modify it in-place, so make a copy.
		oddData := make([]Big, fs.maxWidth/2, fs.maxWidth/2)
		for i := 0; i < len(oddData); i++ {
			CopyBigNum(&oddData[i], &evenData[i])
		}
		fs.DASFFTExtension(oddData)
		debugBigs("output data", oddData)

		// reconstruct data
		data := make([]Big, fs.maxWidth, fs.maxWidth)
		for i := uint64(0); i < fs.maxWidth; i += 2 {
			CopyBigNum(&data[i], &evenData[i>>1])
			CopyBigNum(&data[i+1], &oddData[i>>1])
		}
		debugBigs("reconstructed data", data)
		// get coefficients of reconstructed data with inverse FFT
		coeffs, err := fs.FFT(data, true)
		if err != nil {
			t.Fatal(err)
		}
		debugBigs("coeffs data", coeffs)
		// second half of all coefficients should be zero
		for i := fs.maxWidth / 2; i < fs.maxWidth; i++ {
			if !equalZero(&coeffs[i]) {
				t.Errorf("expected zero coefficient on index %d", i)
			}
		}
	}
	for scale := uint8(4); scale < 10; scale++ {
		for i := int64(0); i < 4; i++ {
			t.Run(fmt.Sprintf("scale_%d_i_%d", scale, i), func(t *testing.T) {
				testScale(i, scale, t)
			})
		}
	}
}
