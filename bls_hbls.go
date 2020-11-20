// +build bignum_hbls

package go_verkle

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

type G2 struct {}
