// +build bignum_kilic

package bls

import (
	"crypto/rand"
	"encoding/binary"
	kbls "github.com/kilic/bls12-381"
)

func init() {
	initGlobals()
	ClearG1(&ZERO_G1)
	initG1G2()
}

type Fr kbls.Fr

func SetFr(dst *Fr, v string) {
	var bv fr.Int
	bv.SetString(v, 10)
	(*kbls.Fr)(dst).FromBytes(bv.Bytes())
}

// FrFrom32 mutates the fr num. The value v is little-endian 32-bytes.
func FrFrom32(dst *Fr, v [32]byte) {
	// reverse endianness, Kilic Fr takes big-endian bytes
	for i := 0; i < 16; i++ {
		v[i], v[31-i] = v[31-i], v[i]
	}
	(*kbls.Fr)(dst).FromBytes(v[:])
}

// FrTo32 serializes a fr number to 32 bytes. Encoded little-endian.
func FrTo32(src *Fr) (v [32]byte) {
	b := (*kbls.Fr)(src).ToBytes()
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
	binary.FrEndian.PutUint64(data[:], i)
	(*kbls.Fr)(dst).FromBytes(data[:])
}

func FrStr(b *Fr) string {
	if b == nil {
		return "<nil>"
	}
	return (*kbls.Fr)(b).ToFr().String()
}

func EqualOne(v *Fr) bool {
	return (*kbls.Fr)(v).IsOne()
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
	tmp.Inverse((*kbls.Fr)(b))
	(*kbls.Fr)(dst).Mul(&tmp, (*kbls.Fr)(a))
}

func MulModFr(dst *Fr, a, b *Fr) {
	(*kbls.Fr)(dst).Mul((*kbls.Fr)(a), (*kbls.Fr)(b))
}

func InvModFr(dst *Fr, v *Fr) {
	(*kbls.Fr)(dst).Inverse((*kbls.Fr)(v))
}

//func SqrModFr(dst *Fr, v *Fr) {
//	kbls.FrSqr((*kbls.Fr)(dst), (*kbls.Fr)(v))
//}
