// +build bignum_hbls

package kate

import hbls "github.com/herumi/bls-eth-go-binary/bls"

// TODO types file, swap BLS with build args
type G1 hbls.G1

func ClearG1(x *G1) {
	(*hbls.G1)(x).Clear()
}

func CopyG1(dst *G1, v *G1) {
	*dst = *v
}

func mulG1(dst *G1, a *G1, b *Big) {
	hbls.G1Mul((*hbls.G1)(dst), (*hbls.G1)(a), (*hbls.Fr)(b))
}

func addG1(dst *G1, a *G1, b *G1) {
	hbls.G1Add((*hbls.G1)(dst), (*hbls.G1)(a), (*hbls.G1)(b))
}

func subG1(dst *G1, a *G1, b *G1) {
	hbls.G1Sub((*hbls.G1)(dst), (*hbls.G1)(a), (*hbls.G1)(b))
}

type G2 hbls.G2

func ClearG2(x *G2) {
	(*hbls.G2)(x).Clear()
}

func CopyG2(dst *G2, v *G2) {
	*dst = *v
}

func mulG2(dst *G2, a *G2, b *Big) {
	hbls.G2Mul((*hbls.G2)(dst), (*hbls.G2)(a), (*hbls.Fr)(b))
}

func addG2(dst *G2, a *G2, b *G2) {
	hbls.G2Add((*hbls.G2)(dst), (*hbls.G2)(a), (*hbls.G2)(b))
}

func subG2(dst *G2, a *G2, b *G2) {
	hbls.G2Sub((*hbls.G2)(dst), (*hbls.G2)(a), (*hbls.G2)(b))
}
