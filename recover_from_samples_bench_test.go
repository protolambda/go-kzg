package kzg

import (
	"fmt"
	"math/rand"
	"testing"
)

func benchRecoverPolyFromSamples(scale uint8, seed int64, b *testing.B) {
	fs := NewFFTSettings(scale)
	poly := make([]Big, fs.maxWidth, fs.maxWidth)
	for i := uint64(0); i < fs.maxWidth/2; i++ {
		asBig(&poly[i], i)
	}
	rng := rand.New(rand.NewSource(seed))
	data, _ := fs.FFT(poly, false)
	samples := make([]*Big, fs.maxWidth, fs.maxWidth)
	for i := 0; i < len(data); i++ {
		samples[i] = &data[i]
	}
	// randomly zero out half of the points
	for i := 0; i < len(samples)/2; i++ {
		j := rng.Intn(len(samples))
		for samples[j] == nil {
			j = rng.Intn(len(samples))
		}
		samples[j] = nil
	}

	b.ResetTimer()

	for bi := 0; bi < b.N; bi++ {
		recovered, err := fs.RecoverPolyFromSamples(samples, fs.ZeroPolyViaMultiplication)
		if err != nil {
			b.Fatal(err)
		}
		for i := 0; i < len(data); i++ {
			if !equalBig(&recovered[i], &data[i]) {
				b.Fatalf("bad recovered output %d: %s <> %s", i, bigStr(&recovered[i]), bigStr(&data[i]))
			}
		}
	}
}

func BenchmarkFFTSettings_RecoverPolyFromSamples(b *testing.B) {
	for scale := uint8(5); scale < 16; scale++ {
		b.Run(fmt.Sprintf("scale_%d", scale), func(b *testing.B) {
			benchRecoverPolyFromSamples(scale, int64(scale), b)
		})
	}
}
