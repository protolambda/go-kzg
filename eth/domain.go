//go:build !bignum_pure && !bignum_hol256
// +build !bignum_pure,!bignum_hol256

package eth

import (
	"math/big"

	"github.com/protolambda/go-kzg/bls"
)

var (
	BLSModulus *big.Int
	DomainFr   []bls.Fr
)

func initDomain() {
	BLSModulus = new(big.Int)
	BLSModulus.SetString(bls.ModulusStr, 10)

	// ROOT_OF_UNITY = pow(PRIMITIVE_ROOT, (MODULUS - 1) // WIDTH, MODULUS)
	primitiveRoot := big.NewInt(7)
	width := big.NewInt(int64(FieldElementsPerBlob))
	exp := new(big.Int).Div(new(big.Int).Sub(BLSModulus, big.NewInt(1)), width)
	rootOfUnity := new(big.Int).Exp(primitiveRoot, exp, BLSModulus)
	DomainFr = make([]bls.Fr, FieldElementsPerBlob)
	for i := 0; i < FieldElementsPerBlob; i++ {
		// We reverse the bits of the index as specified in https://github.com/ethereum/consensus-specs/pull/3011
		// This effectively permutes the order of the elements in Domain
		reversedIndex := reverseBits(uint64(i), FieldElementsPerBlob)
		domain := new(big.Int).Exp(rootOfUnity, big.NewInt(int64(reversedIndex)), BLSModulus)
		_ = bigToFr(&DomainFr[i], domain)
	}
}
