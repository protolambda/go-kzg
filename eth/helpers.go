//go:build !bignum_pure && !bignum_hol256
// +build !bignum_pure,!bignum_hol256

package eth

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"math/bits"

	"github.com/protolambda/go-kzg/bls"
)

const (
	FIAT_SHAMIR_PROTOCOL_DOMAIN = "FSBLOBVERIFY_V1_"
)

type Polynomial []bls.Fr
type Polynomials [][]bls.Fr

// Bit-reversal permutation helper functions

// Check if `value` is a power of two integer.
func isPowerOfTwo(value uint64) bool {
	return value > 0 && (value&(value-1) == 0)
}

// Reverse `order` bits of integer n
func reverseBits(n, order uint64) uint64 {
	if !isPowerOfTwo(order) {
		panic("Order must be a power of two.")
	}

	return bits.Reverse64(n) >> (65 - bits.Len64(order))
}

// Return a copy of the input array permuted by bit-reversing the indexes.
func bitReversalPermutation(l []bls.G1Point) []bls.G1Point {
	out := make([]bls.G1Point, len(l))

	order := uint64(len(l))

	for i := range l {
		out[i] = l[reverseBits(uint64(i), order)]
	}

	return out
}

// VerifyKZGProofFromPoints implements verify_kzg_proof from the EIP-4844 consensus spec,
// only with the byte inputs already parsed into points & field elements.
func VerifyKZGProofFromPoints(polynomialKZG *bls.G1Point, z *bls.Fr, y *bls.Fr, kzgProof *bls.G1Point) bool {
	var zG2 bls.G2Point
	bls.MulG2(&zG2, &bls.GenG2, z)
	var yG1 bls.G1Point
	bls.MulG1(&yG1, &bls.GenG1, y)

	var xMinusZ bls.G2Point
	bls.SubG2(&xMinusZ, &kzgSetupG2[1], &zG2)
	var pMinusY bls.G1Point
	bls.SubG1(&pMinusY, polynomialKZG, &yG1)

	return bls.PairingsVerify(&pMinusY, &bls.GenG2, kzgProof, &xMinusZ)
}

// VerifyAggregateKZGProofFromPolynomials implements verify_aggregate_kzg_proof from the EIP-4844 consensus spec,
// only operating on blobs that have already been converted into polynomials.
func VerifyAggregateKZGProofFromPolynomials(blobs Polynomials, expectedKZGCommitments KZGCommitmentSequence, kzgAggregatedProof KZGProof) (bool, error) {
	aggregatedPoly, aggregatedPolyCommitment, evaluationChallenge, err :=
		ComputeAggregatedPolyAndCommitment(blobs, expectedKZGCommitments)
	if err != nil {
		return false, err
	}
	y := EvaluatePolynomialInEvaluationForm(aggregatedPoly, evaluationChallenge)
	kzgProofG1, err := bls.FromCompressedG1(kzgAggregatedProof[:])
	if err != nil {
		return false, fmt.Errorf("failed to decode kzgProof: %v", err)
	}
	return VerifyKZGProofFromPoints(aggregatedPolyCommitment, evaluationChallenge, y, kzgProofG1), nil
}

// ComputePowers implements compute_powers from the EIP-4844 consensus spec:
// https://github.com/ethereum/consensus-specs/blob/dev/specs/eip4844/polynomial-commitments.md#compute_powers
func ComputePowers(r *bls.Fr, n int) []bls.Fr {
	var currentPower bls.Fr
	bls.AsFr(&currentPower, 1)
	powers := make([]bls.Fr, n)
	for i := range powers {
		powers[i] = currentPower
		bls.MulModFr(&currentPower, &currentPower, r)
	}
	return powers
}

func PolynomialToKZGCommitment(eval Polynomial) KZGCommitment {
	g1 := bls.LinCombG1(kzgSetupLagrange, []bls.Fr(eval))
	var out KZGCommitment
	copy(out[:], bls.ToCompressedG1(g1))
	return out
}

