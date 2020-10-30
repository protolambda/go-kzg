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

func copyBigNum(dst *Big, v *Big) {
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
