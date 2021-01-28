// +build bignum_pure

package kzg

import (
	"crypto/rand"
	"math/big"
)

var _modulus big.Int

func init() {
	bigNum((*Big)(&_modulus), "52435875175126190479447740508185965837690552500527637822603658699938581184513")
	initGlobals()
}

type Big big.Int

// BigNumFrom32 mutates the big num. The value v is little-endian 32-bytes.
func BigNumFrom32(dst *Big, v [32]byte) {
	// reverse endianness, big.Int takes big-endian bytes
	for i := 0; i < 16; i++ {
		v[i], v[31-i] = v[31-i], v[i]
	}
	(*big.Int)(dst).SetBytes(v[:])
}

// BigNumTo32 serializes a big number to 32 bytes. Encoded little-endian.
func BigNumTo32(src *Big) (v [32]byte) {
	b := (*big.Int)(src).Bytes()
	last := len(b) - 1
	// reverse endianness, u256.Int outputs big-endian bytes
	for i := 0; i < 16; i++ {
		b[i], b[last-i] = b[last-i], b[i]
	}
	copy(v[:], b)
	return
}

func CopyBigNum(dst *Big, v *Big) {
	(*big.Int)(dst).Set((*big.Int)(v))
}

func bigNum(dst *Big, v string) {
	if err := (*big.Int)(dst).UnmarshalText([]byte(v)); err != nil {
		panic(err)
	}
}

func asBig(dst *Big, i uint64) {
	(*big.Int)(dst).SetUint64(i)
}

func bigStr(b *Big) string {
	return (*big.Int)(b).String()
}

func equalOne(v *Big) bool {
	return (*big.Int)(v).Cmp((*big.Int)(&ONE)) == 0
}

func equalZero(v *Big) bool {
	return (*big.Int)(v).Cmp((*big.Int)(&ZERO)) == 0
}

func equalBig(a *Big, b *Big) bool {
	return (*big.Int)(a).Cmp((*big.Int)(b)) == 0
}

func randomBig() *Big {
	v, err := rand.Int(rand.Reader, &_modulus)
	if err != nil {
		panic(err)
	}
	return (*Big)(v)
}

func subModBig(dst *Big, a, b *Big) {
	(*big.Int)(dst).Sub((*big.Int)(a), (*big.Int)(b))
	(*big.Int)(dst).Mod((*big.Int)(dst), &_modulus)
}

func addModBig(dst *Big, a, b *Big) {
	(*big.Int)(dst).Add((*big.Int)(a), (*big.Int)(b))
	(*big.Int)(dst).Mod((*big.Int)(dst), &_modulus)
}

func divModBig(dst *Big, a, b *Big) {
	(*big.Int)(dst).DivMod((*big.Int)(a), (*big.Int)(b))
}

func mulModBig(dst *Big, a, b *Big) {
	(*big.Int)(dst).Mul((*big.Int)(a), (*big.Int)(b))
	(*big.Int)(dst).Mod((*big.Int)(dst), &_modulus)
}

func powModBig(dst *Big, a, b *Big) {
	(*big.Int)(dst).Exp((*big.Int)(a), (*big.Int)(b), &_modulus)
}

func invModBig(dst *Big, v *Big) {
	(*big.Int)(dst).ModInverse((*big.Int)(v), &_modulus)
}

//func sqrModBig(dst *Big, v *Big) {
//	(*big.Int)(dst).ModSqrt((*big.Int)(v), &_modulus)
//}
