package kzg

import (
	"fmt"
	"testing"
)

func benchFFTExtension(scale uint8, b *testing.B) {
	fs := NewFFTSettings(scale)
	data := make([]Big, fs.maxWidth/2, fs.maxWidth/2)
	for i := uint64(0); i < fs.maxWidth/2; i++ {
		CopyBigNum(&data[i], randomBig())
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
