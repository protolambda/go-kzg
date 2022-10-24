//go:build bignum_pure
// +build bignum_pure

package bls

import (
	"crypto/rand"
	"math/big"
)

var _modulus big.Int

func init() {
	initGlobals()
	CopyFr((*Fr)(&_modulus), &MODULUS)
}

type Fr big.Int

// FrFrom32 mutates the fr num. The value v is little-endian 32-bytes.
// Returns false, without modifying dst, if the value is out of range.
func FrFrom32(dst *Fr, v [32]byte) (ok bool) {
	if !ValidFr(v) {
		return false
	}
	// reverse endianness, big.Int takes big-endian bytes
	for i := 0; i < 16; i++ {
		v[i], v[31-i] = v[31-i], v[i]
	}
	(*big.Int)(dst).SetBytes(v[:])
	return true
}

// FrTo32 serializes a fr number to 32 bytes. Encoded little-endian.
func FrTo32(src *Fr) (v [32]byte) {
	b := (*big.Int)(src).Bytes()
	last := len(b) - 1
	// reverse endianness, u256.Int outputs big-endian bytes
	for i := 0; i < 16; i++ {
		b[i], b[last-i] = b[last-i], b[i]
	}
	copy(v[:], b)
	return
}

func CopyFr(dst *Fr, v *Fr) {
	(*big.Int)(dst).Set((*big.Int)(v))
}

func SetFr(dst *Fr, v string) {
	if err := (*big.Int)(dst).UnmarshalText([]byte(v)); err != nil {
		panic(err)
	}
}

func AsFr(dst *Fr, i uint64) {
	(*big.Int)(dst).SetUint64(i)
}

func FrStr(b *Fr) string {
	return (*big.Int)(b).String()
}

func EqualOne(v *Fr) bool {
	return (*big.Int)(v).Cmp((*big.Int)(&ONE)) == 0
}

func EqualZero(v *Fr) bool {
	return (*big.Int)(v).Cmp((*big.Int)(&ZERO)) == 0
}

func EqualFr(a *Fr, b *Fr) bool {
	return (*big.Int)(a).Cmp((*big.Int)(b)) == 0
}

func RandomFr() *Fr {
	v, err := rand.Int(rand.Reader, &_modulus)
	if err != nil {
		panic(err)
	}
	return (*Fr)(v)
}

func SubModFr(dst *Fr, a, b *Fr) {
	(*big.Int)(dst).Sub((*big.Int)(a), (*big.Int)(b))
	(*big.Int)(dst).Mod((*big.Int)(dst), &_modulus)
}

func AddModFr(dst *Fr, a, b *Fr) {
	(*big.Int)(dst).Add((*big.Int)(a), (*big.Int)(b))
	(*big.Int)(dst).Mod((*big.Int)(dst), &_modulus)
}

func DivModFr(dst *Fr, a, b *Fr) {
	var tmp Fr
	InvModFr(&tmp, b)
	(*big.Int)(dst).Mul((*big.Int)(a), (*big.Int)(&tmp))
	(*big.Int)(dst).Mod((*big.Int)(dst), &_modulus)
}

func MulModFr(dst *Fr, a, b *Fr) {
	(*big.Int)(dst).Mul((*big.Int)(a), (*big.Int)(b))
	(*big.Int)(dst).Mod((*big.Int)(dst), &_modulus)
}

func PowModFr(dst *Fr, a, b *Fr) {
	(*big.Int)(dst).Exp((*big.Int)(a), (*big.Int)(b), &_modulus)
}

func InvModFr(dst *Fr, v *Fr) {
	(*big.Int)(dst).ModInverse((*big.Int)(v), &_modulus)
}

// BatchInvModFr computes the inverse for each input.
// Warning: this does not actually batch, this is just here for compatibility with other BLS backends that do.
func BatchInvModFr(f []Fr) {
	for i := 0; i < len(f); i++ {
		InvModFr(&f[i], &f[i])
	}
}

//func sqrModFr(dst *Fr, v *Fr) {
//	(*big.Int)(dst).ModSqrt((*big.Int)(v), &_modulus)
//}

func EvalPolyAt(dst *Fr, p []Fr, x *Fr) {
	EvalPolyAtUnoptimized(dst, p, x)
}
