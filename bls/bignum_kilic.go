// +build bignum_kilic

package bls

import (
	"crypto/rand"
	"encoding/binary"
	kbls "github.com/kilic/bls12-381"
	"math/big"
)

func init() {
	initGlobals()
	ClearG1(&ZERO_G1)
	initG1G2()
}

type Big kbls.Fr

func BigNum(dst *Big, v string) {
	var bv big.Int
	bv.SetString(v, 10)
	(*kbls.Fr)(dst).FromBytes(bv.Bytes())
}

// BigNumFrom32 mutates the big num. The value v is little-endian 32-bytes.
func BigNumFrom32(dst *Big, v [32]byte) {
	// reverse endianness, Kilic Fr takes big-endian bytes
	for i := 0; i < 16; i++ {
		v[i], v[31-i] = v[31-i], v[i]
	}
	(*kbls.Fr)(dst).FromBytes(v[:])
}

// BigNumTo32 serializes a big number to 32 bytes. Encoded little-endian.
func BigNumTo32(src *Big) (v [32]byte) {
	b := (*kbls.Fr)(src).ToBytes()
	last := len(b) - 1
	// reverse endianness, Kilic Fr outputs big-endian bytes
	for i := 0; i < 16; i++ {
		b[i], b[last-i] = b[last-i], b[i]
	}
	copy(v[:], b)
	return
}

func CopyBigNum(dst *Big, v *Big) {
	*dst = *v
}

func AsBig(dst *Big, i uint64) {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], i)
	(*kbls.Fr)(dst).FromBytes(data[:])
}

func BigStr(b *Big) string {
	if b == nil {
		return "<nil>"
	}
	return (*kbls.Fr)(b).ToBig().String()
}

func EqualOne(v *Big) bool {
	return (*kbls.Fr)(v).IsOne()
}

func EqualZero(v *Big) bool {
	return (*kbls.Fr)(v).IsZero()
}

func EqualBig(a *Big, b *Big) bool {
	return (*kbls.Fr)(a).Equal((*kbls.Fr)(b))
}

func RandomBig() *Big {
	var out kbls.Fr
	if _, err := out.Rand(rand.Reader); err != nil {
		panic(err)
	}
	return (*Big)(&out)
}

func SubModBig(dst *Big, a, b *Big) {
	(*kbls.Fr)(dst).Sub((*kbls.Fr)(a), (*kbls.Fr)(b))
}

func AddModBig(dst *Big, a, b *Big) {
	(*kbls.Fr)(dst).Add((*kbls.Fr)(a), (*kbls.Fr)(b))
}

func DivModBig(dst *Big, a, b *Big) {
	var tmp kbls.Fr
	tmp.Inverse((*kbls.Fr)(b))
	(*kbls.Fr)(dst).Mul(&tmp, (*kbls.Fr)(a))
}

func MulModBig(dst *Big, a, b *Big) {
	(*kbls.Fr)(dst).Mul((*kbls.Fr)(a), (*kbls.Fr)(b))
}

func InvModBig(dst *Big, v *Big) {
	(*kbls.Fr)(dst).Inverse((*kbls.Fr)(v))
}

//func SqrModBig(dst *Big, v *Big) {
//	kbls.FrSqr((*kbls.Fr)(dst), (*kbls.Fr)(v))
//}
