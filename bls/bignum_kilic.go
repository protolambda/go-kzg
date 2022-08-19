//go:build !bignum_pure && !bignum_hol256 && !bignum_hbls
// +build !bignum_pure,!bignum_hol256,!bignum_hbls

package bls

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"unsafe"

	kbls "github.com/kilic/bls12-381"
)

func init() {
	initGlobals()
	ClearG1(&ZERO_G1)
	initG1G2()
}

// Note: with Kilic BLS, we exclusively represent Fr in mont-red form.
// Whenever it is used with G1/G2, it needs to be normalized first.
type Fr kbls.Fr

func SetFr(dst *Fr, v string) {
	var bv big.Int
	bv.SetString(v, 10)
	(*kbls.Fr)(dst).RedFromBytes(bv.Bytes())
}

// FrFrom32 mutates the fr num. The value v is little-endian 32-bytes.
// Returns false, without modifying dst, if the value is out of range.
func FrFrom32(dst *Fr, v [32]byte) (ok bool) {
	if !ValidFr(v) {
		return false
	}
	// reverse endianness, Kilic Fr takes big-endian bytes
	for i := 0; i < 16; i++ {
		v[i], v[31-i] = v[31-i], v[i]
	}
	(*kbls.Fr)(dst).RedFromBytes(v[:])
	return true
}

// FrTo32 serializes a fr number to 32 bytes. Encoded little-endian.
func FrTo32(src *Fr) (v [32]byte) {
	b := (*kbls.Fr)(src).RedToBytes()
	last := len(b) - 1
	// reverse endianness, Kilic Fr outputs big-endian bytes
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
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], i)
	(*kbls.Fr)(dst).RedFromBytes(data[:])
}

func FrStr(b *Fr) string {
	if b == nil {
		return "<nil>"
	}
	return (*kbls.Fr)(b).RedToBig().String()
}

func EqualOne(v *Fr) bool {
	return (*kbls.Fr)(v).IsRedOne()
}

func EqualZero(v *Fr) bool {
	return (*kbls.Fr)(v).IsZero()
}

func EqualFr(a *Fr, b *Fr) bool {
	return (*kbls.Fr)(a).Equal((*kbls.Fr)(b))
}

func RandomFr() *Fr {
	var out kbls.Fr
	if _, err := out.Rand(rand.Reader); err != nil {
		panic(err)
	}
	out.ToRed()
	return (*Fr)(&out)
}

func SubModFr(dst *Fr, a, b *Fr) {
	(*kbls.Fr)(dst).Sub((*kbls.Fr)(a), (*kbls.Fr)(b))
}

func AddModFr(dst *Fr, a, b *Fr) {
	(*kbls.Fr)(dst).Add((*kbls.Fr)(a), (*kbls.Fr)(b))
}

func DivModFr(dst *Fr, a, b *Fr) {
	var tmp kbls.Fr
	tmp.RedInverse((*kbls.Fr)(b))
	(*kbls.Fr)(dst).RedMul(&tmp, (*kbls.Fr)(a))
}

func MulModFr(dst *Fr, a, b *Fr) {
	(*kbls.Fr)(dst).RedMul((*kbls.Fr)(a), (*kbls.Fr)(b))
}

func InvModFr(dst *Fr, v *Fr) {
	(*kbls.Fr)(dst).RedInverse((*kbls.Fr)(v))
}

func BatchInvModFr(f []Fr) {
	kbls.RedInverseBatchFr(*(*[]kbls.Fr)(unsafe.Pointer(&f)))
}

//func SqrModFr(dst *Fr, v *Fr) {
//	kbls.FrSqr((*kbls.Fr)(dst), (*kbls.Fr)(v))
//}

func EvalPolyAt(dst *Fr, p []Fr, x *Fr) {
	// TODO: kilic BLS has no optimized evaluation function
	EvalPolyAtUnoptimized(dst, p, x)
}
