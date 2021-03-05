package kzg

import (
	"fmt"
	"github.com/protolambda/go-kzg/bls"
	"testing"
)

func benchFFTExtension(scale uint8, b *testing.B) {
	fs := NewFFTSettings(scale)
	data := make([]bls.Fr, fs.MaxWidth/2, fs.MaxWidth/2)
	for i := uint64(0); i < fs.MaxWidth/2; i++ {
		bls.CopyFr(&data[i], bls.RandomFr())
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
