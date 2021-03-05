// +build !bignum_pure,!bignum_hol256

package kzg

import "github.com/protolambda/go-kzg/bls"

// GenerateTestingSetup creates a setup of n values from the given secret. **for testing purposes only**
func GenerateTestingSetup(secret string, n uint64) ([]bls.G1Point, []bls.G2Point) {
	var s bls.Fr
	bls.SetFr(&s, secret)

	var sPow bls.Fr
	bls.CopyFr(&sPow, &bls.ONE)

	s1Out := make([]bls.G1Point, n, n)
	s2Out := make([]bls.G2Point, n, n)
	for i := uint64(0); i < n; i++ {
		bls.MulG1(&s1Out[i], &bls.GenG1, &sPow)
		bls.MulG2(&s2Out[i], &bls.GenG2, &sPow)
		var tmp bls.Fr
		bls.CopyFr(&tmp, &sPow)
		bls.MulModFr(&sPow, &tmp, &s)
	}
	return s1Out, s2Out
}
