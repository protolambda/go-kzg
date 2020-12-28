package kate

import (
	"fmt"
	"testing"
)

func benchFFT(scale uint8, b *testing.B) {
	fs := NewFFTSettings(scale)
	data := make([]Big, fs.maxWidth, fs.maxWidth)
	for i := uint64(0); i < fs.maxWidth; i++ {
		CopyBigNum(&data[i], randomBig())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := fs.FFT(data, false)
		if err != nil {
			b.Fatal(err)
		}
		if len(out) != len(data) {
			panic("output len doesn't match input")
		}
	}
}

func BenchmarkFFTSettings_FFT(b *testing.B) {
	for scale := uint8(17); scale < 18; scale++ {
		b.Run(fmt.Sprintf("scale_%d", scale), func(b *testing.B) {
			benchFFT(scale, b)
		})
	}
}

func benchFFTExtension(scale uint8, b *testing.B) {
	fs := NewFFTSettings(scale)
	data := make([]Big, fs.maxWidth/2, fs.maxWidth/2)
	for i := uint64(0); i < fs.maxWidth/2; i++ {
		asBig(&data[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// it alternates between producing values for odd indices,
		// and retrieving back the original data (but it's rotated by 1 index)
		fs.DASFFTExtension(data)
		fs.DASFFTExtension(data)
	}
}

func BenchmarkFFTExtension(b *testing.B) {
	for scale := uint8(15); scale < 16; scale++ {
		b.Run(fmt.Sprintf("scale_%d", scale), func(b *testing.B) {
			benchFFTExtension(scale, b)
		})
	}
}

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
