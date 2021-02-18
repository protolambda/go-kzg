// +build !bignum_pure,!bignum_hol256,!bignum_kilic

package bls

import (
	"fmt"
	hbls "github.com/herumi/bls-eth-go-binary/bls"
	"strings"
	"unsafe"
)

var ZERO_G1 G1

var GenG1 G1
var GenG2 G2

var ZeroG1 G1
var ZeroG2 G2

// Herumi BLS doesn't offer these points to us, so we have to work around it by declaring them ourselves.
func initG1G2() {
	GenG1.X.SetString("3685416753713387016781088315183077757961620795782546409894578378688607592378376318836054947676345821548104185464507", 10)
	GenG1.Y.SetString("1339506544944476473020471379941921221584933875938349620426543736416511423956333506472724655353366534992391756441569", 10)
	GenG1.Z.SetInt64(1)

	GenG2.X.D[0].SetString("352701069587466618187139116011060144890029952792775240219908644239793785735715026873347600343865175952761926303160", 10)
	GenG2.X.D[1].SetString("3059144344244213709971259814753781636986470325476647558659373206291635324768958432433509563104347017837885763365758", 10)
	GenG2.Y.D[0].SetString("1985150602287291935568054521177171638300868978215655730859378665066344726373823718423869104263333984641494340347905", 10)
	GenG2.Y.D[1].SetString("927553665492332455747201965776037880757740193453592970025027978793976877002675564980949289727957565575433344219582", 10)
	GenG2.Z.D[0].SetInt64(1)
	GenG2.Z.D[1].Clear()

	ZeroG1.X.SetInt64(1)
	ZeroG1.Y.SetInt64(1)
	ZeroG1.Z.SetInt64(0)

	ZeroG2.X.D[0].SetInt64(1)
	ZeroG2.X.D[1].SetInt64(0)
	ZeroG2.Y.D[0].SetInt64(1)
	ZeroG2.Y.D[1].SetInt64(0)
	ZeroG2.Z.D[0].SetInt64(0)
	ZeroG2.Z.D[1].SetInt64(0)
}

// TODO types file, swap BLS with build args
type G1 hbls.G1

func ClearG1(x *G1) {
	(*hbls.G1)(x).Clear()
}

func CopyG1(dst *G1, v *G1) {
	*dst = *v
}

func MulG1(dst *G1, a *G1, b *Big) {
	hbls.G1Mul((*hbls.G1)(dst), (*hbls.G1)(a), (*hbls.Fr)(b))
}

func AddG1(dst *G1, a *G1, b *G1) {
	hbls.G1Add((*hbls.G1)(dst), (*hbls.G1)(a), (*hbls.G1)(b))
}

func SubG1(dst *G1, a *G1, b *G1) {
	hbls.G1Sub((*hbls.G1)(dst), (*hbls.G1)(a), (*hbls.G1)(b))
}

func StrG1(v *G1) string {
	return (*hbls.G1)(v).GetString(10)
}

func NegG1(dst *G1) {
	// in-place should be safe here (TODO double check)
	hbls.G1Neg((*hbls.G1)(dst), (*hbls.G1)(dst))
}

type G2 hbls.G2

func ClearG2(x *G2) {
	(*hbls.G2)(x).Clear()
}

func CopyG2(dst *G2, v *G2) {
	*dst = *v
}

func MulG2(dst *G2, a *G2, b *Big) {
	hbls.G2Mul((*hbls.G2)(dst), (*hbls.G2)(a), (*hbls.Fr)(b))
}

func AddG2(dst *G2, a *G2, b *G2) {
	hbls.G2Add((*hbls.G2)(dst), (*hbls.G2)(a), (*hbls.G2)(b))
}

func SubG2(dst *G2, a *G2, b *G2) {
	hbls.G2Sub((*hbls.G2)(dst), (*hbls.G2)(a), (*hbls.G2)(b))
}

func NegG2(dst *G2) {
	// in-place should be safe here (TODO double check)
	hbls.G2Neg((*hbls.G2)(dst), (*hbls.G2)(dst))
}

func StrG2(v *G2) string {
	return (*hbls.G2)(v).GetString(10)
}

func EqualG1(a *G1, b *G1) bool {
	return (*hbls.G1)(a).IsEqual((*hbls.G1)(b))
}

func EqualG2(a *G2, b *G2) bool {
	return (*hbls.G2)(a).IsEqual((*hbls.G2)(b))
}

func LinCombG1(numbers []G1, factors []Big) *G1 {
	var out G1
	// We're just using unsafe to cast elements that are an alias anyway, no problem.
	// Go doesn't let us do the cast otherwise without copy.
	hbls.G1MulVec((*hbls.G1)(&out), *(*[]hbls.G1)(unsafe.Pointer(&numbers)), *(*[]hbls.Fr)(unsafe.Pointer(&factors)))
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
	if err := hbls.FrEvaluatePolynomial(
		(*hbls.Fr)(dst),
		*(*[]hbls.Fr)(unsafe.Pointer(&p)),
		(*hbls.Fr)(x),
	); err != nil {
		panic(err) // TODO: why does the herumi API return an error? When coefficients are empty?
	}
}

// e(a1^(-1), a2) * e(b1,  b2) = 1_T
func PairingsVerify(a1 *G1, a2 *G2, b1 *G1, b2 *G2) bool {
	var tmp hbls.GT
	hbls.Pairing(&tmp, (*hbls.G1)(a1), (*hbls.G2)(a2))
	//fmt.Println("tmp", tmp.GetString(10))
	var tmp2 hbls.GT
	hbls.Pairing(&tmp2, (*hbls.G1)(b1), (*hbls.G2)(b2))

	// invert left pairing
	var tmp3 hbls.GT
	hbls.GTInv(&tmp3, &tmp)

	// multiply the two
	var tmp4 hbls.GT
	hbls.GTMul(&tmp4, &tmp3, &tmp2)

	// final exp.
	var tmp5 hbls.GT
	hbls.FinalExp(&tmp5, &tmp4)

	// = 1_T
	return tmp5.IsOne()

	// TODO, alternatively use the equal check (faster or slower?):
	////fmt.Println("tmp2", tmp2.GetString(10))
	//return tmp.IsEqual(&tmp2)
}

func DebugG1s(msg string, values []G1) {
	var out strings.Builder
	for i := range values {
		out.WriteString(fmt.Sprintf("%s %d: %s\n", msg, i, StrG1(&values[i])))
	}
	fmt.Println(out.String())
}