// bytesToBLSField implements bytes_to_bls_field from the EIP-4844 consensus spec:
// https://github.com/ethereum/consensus-specs/blob/dev/specs/eip4844/polynomial-commitments.md#bytes_to_bls_field
func bytesToBLSField(element *bls.Fr, bytes32 [32]byte) bool {
	return bls.FrFrom32(element, bytes32)
}

// hashToBytesField implements hash_to_bls_field from the EIP-4844 consensus spec:
// https://github.com/ethereum/consensus-specs/blob/dev/specs/eip4844/polynomial-commitments.md#hash_to_bls_field
func hashToBLSField(input []byte) *bls.Fr {
	// First hash the input -- For the fiat-shamir protocol this
	// will be the compressed state with a challenge index appended
	// to it.

	sha := sha256.New()
	sha.Write(input[:])

	var hash32 [32]byte
	copy(hash32[:], sha.Sum(nil))

	// Then interpret the hash digest as a little-endian integer
	// modulo the bls field modulus
	reverseArr32(&hash32)
	zB := new(big.Int).Mod(new(big.Int).SetBytes(hash32[:]), BLSModulus)

	// Convert the big integer into a field element.
	out := new(bls.Fr)
	bigToFr(out, zB)
	return out
}

// ComputeAggregatedPolyAndCommitment implements compute_aggregated_poly_and_commitment from the EIP-4844 consensus spec:
// https://github.com/ethereum/consensus-specs/blob/dev/specs/eip4844/polynomial-commitments.md#compute_aggregated_poly_and_commitment
func ComputeAggregatedPolyAndCommitment(blobs Polynomials, commitments KZGCommitmentSequence) ([]bls.Fr, *bls.G1Point, *bls.Fr, error) {
	// create challenges
	powers, evaluationChallenge, err := ComputeChallenges(blobs, commitments)
	if err != nil {
		return nil, nil, nil, err
	}

	aggregatedPoly, err := bls.PolyLinComb(blobs, powers, FieldElementsPerBlob)
	if err != nil {
		return nil, nil, nil, err
	}

	l := commitments.Len()
	commitmentsG1 := make([]bls.G1Point, l)
	for i := 0; i < l; i++ {
		c := commitments.At(i)
		p, err := bls.FromCompressedG1(c[:])
		if err != nil {
			return nil, nil, nil, err
		}
		bls.CopyG1(&commitmentsG1[i], p)
	}
	aggregatedCommitmentG1 := bls.LinCombG1(commitmentsG1, powers)
	return aggregatedPoly, aggregatedCommitmentG1, evaluationChallenge, nil
}

// ComputeAggregateKZGProofFromPolynomials implements compute_aggregate_kzg_proof from the EIP-4844
// consensus spec, only operating over blobs that are already parsed into a polynomial.
func ComputeAggregateKZGProofFromPolynomials(blobs Polynomials) (KZGProof, error) {
	commitments := make(KZGCommitmentSequenceImpl, len(blobs))
	for i, b := range blobs {
		commitments[i] = PolynomialToKZGCommitment(Polynomial(b))
	}
	aggregatedPoly, _, evaluationChallenge, err := ComputeAggregatedPolyAndCommitment(blobs, commitments)
	if err != nil {
		return KZGProof{}, err
	}
	return ComputeKZGProof(aggregatedPoly, evaluationChallenge)
}

// ComputeKZGProof implements compute_kzg_proof from the EIP-4844 consensus spec:
// https://github.com/ethereum/consensus-specs/blob/dev/specs/eip4844/polynomial-commitments.md#compute_kzg_proof
func ComputeKZGProof(polynomial []bls.Fr, z *bls.Fr) (KZGProof, error) {
	y := EvaluatePolynomialInEvaluationForm(polynomial, z)
	polynomialShifted := make([]bls.Fr, len(polynomial))
	for i := range polynomial {
		bls.SubModFr(&polynomialShifted[i], &polynomial[i], y)
	}
	denominatorPoly := make([]bls.Fr, len(polynomial))
	if len(polynomial) != len(DomainFr) {
		return KZGProof{}, errors.New("polynomial has invalid length")
	}
	for i := range polynomial {
		if bls.EqualFr(&DomainFr[i], z) {
			return KZGProof{}, errors.New("invalid z challenge")
		}
		bls.SubModFr(&denominatorPoly[i], &DomainFr[i], z)
	}
	quotientPolynomial := make([]bls.Fr, len(polynomial))
	for i := range polynomial {
		bls.DivModFr(&quotientPolynomial[i], &polynomialShifted[i], &denominatorPoly[i])
	}
	rG1 := bls.LinCombG1(kzgSetupLagrange, quotientPolynomial)
	var proof KZGProof
	copy(proof[:], bls.ToCompressedG1(rG1))
	return proof, nil
}

