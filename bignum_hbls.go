// +build bignum_hbls

package go_verkle

import (
	hbls "github.com/herumi/bls-eth-go-binary/bls"
)

func init() {
	hbls.Init(hbls.BLS12_381)
	initGlobals()
}

type Big hbls.Fr

func bigNum(dst *Big, v string) {
	if err := (*hbls.Fr)(dst).SetString(v, 10); err != nil {
		panic(err)
	}
}

// BigNumFrom31 mutates the big num. The value v is little-endian 31-bytes.
func BigNumFrom31(dst *Big, v [31]byte) {
	(*hbls.Fr)(dst).SetLittleEndian(v[:])
}

// BigNumTo31 serializes a big number to 31 bytes. Any remaining bits after clipped off. Encoded little-endian.
func BigNumTo31(src *Big) (v [31]byte) {
	b := (*hbls.Fr)(src).Serialize()
	if len(b) >= 32 { // clip of bytes beyond 31 bytes (big endian, so from start)
		b = b[len(b)-31:]
	}
	last := len(b) - 1
	half := last / 2
	// reverse endianness, u256.Int outputs big-endian bytes
	for i := 0; i < half; i++ {
		b[i], b[last-i] = b[last-i], b[i]
	}
	copy(v[31-len(b):], b)
	return
}

func CopyBigNum(dst *Big, v *Big) {
	*dst = *v
}

func asBig(dst *Big, i uint64) {
	(*hbls.Fr)(dst).SetInt64(int64(i))
}

func bigStr(b *Big) string {
	if b == nil {
		return "<nil>"
	}
	return (*hbls.Fr)(b).GetString(10)
}

func equalOne(v *Big) bool {
	return (*hbls.Fr)(v).IsOne()
}

func equalZero(v *Big) bool {
	return (*hbls.Fr)(v).IsZero()
}

func equalBig(a *Big, b *Big) bool {
	return (*hbls.Fr)(a).IsEqual((*hbls.Fr)(b))
}

func subModBig(dst *Big, a, b *Big) {
	hbls.FrSub((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func addModBig(dst *Big, a, b *Big) {
	hbls.FrAdd((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func divModBig(dst *Big, a, b *Big) {
	hbls.FrDiv((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func mulModBig(dst *Big, a, b *Big) {
	hbls.FrMul((*hbls.Fr)(dst), (*hbls.Fr)(a), (*hbls.Fr)(b))
}

func invModBig(dst *Big, v *Big) {
	hbls.FrInv((*hbls.Fr)(dst), (*hbls.Fr)(v))
}

//func sqrModBig(dst *Big, v *Big) {
//	hbls.FrSqr((*hbls.Fr)(dst), (*hbls.Fr)(v))
//}
