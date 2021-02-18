// +build !bignum_pure,!bignum_hol256,!bignum_kilic

package bls

import (
	hbls "github.com/herumi/bls-eth-go-binary/bls"
)

func init() {
	hbls.Init(hbls.BLS12_381)
	initGlobals()
	ClearG1(&ZERO_G1)
	initG1G2()
}

type Big hbls.Fr

func BigNum(dst *Big, v string) {
	if err := (*hbls.Fr)(dst).SetString(v, 10); err != nil {
		panic(err)
	}
}

// BigNumFrom32 mutates the big num. The value v is little-endian 32-bytes.
func BigNumFrom32(dst *Big, v [32]byte) {
	(*hbls.Fr)(dst).SetLittleEndian(v[:])
}

// BigNumTo32 serializes a big number to 32 bytes. Encoded little-endian.
func BigNumTo32(src *Big) (v [32]byte) {
	b := (*hbls.Fr)(src).Serialize()
	last := len(b) - 1
	// reverse endianness, Herumi outputs big-endian bytes
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
	(*hbls.Fr)(dst).SetInt64(int64(i))
}

func BigStr(b *Big) string {
	if b == nil {
		return "<nil>"
	}
	return (*hbls.Fr)(b).GetString(10)
}

func EqualOne(v *Big) bool {
	return (*hbls.Fr)(v).IsOne()
}

func EqualZero(v *Big) bool {
	return (*hbls.Fr)(v).IsZero()
}

func EqualBig(a *Big, b *Big) bool {
	return (*hbls.Fr)(a).IsEqual((*hbls.Fr)(b))
}

func RandomBig() *Big {
	var out hbls.Fr
	out.SetByCSPRNG()
	return (*Big)(&out)
}

func SubModBig(dst *Big, a, b *Big) {
	hbls.FrSub((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func AddModBig(dst *Big, a, b *Big) {
	hbls.FrAdd((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func DivModBig(dst *Big, a, b *Big) {
	hbls.FrDiv((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func MulModBig(dst *Big, a, b *Big) {
	hbls.FrMul((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func InvModBig(dst *Big, v *Big) {
	hbls.FrInv((*hbls.Fr)(dst), (*hbls.Fr)(v))
}

//func SqrModBig(dst *Big, v *Big) {
//	hbls.FrSqr((*hbls.Fr)(dst), (*hbls.Fr)(v))
//}
