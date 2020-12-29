// +build !bignum_pure,!bignum_hol256

package kate

import (
	"fmt"
	"testing"
)

func benchFFTG1(scale uint8, b *testing.B) {
	fs := NewFFTSettings(scale)
	data := make([]G1, fs.maxWidth, fs.maxWidth)
	for i := uint64(0); i < fs.maxWidth; i++ {
		var tmpG1 G1
		CopyG1(&tmpG1, &genG1)
		mulG1(&data[i], &tmpG1, randomBig())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := fs.FFTG1(data, false)
		if err != nil {
			b.Fatal(err)
		}
		if len(out) != len(data) {
			panic("output len doesn't match input")
		}
	}
}

func BenchmarkFFTSettings_FFTG1(b *testing.B) {
	for scale := uint8(4); scale < 16; scale++ {
		b.Run(fmt.Sprintf("scale_%d", scale), func(b *testing.B) {
			benchFFTG1(scale, b)
		})
	}
}
