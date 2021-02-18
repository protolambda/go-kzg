// +build bignum_kilic

package bls

import (
	"fmt"
	kbls "github.com/kilic/bls12-381"
	"math/big"
	"strings"
)

var ZERO_G1 G1

var curveG1 kbls.G1
var curveG2 kbls.G2

var GenG1 G1
var GenG2 G2

var ZeroG1 G1
var ZeroG2 G2

// Herumi BLS doesn't offer these points to us, so we have to work around it by declaring them ourselves.
func initG1G2() {
	curveG1 = *kbls.NewG1()
	curveG2 = *kbls.NewG2()
	GenG1 = G1(*curveG1.One())
	GenG2 = G2(*curveG2.One())
	ZeroG1 = G1(*curveG1.Zero())
	ZeroG2 = G2(*curveG2.Zero())
}

// TODO types file, swap BLS with build args
type G1 kbls.PointG1

// zeroes the point (like herumi BLS does with theirs). This is not co-factor clearing.
func ClearG1(x *G1) {
	(*kbls.PointG1)(x).Zero()
}

func CopyG1(dst *G1, v *G1) {
	*dst = *v
}

func MulG1(dst *G1, a *G1, b *Big) {
	curveG1.MulScalar((*kbls.PointG1)(dst), (*kbls.PointG1)(a), (*kbls.Fr)(b))
}

func AddG1(dst *G1, a *G1, b *G1) {
	curveG1.Add((*kbls.PointG1)(dst), (*kbls.PointG1)(a), (*kbls.PointG1)(b))
}

func SubG1(dst *G1, a *G1, b *G1) {
	curveG1.Sub((*kbls.PointG1)(dst), (*kbls.PointG1)(a), (*kbls.PointG1)(b))
}

func StrG1(v *G1) string {
	data := curveG1.ToUncompressed((*kbls.PointG1)(v))
	var a, b big.Int
	a.SetBytes(data[:48])
	b.SetBytes(data[48:])
	return a.String() + "\n" + b.String()
}

func NegG1(dst *G1) {
	// in-place should be safe here (TODO double check)
	curveG1.Neg((*kbls.PointG1)(dst), (*kbls.PointG1)(dst))
}

type G2 kbls.PointG2

// zeroes the point (like herumi BLS does with theirs). This is not co-factor clearing.
func ClearG2(x *G2) {
	(*kbls.PointG2)(x).Zero()
}

func CopyG2(dst *G2, v *G2) {
	*dst = *v
}

func MulG2(dst *G2, a *G2, b *Big) {
	curveG2.MulScalar((*kbls.PointG2)(dst), (*kbls.PointG2)(a), (*kbls.Fr)(b))
}

func AddG2(dst *G2, a *G2, b *G2) {
	curveG2.Add((*kbls.PointG2)(dst), (*kbls.PointG2)(a), (*kbls.PointG2)(b))
}

func SubG2(dst *G2, a *G2, b *G2) {
	curveG2.Sub((*kbls.PointG2)(dst), (*kbls.PointG2)(a), (*kbls.PointG2)(b))
}

func NegG2(dst *G2) {
	// in-place should be safe here (TODO double check)
	curveG2.Neg((*kbls.PointG2)(dst), (*kbls.PointG2)(dst))
}

func StrG2(v *G2) string {
	data := curveG2.ToUncompressed((*kbls.PointG2)(v))
	var a, b big.Int
	a.SetBytes(data[:96])
	b.SetBytes(data[96:])
	return a.String() + "\n" + b.String()
}

func EqualG1(a *G1, b *G1) bool {
	return curveG1.Equal((*kbls.PointG1)(a), (*kbls.PointG1)(b))
}

func EqualG2(a *G2, b *G2) bool {
	return curveG2.Equal((*kbls.PointG2)(a), (*kbls.PointG2)(b))
}

func LinCombG1(numbers []G1, factors []Big) *G1 {
	if len(numbers) != len(factors) {
		panic("got LinCombG1 numbers/factors length mismatch")
	}
	var out G1
	tmpG1s := make([]*kbls.PointG1, len(numbers), len(numbers))
	for i := 0; i < len(numbers); i++ {
		tmpG1s[i] = (*kbls.PointG1)(&numbers[i])
	}
	tmpFrs := make([]*kbls.Fr, len(factors), len(factors))
	for i := 0; i < len(factors); i++ {
		tmpFrs[i] = (*kbls.Fr)(&factors[i])
	}
	_, _ = curveG1.MultiExp((*kbls.PointG1)(&out), tmpG1s, tmpFrs)
	return &out
}

func EvalPolyAtUnoptimized(dst *Big, coeffs []Big, x *Big) {
	if len(coeffs) == 0 {
		CopyBigNum(dst, &ZERO)
		return
	}
	if EqualZero(x) {
		CopyBigNum(dst, &coeffs[0])
		return
	}
	// Horner's method: work backwards, avoid doing more than N multiplications
	// https://en.wikipedia.org/wiki/Horner%27s_method
	var last Big
	CopyBigNum(&last, &coeffs[len(coeffs)-1])
	var tmp Big
	for i := len(coeffs) - 2; i >= 0; i-- {
		MulModBig(&tmp, &last, x)
		AddModBig(&last, &tmp, &coeffs[i])
	}
	CopyBigNum(dst, &last)
}

func EvalPolyAt(dst *Big, p []Big, x *Big) {
	// TODO: kilic BLS has no optimized evaluation function
	EvalPolyAtUnoptimized(dst, p, x)
}

// e(a1^(-1), a2) * e(b1,  b2) = 1_T
func PairingsVerify(a1 *G1, a2 *G2, b1 *G1, b2 *G2) bool {
	pairingEngine := kbls.NewEngine()
	pairingEngine.AddPairInv((*kbls.PointG1)(a1), (*kbls.PointG2)(a2))
	pairingEngine.AddPair((*kbls.PointG1)(b1), (*kbls.PointG2)(b2))
	return pairingEngine.Check()
}

func DebugG1s(msg string, values []G1) {
	var out strings.Builder
	for i := range values {
		out.WriteString(fmt.Sprintf("%s %d: %s\n", msg, i, StrG1(&values[i])))
	}
	fmt.Println(out.String())
}
