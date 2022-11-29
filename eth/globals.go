//go:build !bignum_pure && !bignum_hol256
// +build !bignum_pure,!bignum_hol256

package eth

import (
	_ "embed"
	"encoding/json"
	"math/big"

	"github.com/protolambda/go-kzg/bls"
)

var (
	BLSModulus *big.Int
	DomainFr   []bls.Fr

	// KZG CRS for G2
	kzgSetupG2 []bls.G2Point

	// KZG CRS for commitment computation
	kzgSetupLagrange []bls.G1Point

	// KZG CRS for G1 (only used in tests (for proof creation))
	KzgSetupG1 []bls.G1Point

	//go:embed trusted_setup.json
	kzgSetupStr string

	precompileReturnValue [64]byte
)

type JSONTrustedSetup struct {
	SetupG1       []bls.G1Point `json:"setup_G1"`
	SetupG2       []bls.G2Point `json:"setup_G2"`
	SetupLagrange []bls.G1Point `json:"setup_G1_lagrange"`
}

func init() {
	// Initialize KZG subsystem (load the trusted setup data)
	var parsedSetup = JSONTrustedSetup{}

	err := json.Unmarshal([]byte(kzgSetupStr), &parsedSetup)
	if err != nil {
		panic(err)
	}
	kzgSetupG2 = parsedSetup.SetupG2
	kzgSetupLagrange = bitReversalPermutation(parsedSetup.SetupLagrange)
	KzgSetupG1 = parsedSetup.SetupG1

	BLSModulus = new(big.Int)
	BLSModulus.SetString(bls.ModulusStr, 10)

	// Initialize the domain
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

	// initialize the 64 bytes of precompile return data: field elements per blob, field modulus (big-endian uint256)
	new(big.Int).SetUint64(FieldElementsPerBlob).FillBytes(precompileReturnValue[:32])
	BLSModulus.FillBytes(precompileReturnValue[32:])
}
