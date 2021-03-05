package kzg

import (
	"fmt"
	"math/rand"
	"testing"
)

func benchZeroPoly(scale uint8, seed int64, b *testing.B) {
	fs := NewFFTSettings(scale)
	missing := make([]uint64, fs.MaxWidth, fs.MaxWidth)
	for i := uint64(0); i < uint64(len(missing)); i++ {
		missing[i] = i
	}
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(missing), func(i, j int) {
		missing[i], missing[j] = missing[j], missing[i]
	})
	// Only consider 50% as missing. Leaves enough FFT computation room.
	missing = missing[:len(missing)/2]
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		zeroEval, zeroPoly := fs.ZeroPolyViaMultiplication(missing, fs.MaxWidth)
		if len(zeroEval) != len(zeroPoly) {
			panic("sanity check failed, length mismatch")
		}
	}
}

func BenchmarkFFTSettings_ZeroPolyViaMultiplication(b *testing.B) {
	for scale := uint8(5); scale < 16; scale++ {
		b.Run(fmt.Sprintf("scale_%d", scale), func(b *testing.B) {
			benchZeroPoly(scale, int64(scale), b)
		})
	}
}
