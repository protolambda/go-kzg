package kate

import (
	"fmt"
	"testing"
)

func benchFFTRoundtrip(scale uint8, b *testing.B) {
	fs := NewFFTSettings(scale)
	data := make([]Big, fs.width, fs.width)
	for i := uint64(0); i < fs.width; i++ {
		asBig(&data[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coeffs, err := fs.FFT(data, false)
		if err != nil {
			b.Fatal(err)
		}
		res, err := fs.FFT(coeffs, true)
		if err != nil {
			b.Fatal(err)
		}
		for i := range res {
			if got, expected := &res[i], &data[i]; !equalBig(got, expected) {
				b.Fatalf("difference: %d: got: %s  expected: %s", i, bigStr(got), bigStr(expected))
			}
		}
	}
}

func BenchmarkFFTSettings_FFT(b *testing.B) {
	for scale := uint8(4); scale < 16; scale++ {
		b.Run(fmt.Sprintf("scale_%d", scale), func(b *testing.B) {
			benchFFTRoundtrip(scale, b)
		})
	}
}

func benchFFTExtension(scale uint8, b *testing.B) {
	fs := NewFFTSettings(scale)
	data := make([]Big, fs.width/2, fs.width/2)
	for i := uint64(0); i < fs.width/2; i++ {
		asBig(&data[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// it alternates between producing values for odd indices,
		// and retrieving back the original data (but it's rotated by 1 index)
		fs.DASFFTExtension(data)
	}
}

func BenchmarkFFTExtension(b *testing.B) {
	for scale := uint8(4); scale < 16; scale++ {
		b.Run(fmt.Sprintf("scale_%d", scale), func(b *testing.B) {
			benchFFTExtension(scale, b)
		})
	}
}
