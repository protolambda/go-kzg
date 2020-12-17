// +build bignum_hbls

package kate

import (
	hbls "github.com/herumi/bls-eth-go-binary/bls"
)

func init() {
	hbls.Init(hbls.BLS12_381)
	initGlobals()
	initG1G2()
}

type Big hbls.Fr

func bigNum(dst *Big, v string) {
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
	half := last / 2
	// reverse endianness, u256.Int outputs big-endian bytes
	for i := 0; i < half; i++ {
		b[i], b[last-i] = b[last-i], b[i]
	}
	copy(v[:], b)
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
