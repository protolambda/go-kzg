// +build bignum_pure

package bls

import (
	"crypto/rand"
)

var _modulus fr.Int

func init() {
	Fr((*Fr)(&_modulus), "52435875175126190479447740508185965837690552500527637822603658699938581184513")
	initGlobals()
}

type Fr fr.Int

// FrFrom32 mutates the fr num. The value v is little-endian 32-bytes.
func FrFrom32(dst *Fr, v [32]byte) {
	// reverse endianness, big.Int takes big-endian bytes
	for i := 0; i < 16; i++ {
		v[i], v[31-i] = v[31-i], v[i]
	}
	(*fr.Int)(dst).SetBytes(v[:])
}

// FrTo32 serializes a fr number to 32 bytes. Encoded little-endian.
func FrTo32(src *Fr) (v [32]byte) {
	b := (*fr.Int)(src).Bytes()
	last := len(b) - 1
	// reverse endianness, u256.Int outputs big-endian bytes
	for i := 0; i < 16; i++ {
		b[i], b[last-i] = b[last-i], b[i]
	}
	copy(v[:], b)
	return
}

func CopyFr(dst *Fr, v *Fr) {
	(*fr.Int)(dst).Set((*fr.Int)(v))
}

func Fr(dst *Fr, v string) {
	if err := (*fr.Int)(dst).UnmarshalText([]byte(v)); err != nil {
		panic(err)
	}
}

func AsFr(dst *Fr, i uint64) {
	(*fr.Int)(dst).SetUint64(i)
}

func FrStr(b *Fr) string {
	return (*fr.Int)(b).String()
}

func EqualOne(v *Fr) bool {
	return (*fr.Int)(v).Cmp((*fr.Int)(&ONE)) == 0
}

func EqualZero(v *Fr) bool {
	return (*fr.Int)(v).Cmp((*fr.Int)(&ZERO)) == 0
}

func EqualFr(a *Fr, b *Fr) bool {
	return (*fr.Int)(a).Cmp((*fr.Int)(b)) == 0
}

func RandomFr() *Fr {
	v, err := rand.Int(rand.Reader, &_modulus)
	if err != nil {
		panic(err)
	}
	return (*Fr)(v)
}

func SubModFr(dst *Fr, a, b *Fr) {
	(*fr.Int)(dst).Sub((*fr.Int)(a), (*fr.Int)(b))
	(*fr.Int)(dst).Mod((*fr.Int)(dst), &_modulus)
}

func AddModFr(dst *Fr, a, b *Fr) {
	(*fr.Int)(dst).Add((*fr.Int)(a), (*fr.Int)(b))
	(*fr.Int)(dst).Mod((*fr.Int)(dst), &_modulus)
}

func DivModFr(dst *Fr, a, b *Fr) {
	(*fr.Int)(dst).DivMod((*fr.Int)(a), (*fr.Int)(b))
}

func MulModFr(dst *Fr, a, b *Fr) {
	(*fr.Int)(dst).Mul((*fr.Int)(a), (*fr.Int)(b))
	(*fr.Int)(dst).Mod((*fr.Int)(dst), &_modulus)
}

func PowModFr(dst *Fr, a, b *Fr) {
	(*fr.Int)(dst).Exp((*fr.Int)(a), (*fr.Int)(b), &_modulus)
}

func InvModFr(dst *Fr, v *Fr) {
	(*fr.Int)(dst).ModInverse((*fr.Int)(v), &_modulus)
}

//func sqrModFr(dst *Fr, v *Fr) {
//	(*big.Int)(dst).ModSqrt((*big.Int)(v), &_modulus)
//}
