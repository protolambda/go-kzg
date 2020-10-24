package go_verkle

import (
	"fmt"
	"testing"
)

func benchFFTRoundtrip(scale uint8, b *testing.B) {
	fs := NewFFTSettings(scale)
	data := make([]Big, fs.width, fs.width)
	for i := uint64(0); i < fs.width; i++ {
		data[i] = asBig(i)
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
		for i, got := range res {
			if !equalBig(got, data[i]) {
				b.Fatalf("difference: %d: got: %s  expected: %s", i, bigStr(got), bigStr(data[i]))
			}
		}
	}
}
func BenchmarkFFTSettings_FFT(b *testing.B) {
	for scale := uint8(4); scale < 10; scale++ {
		b.Run(fmt.Sprintf("scale_%d", scale), func(b *testing.B) {
			benchFFTRoundtrip(scale, b)
		})
	}
}
