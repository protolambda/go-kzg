// +build bignum_blst

package bls

import (
	"crypto/rand"
	"encoding/binary"
	blst "github.com/supranational/blst/bindings/go"
	"math/big"
)

var _modulus big.Int

var oneScalar, zeroScalar blst.Scalar

func init() {
	_modulus.SetString("52435875175126190479447740508185965837690552500527637822603658699938581184513", 10)
	zeroLE := [32]byte{}
	zeroScalar.FromLEndian(zeroLE[:])
	oneLE := [32]byte{0: 1}
	oneScalar.FromLEndian(oneLE[:])
	initGlobals()
	ClearG1(&ZERO_G1)
	initG1G2()
}

type Fr blst.Fr

func SetFr(dst *Fr, v string) {
	var bv big.Int
	bv.SetString(v, 10)
	var sc blst.Scalar
	sc.FromBEndian(bv.Bytes())
	(*blst.Fr)(dst).FromScalar(&sc)
}

// FrFrom32 mutates the fr num. The value v is little-endian 32-bytes.
// Returns false, without modifying dst, if the value is out of range.
func FrFrom32(dst *Fr, v [32]byte) (ok bool) {
	if !ValidFr(v) {
		return false
	}
	var tmp blst.Scalar
	tmp.FromLEndian(v[:])
	(*blst.Fr)(dst).FromScalar(&tmp)
	return true
}

// FrTo32 serializes a fr number to 32 bytes. Encoded little-endian.
func FrTo32(src *Fr) (v [32]byte) {
	var tmp blst.Scalar
	tmp.FromFr((*blst.Fr)(src))
	copy(v[:], tmp.ToLEndian())
	return
}

func CopyFr(dst *Fr, v *Fr) {
	*dst = *v
}

func AsFr(dst *Fr, i uint64) {
	var data [32]byte
	binary.BigEndian.PutUint64(data[:8], i)
	var tmp blst.Scalar
	tmp.FromBEndian(data[:])
	(*blst.Fr)(dst).FromScalar(&tmp)
}

func FrStr(b *Fr) string {
	if b == nil {
		return "<nil>"
	}
	var tmp blst.Scalar
	tmp.FromFr((*blst.Fr)(b))
	var bv big.Int
	bv.SetBytes(tmp.ToBEndian())
	return bv.String()
}

// TODO: would direct memory comparison work here?

func EqualOne(v *Fr) bool {
	var tmp blst.Scalar
	tmp.FromFr((*blst.Fr)(v))
	return tmp.Equals(&oneScalar)
}

func EqualZero(v *Fr) bool {
	var tmp blst.Scalar
	tmp.FromFr((*blst.Fr)(v))
	return tmp.Equals(&zeroScalar)
}

func EqualFr(a *Fr, b *Fr) bool {
	var aSc, bSc blst.Scalar
	aSc.FromFr((*blst.Fr)(a))
	bSc.FromFr((*blst.Fr)(b))
	return aSc.Equals(&bSc)
}

func RandomFr() *Fr {
	v, err := rand.Int(rand.Reader, &_modulus)
	if err != nil {
		panic(err)
	}
	var tmp blst.Scalar
	tmp.FromBEndian(v.Bytes())
	var out blst.Fr
	out.FromScalar(&tmp)
	return (*Fr)(&out)
}

func SubModFr(dst *Fr, a, b *Fr) {
	(*blst.Fr)(dst).Sub((*blst.Fr)(a), (*blst.Fr)(b))
}

func AddModFr(dst *Fr, a, b *Fr) {
	var tmp blst.Fr
	tmp.Add((*blst.Fr)(a), (*blst.Fr)(b))
	*dst = (Fr)(tmp)
}

func DivModFr(dst *Fr, a, b *Fr) {
	var tmp blst.Fr
	tmp.EuclInverse((*blst.Fr)(b))
	(*blst.Fr)(dst).Mul(&tmp, (*blst.Fr)(a))
}

func MulModFr(dst *Fr, a, b *Fr) {
	(*blst.Fr)(dst).Mul((*blst.Fr)(a), (*blst.Fr)(b))
}

func InvModFr(dst *Fr, v *Fr) {
	(*blst.Fr)(dst).EuclInverse((*blst.Fr)(v))
}

//func SqrModFr(dst *Fr, v *Fr) {
//	dst.Sqr(v)
//}

func EvalPolyAt(dst *Fr, p []Fr, x *Fr) {
	// TODO: BLST BLS has no optimized evaluation function
	EvalPolyAtUnoptimized(dst, p, x)
}
