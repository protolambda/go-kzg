//go:build !bignum_pure && !bignum_hol256
// +build !bignum_pure,!bignum_hol256

package kzg

import (
	"fmt"
	"github.com/protolambda/go-kzg/bls"
	"testing"
)

func BenchmarkCommit(b *testing.B) {
	for scale := uint8(12); scale < 13; scale++ {
		b.Run(fmt.Sprintf("scale_%d", scale), func(b *testing.B) {
			benchCommit(scale, b)
		})
	}
}

func benchCommit(scale uint8, b *testing.B) {
	fs := NewFFTSettings(scale)
	setupG1, setupG2 := GenerateTestingSetup("1234", uint64(1)<<scale)
	ks := NewKZGSettings(fs, setupG1, setupG2)
	setupLagrange, err := ks.FFTG1(setupG1, true)
	if err != nil {
		b.Fatal(err)
	}
	blob := make([]bls.Fr, uint64(1)<<scale)
	for i := 0; i < len(blob); i++ {
		blob[i] = *bls.RandomFr()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bls.LinCombG1(setupLagrange, blob)
	}
}
