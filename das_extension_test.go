package kzg

import (
	"fmt"
	"github.com/protolambda/go-kzg/bls"
	"math/rand"
	"testing"
)

func TestDASFFTExtension(t *testing.T) {
	fs := NewFFTSettings(4)
	half := fs.maxWidth / 2
	data := make([]bls.Fr, half, half)
	for i := uint64(0); i < half; i++ {
		bls.AsFr(&data[i], i)
	}
	debugFrs("even data", data)
	fs.DASFFTExtension(data)
	debugFrs("odd data", data)
	ToFr := func(v string) (out bls.Fr) {
		bls.SetFr(&out, v)
		return
	}
	expected := []bls.Fr{
		ToFr("40848550508281085032507004530576241411780082424652766156356301038276798860159"),
		ToFr("6142039928270026418094108197259568351390689035055085818561263188953927618475"),
		ToFr("11587324666845105443475591151536072107134200545187128887977105192896420361353"),
		ToFr("22364018979440222199939016627179319600179064631238957550218800890988804372329"),
		ToFr("11587324666845105443475591151536072107134200545187128887977105192896420361353"),
		ToFr("6142039928270026418094108197259568351390689035055085818561263188953927618475"),
		ToFr("40848550508281085032507004530576241411780082424652766156356301038276798860159"),
		ToFr("17787776339145915450250797138634814172282648860553994191802836368572645501264"),
	}
	for i := range data {
		if got := &data[i]; !bls.EqualFr(got, &expected[i]) {
			t.Errorf("difference: %d: got: %s  expected: %s", i, bls.FrStr(got), bls.FrStr(&expected[i]))
		}
	}
}

func TestParametrizedDASFFTExtension(t *testing.T) {
	testScale := func(seed int64, scale uint8, t *testing.T) {
		fs := NewFFTSettings(scale)
		evenData := make([]bls.Fr, fs.maxWidth/2, fs.maxWidth/2)
		rng := rand.New(rand.NewSource(seed))
		for i := uint64(0); i < fs.maxWidth/2; i++ {
			bls.AsFr(&evenData[i], rng.Uint64()) // TODO could be a full random F_r instead of uint64
		}
		debugFrs("input data", evenData)
		// we don't want to modify the original input, and the inner function would modify it in-place, so make a copy.
		oddData := make([]bls.Fr, fs.maxWidth/2, fs.maxWidth/2)
		for i := 0; i < len(oddData); i++ {
			bls.CopyFr(&oddData[i], &evenData[i])
		}
		fs.DASFFTExtension(oddData)
		debugFrs("output data", oddData)

		// reconstruct data
		data := make([]bls.Fr, fs.maxWidth, fs.maxWidth)
		for i := uint64(0); i < fs.maxWidth; i += 2 {
			bls.CopyFr(&data[i], &evenData[i>>1])
			bls.CopyFr(&data[i+1], &oddData[i>>1])
		}
		debugFrs("reconstructed data", data)
		// get coefficients of reconstructed data with inverse FFT
		coeffs, err := fs.FFT(data, true)
		if err != nil {
			t.Fatal(err)
		}
		debugFrs("coeffs data", coeffs)
		// second half of all coefficients should be zero
		for i := fs.maxWidth / 2; i < fs.maxWidth; i++ {
			if !bls.EqualZero(&coeffs[i]) {
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
