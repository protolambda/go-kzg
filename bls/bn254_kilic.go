// +build bn254

package bls

import (
	"fmt"
	kbn254 "github.com/kilic/bn254"
	"math/big"
	"strings"
)

var ZERO_G1 G1Point

var GenG1 G1Point
var GenG2 G2Point

var ZeroG1 G1Point
var ZeroG2 G2Point

// Herumi BLS doesn't offer these points to us, so we have to work around it by declaring them ourselves.
func initG1G2() {
	GenG1 = G1Point(*kbn254.NewG1().One())
	GenG2 = G2Point(*kbn254.NewG2().One())
	ZeroG1 = G1Point(*kbn254.NewG1().Zero())
	ZeroG2 = G2Point(*kbn254.NewG2().Zero())
}

// TODO types file, swap BLS with build args
type G1Point kbn254.PointG1

// zeroes the point (like herumi BLS does with theirs). This is not co-factor clearing.
func ClearG1(x *G1Point) {
	(*kbn254.PointG1)(x).Zero()
}

func CopyG1(dst *G1Point, v *G1Point) {
	*dst = *v
}

func MulG1(dst *G1Point, a *G1Point, b *Fr) {
	kbn254.NewG1().MulScalar((*kbn254.PointG1)(dst), (*kbn254.PointG1)(a), (*big.Int)(b))
}

func AddG1(dst *G1Point, a *G1Point, b *G1Point) {
	kbn254.NewG1().Add((*kbn254.PointG1)(dst), (*kbn254.PointG1)(a), (*kbn254.PointG1)(b))
}

func SubG1(dst *G1Point, a *G1Point, b *G1Point) {
	kbn254.NewG1().Sub((*kbn254.PointG1)(dst), (*kbn254.PointG1)(a), (*kbn254.PointG1)(b))
}

func StrG1(v *G1Point) string {
	data := kbn254.NewG1().ToBytes((*kbn254.PointG1)(v))
	var a, b big.Int
	a.SetBytes(data[:32])
	b.SetBytes(data[32:])
	return a.String() + "\n" + b.String()
}

func NegG1(dst *G1Point) {
	// in-place should be safe here (TODO double check)
	kbn254.NewG1().Neg((*kbn254.PointG1)(dst), (*kbn254.PointG1)(dst))
}

type G2Point kbn254.PointG2

// zeroes the point (like herumi BLS does with theirs). This is not co-factor clearing.
func ClearG2(x *G2Point) {
	(*kbn254.PointG2)(x).Zero()
}

func CopyG2(dst *G2Point, v *G2Point) {
	*dst = *v
}

func MulG2(dst *G2Point, a *G2Point, b *Fr) {
	kbn254.NewG2().MulScalar((*kbn254.PointG2)(dst), (*kbn254.PointG2)(a), (*big.Int)(b))
}

func AddG2(dst *G2Point, a *G2Point, b *G2Point) {
	kbn254.NewG2().Add((*kbn254.PointG2)(dst), (*kbn254.PointG2)(a), (*kbn254.PointG2)(b))
}

func SubG2(dst *G2Point, a *G2Point, b *G2Point) {
	kbn254.NewG2().Sub((*kbn254.PointG2)(dst), (*kbn254.PointG2)(a), (*kbn254.PointG2)(b))
}

func NegG2(dst *G2Point) {
	// in-place should be safe here (TODO double check)
	kbn254.NewG2().Neg((*kbn254.PointG2)(dst), (*kbn254.PointG2)(dst))
}

func StrG2(v *G2Point) string {
	data := kbn254.NewG2().ToBytes((*kbn254.PointG2)(v))
	var a, b big.Int
	a.SetBytes(data[:64])
	b.SetBytes(data[64:])
	return a.String() + "\n" + b.String()
}

func EqualG1(a *G1Point, b *G1Point) bool {
	return kbn254.NewG1().Equal((*kbn254.PointG1)(a), (*kbn254.PointG1)(b))
}

func EqualG2(a *G2Point, b *G2Point) bool {
	return kbn254.NewG2().Equal((*kbn254.PointG2)(a), (*kbn254.PointG2)(b))
}

func ToCompressedG1(p *G1Point) []byte {
	// NOTE this is actually the uncompressed form
	return kbn254.NewG1().ToBytes((*kbn254.PointG1)(p))
}

func LinCombG1(numbers []G1Point, factors []Fr) *G1Point {
	if len(numbers) != len(factors) {
		panic("got LinCombG1 numbers/factors length mismatch")
	}
	var out G1Point
	tmpG1s := make([]*kbn254.PointG1, len(numbers), len(numbers))
	for i := 0; i < len(numbers); i++ {
		tmpG1s[i] = (*kbn254.PointG1)(&numbers[i])
	}
	tmpFrs := make([]*big.Int, len(factors), len(factors))
	for i := 0; i < len(factors); i++ {
		tmpFrs[i] = (*big.Int)(&factors[i])
	}
	_, _ = kbn254.NewG1().MultiExp((*kbn254.PointG1)(&out), tmpG1s, tmpFrs)
	return &out
}

// e(a1^(-1), a2) * e(b1,  b2) = 1_T
func PairingsVerify(a1 *G1Point, a2 *G2Point, b1 *G1Point, b2 *G2Point) bool {
	pairingEngine := kbn254.NewEngine()
	pairingEngine.AddPairInv((*kbn254.PointG1)(a1), (*kbn254.PointG2)(a2))
	pairingEngine.AddPair((*kbn254.PointG1)(b1), (*kbn254.PointG2)(b2))
	return pairingEngine.Check()
}

func DebugG1s(msg string, values []G1Point) {
	var out strings.Builder
	for i := range values {
		out.WriteString(fmt.Sprintf("%s %d: %s\n", msg, i, StrG1(&values[i])))
	}
	fmt.Println(out.String())
}
