// +build !bignum_pure,!bignum_hol256,!bignum_kilic,!bignum_blst

package bls

import (
	"fmt"
	hbls "github.com/herumi/bls-eth-go-binary/bls"
	"strings"
	"unsafe"
)

// TODO duplicate
var ZERO_G1 G1Point

var GenG1 G1Point
var GenG2 G2Point

var ZeroG1 G1Point
var ZeroG2 G2Point

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

type G1Point hbls.G1

func ClearG1(x *G1Point) {
	(*hbls.G1)(x).Clear()
}

func CopyG1(dst *G1Point, v *G1Point) {
	*dst = *v
}

func MulG1(dst *G1Point, a *G1Point, b *Fr) {
	hbls.G1Mul((*hbls.G1)(dst), (*hbls.G1)(a), (*hbls.Fr)(b))
}

func AddG1(dst *G1Point, a *G1Point, b *G1Point) {
	hbls.G1Add((*hbls.G1)(dst), (*hbls.G1)(a), (*hbls.G1)(b))
}

func SubG1(dst *G1Point, a *G1Point, b *G1Point) {
	hbls.G1Sub((*hbls.G1)(dst), (*hbls.G1)(a), (*hbls.G1)(b))
}

func StrG1(v *G1Point) string {
	return (*hbls.G1)(v).GetString(10)
}

func NegG1(dst *G1Point) {
	// in-place should be safe here (TODO double check)
	hbls.G1Neg((*hbls.G1)(dst), (*hbls.G1)(dst))
}

type G2Point hbls.G2

func ClearG2(x *G2Point) {
	(*hbls.G2)(x).Clear()
}

func CopyG2(dst *G2Point, v *G2Point) {
	*dst = *v
}

func MulG2(dst *G2Point, a *G2Point, b *Fr) {
	hbls.G2Mul((*hbls.G2)(dst), (*hbls.G2)(a), (*hbls.Fr)(b))
}

func AddG2(dst *G2Point, a *G2Point, b *G2Point) {
	hbls.G2Add((*hbls.G2)(dst), (*hbls.G2)(a), (*hbls.G2)(b))
}

func SubG2(dst *G2Point, a *G2Point, b *G2Point) {
	hbls.G2Sub((*hbls.G2)(dst), (*hbls.G2)(a), (*hbls.G2)(b))
}

func NegG2(dst *G2Point) {
	// in-place should be safe here (TODO double check)
	hbls.G2Neg((*hbls.G2)(dst), (*hbls.G2)(dst))
}

func StrG2(v *G2Point) string {
	return (*hbls.G2)(v).GetString(10)
}

func EqualG1(a *G1Point, b *G1Point) bool {
	return (*hbls.G1)(a).IsEqual((*hbls.G1)(b))
}

func EqualG2(a *G2Point, b *G2Point) bool {
	return (*hbls.G2)(a).IsEqual((*hbls.G2)(b))
}

func ToCompressedG1(p *G1Point) []byte {
	return hbls.CastToPublicKey((*hbls.G1)(p)).Serialize()
}

func ToCompressedG2(p *G2Point) []byte {
	return hbls.CastToSign((*hbls.G2)(p)).Serialize()
}

func LinCombG1(numbers []G1Point, factors []Fr) *G1Point {
	var out G1Point
	// We're just using unsafe to cast elements that are an alias anyway, no problem.
	// Go doesn't let us do the cast otherwise without copy.
	hbls.G1MulVec((*hbls.G1)(&out), *(*[]hbls.G1)(unsafe.Pointer(&numbers)), *(*[]hbls.Fr)(unsafe.Pointer(&factors)))
	return &out
}

// e(a1^(-1), a2) * e(b1,  b2) = 1_T
func PairingsVerify(a1 *G1Point, a2 *G2Point, b1 *G1Point, b2 *G2Point) bool {
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

func DebugG1s(msg string, values []G1Point) {
	var out strings.Builder
	for i := range values {
		out.WriteString(fmt.Sprintf("%s %d: %s\n", msg, i, StrG1(&values[i])))
	}
	fmt.Println(out.String())
}
