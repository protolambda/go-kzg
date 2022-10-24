//go:build bignum_hbls
// +build bignum_hbls

package bls

import (
	"math/big"
	"unsafe"

	hbls "github.com/herumi/bls-eth-go-binary/bls"
)

var _modulus big.Int

func init() {
	if err := _modulus.UnmarshalText([]byte("52435875175126190479447740508185965837690552500527637822603658699938581184513")); err != nil {
		panic(err)
	}
	hbls.Init(hbls.BLS12_381)
	initGlobals()
	ClearG1(&ZERO_G1)
	initG1G2()
}

type Fr hbls.Fr

func SetFr(dst *Fr, v string) {
	if err := (*hbls.Fr)(dst).SetString(v, 10); err != nil {
		panic(err)
	}
}

// FrFrom32 mutates the fr num. The value v is little-endian 32-bytes.
// Returns false, without modifying dst, if the value is out of range.
func FrFrom32(dst *Fr, v [32]byte) (ok bool) {
	if !ValidFr(v) {
		return false
	}
	(*hbls.Fr)(dst).SetLittleEndian(v[:])
	return true
}

// FrTo32 serializes a fr number to 32 bytes. Encoded little-endian.
func FrTo32(src *Fr) (v [32]byte) {
	b := (*hbls.Fr)(src).Serialize()
	last := len(b) - 1
	// reverse endianness, Herumi outputs big-endian bytes
	for i := 0; i < 16; i++ {
		b[i], b[last-i] = b[last-i], b[i]
	}
	copy(v[:], b)
	return
}

func CopyFr(dst *Fr, v *Fr) {
	*dst = *v
}

func AsFr(dst *Fr, i uint64) {
	(*hbls.Fr)(dst).SetInt64(int64(i))
}

func FrStr(b *Fr) string {
	if b == nil {
		return "<nil>"
	}
	return (*hbls.Fr)(b).GetString(10)
}

func EqualOne(v *Fr) bool {
	return (*hbls.Fr)(v).IsOne()
}

func EqualZero(v *Fr) bool {
	return (*hbls.Fr)(v).IsZero()
}

func EqualFr(a *Fr, b *Fr) bool {
	return (*hbls.Fr)(a).IsEqual((*hbls.Fr)(b))
}

func RandomFr() *Fr {
	var out hbls.Fr
	out.SetByCSPRNG()
	return (*Fr)(&out)
}

func SubModFr(dst *Fr, a, b *Fr) {
	hbls.FrSub((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func AddModFr(dst *Fr, a, b *Fr) {
	hbls.FrAdd((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func DivModFr(dst *Fr, a, b *Fr) {
	hbls.FrDiv((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func MulModFr(dst *Fr, a, b *Fr) {
	hbls.FrMul((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func InvModFr(dst *Fr, v *Fr) {
	hbls.FrInv((*hbls.Fr)(dst), (*hbls.Fr)(v))
}

// BatchInvModFr computes the inverse for each input.
// Warning: this does not actually batch, this is just here for compatibility with other BLS backends that do.
func BatchInvModFr(f []Fr) {
	for i := 0; i < len(f); i++ {
		hbls.FrInv((*hbls.Fr)(&f[i]), (*hbls.Fr)(&f[i]))
	}
}

//func SqrModFr(dst *Fr, v *Fr) {
//	hbls.FrSqr((*hbls.Fr)(dst), (*hbls.Fr)(v))
//}

func EvalPolyAt(dst *Fr, p []Fr, x *Fr) {
	if err := hbls.FrEvaluatePolynomial(
		(*hbls.Fr)(dst),
		*(*[]hbls.Fr)(unsafe.Pointer(&p)),
		(*hbls.Fr)(x),
	); err != nil {
		panic(err) // TODO: why does the herumi API return an error? When coefficients are empty?
	}
}

// ExpModFr computes v**e in Fr. Warning: this is a slow fallback on big int math.
func ExpModFr(dst *Fr, v *Fr, e *big.Int) {
	vBig, ok := new(big.Int).SetString(FrStr(v), 10)
	if !ok {
		panic("failed string hack")
	}
	res := new(big.Int).Exp(vBig, e, &_modulus)
	SetFr(dst, res.String())
}
