// +build bignum_blst

package bls

import (
	"fmt"
	blst "github.com/supranational/blst/bindings/go"
	"math/big"
	"strings"
)

var ZERO_G1 G1Point

var GenG1 G1Point
var GenG2 G2Point

var ZeroG1 G1Point
var ZeroG2 G2Point

func initG1G2() {
	// TODO
	//GenG1 = G1Point(*curveG1.One())
	//GenG2 = G2Point(*curveG2.One())
	//ZeroG1 = G1Point(*curveG1.Zero())
	//ZeroG2 = G2Point(*curveG2.Zero())
}

type G1Point blst.P1

// zeroes the point (like herumi BLS does with theirs). This is not co-factor clearing.
func ClearG1(x *G1Point) {
	*x = ZeroG1
}

func CopyG1(dst *G1Point, v *G1Point) {
	*dst = *v
}

func MulG1(dst *G1Point, a *G1Point, b *Fr) {
	var tmp blst.Scalar
	tmp.FromFr((*blst.Fr)(b))
	// TODO: not always 255 bits? BLST has constant-time multiplication here, not fast anyway
	(*blst.P1)(dst).Mult((*blst.P1)(a), &tmp, 255)
}

func AddG1(dst *G1Point, a *G1Point, b *G1Point) {
	(*blst.P1)(dst).Add((*blst.P1)(a), (*blst.P1)(b))
}

func SubG1(dst *G1Point, a *G1Point, b *G1Point) {
	tmp := *(*blst.P1)(b)
	tmp.Negative()
	(*blst.P1)(dst).Add((*blst.P1)(a), &tmp)
}

func StrG1(v *G1Point) string {
	data := (*blst.P1)(v).Serialize()
	var a, b big.Int
	a.SetBytes(data[:48])
	b.SetBytes(data[48:])
	return a.String() + "\n" + b.String()
}

func NegG1(dst *G1Point) {
	(*blst.P1)(dst).Negative()
}

type G2Point blst.P2

// zeroes the point (like herumi BLS does with theirs). This is not co-factor clearing.
func ClearG2(x *G2Point) {
	*x = ZeroG2
}

func CopyG2(dst *G2Point, v *G2Point) {
	*dst = *v
}

func MulG2(dst *G2Point, a *G2Point, b *Fr) {
	var tmp blst.Scalar
	tmp.FromFr((*blst.Fr)(b))
	// TODO: not always 255 bits? BLST has constant-time multiplication here, not fast anyway
	(*blst.P2)(dst).Mult((*blst.P2)(a), &tmp, 255)
}

func AddG2(dst *G2Point, a *G2Point, b *G2Point) {
	(*blst.P2)(dst).Add((*blst.P2)(a), (*blst.P2)(b))
}

func SubG2(dst *G2Point, a *G2Point, b *G2Point) {
	tmp := *(*blst.P2)(b)
	tmp.Negative()
	(*blst.P2)(dst).Add((*blst.P2)(a), &tmp)
}

func NegG2(dst *G2Point) {
	(*blst.P2)(dst).Negative()
}

func StrG2(v *G2Point) string {
	data := (*blst.P2)(v).Serialize()
	var a, b big.Int
	a.SetBytes(data[:96])
	b.SetBytes(data[96:])
	return a.String() + "\n" + b.String()
}

func EqualG1(a *G1Point, b *G1Point) bool {
	return (*blst.P1)(a).Equals((*blst.P1)(b))
}

func EqualG2(a *G2Point, b *G2Point) bool {
	return (*blst.P2)(a).Equals((*blst.P2)(b))
}

func ToCompressedG1(p *G1Point) []byte {
	return (*blst.P1)(p).Compress()
}

func LinCombG1(numbers []G1Point, factors []Fr) *G1Point {
	var tmp G1Point
	var out G1Point
	CopyG1(&out, &ZERO_G1)
	for i := 0; i < len(numbers); i++ {
		MulG1(&tmp, &numbers[i], &factors[i])
		AddG1(&out, &out, &tmp)
	}
	return &out
}

// e(a1^(-1), a2) * e(b1,  b2) = 1_T
func PairingsVerify(a1 *G1Point, a2 *G2Point, b1 *G1Point, b2 *G2Point) bool {
	return blst.PairingsVerify((*blst.P1)(a1), (*blst.P2)(a2), (*blst.P1)(b1), (*blst.P2)(b2))
}

func DebugG1s(msg string, values []G1Point) {
	var out strings.Builder
	for i := range values {
		out.WriteString(fmt.Sprintf("%s %d: %s\n", msg, i, StrG1(&values[i])))
	}
	fmt.Println(out.String())
}