// EvaluatePolynomialInEvaluationForm implements evaluate_polynomial_in_evaluation_form from the EIP-4844 consensus spec:
// https://github.com/ethereum/consensus-specs/blob/dev/specs/eip4844/polynomial-commitments.md#evaluate_polynomial_in_evaluation_form
func EvaluatePolynomialInEvaluationForm(poly []bls.Fr, x *bls.Fr) *bls.Fr {
	var result bls.Fr
	bls.EvaluatePolyInEvaluationForm(&result, poly, x, DomainFr, 0)
	return &result
}

// ComputeChallenges implements compute_challenges from the EIP-4844 consensus spec:
// https://github.com/ethereum/consensus-specs/blob/dev/specs/eip4844/polynomial-commitments.md#compute_challenges
func ComputeChallenges(polys Polynomials, comms KZGCommitmentSequence) ([]bls.Fr, *bls.Fr, error) {
	hash, err := hashPolysComms(polys, comms)
	if err != nil {
		return nil, nil, err
	}

	var linCombChallengeTranscript = append(hash[:], 0)
	var evalChallengeTranscript = append(hash[:], 1)

	linCombChallenge := hashToBLSField(linCombChallengeTranscript)
	evalChallenge := hashToBLSField(evalChallengeTranscript)

	rPowers := ComputePowers(linCombChallenge, len(polys))

	return rPowers, evalChallenge, nil

}

// Adds the domain separator, polynomials and commitments into a buffer, returning the
// hash of this buffer
func hashPolysComms(polys Polynomials, comms KZGCommitmentSequence) ([32]byte, error) {
	sha := sha256.New()
	var hash [32]byte

	sha.Write([]byte(FIAT_SHAMIR_PROTOCOL_DOMAIN))

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(FieldElementsPerBlob))
	sha.Write(bytes)

	bytes = make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(len(polys)))
	sha.Write(bytes)

	for _, poly := range polys {
		for _, fe := range poly {
			b32 := bls.FrTo32(&fe)
			sha.Write(b32[:])
		}
	}
	l := comms.Len()
	for i := 0; i < l; i++ {
		c := comms.At(i)
		sha.Write(c[:])
	}
	copy(hash[:], sha.Sum(nil))
	return hash, nil
}

func BlobToPolynomial(b Blob) (Polynomial, bool) {
	l := b.Len()
	frs := make(Polynomial, l)
	for i := 0; i < l; i++ {
		if !bytesToBLSField(&frs[i], b.At(i)) {
			return []bls.Fr{}, false
		}
	}
	return frs, true
}

func BlobsToPolynomials(blobs BlobSequence) ([][]bls.Fr, bool) {
	l := blobs.Len()
	out := make(Polynomials, l)
	for i := 0; i < l; i++ {
		blob, ok := BlobToPolynomial(blobs.At(i))
		if !ok {
			return nil, false
		}
		out[i] = blob
	}
	return out, true
}

func bigToFr(out *bls.Fr, in *big.Int) bool {
	// Convert big.Int to a 32 byte array
	// Note that this function will panic if the
	// big Integer needs more than 32 bytes to represent
	// the integer.
	var b [32]byte
	inb := in.Bytes()
	copy(b[32-len(inb):], inb)

	// The byte array `b` is an integer in big endian format.
	// We therefore need to reverse it as go-kzg expects little-endian
	reverseArr32(&b)

	return bls.FrFrom32(out, b)
}

func reverseArr32(input *[32]byte) {
	for i := 0; i < 16; i++ {
		input[31-i], input[i] = input[i], input[31-i]
	}
}
