// +build bignum_hbls

package go_verkle

import (
	hbls "github.com/herumi/bls-eth-go-binary/bls"
	"math/big"
)

var MODULUS Big = bigNum("52435875175126190479447740508185965837690552500527637822603658699938581184513")

type Big *hbls.Fp

func bigNum(v string) Big {
	var p hbls.Fp
	if err := p.SetString(v, 10); err != nil {
		panic(err)
	}
	return &p
}

func asBig(i uint64) Big {
	var p hbls.Fp
	p.SetInt64(int64(i))
	return &p
}

func bigStr(b Big) string {
	return (*hbls.Fp)(b).GetString(10)
}

func equalOne(v Big) bool {
	return (*hbls.Fp)(v).IsOne()
}

func equalZero(v Big) bool {
	return (*hbls.Fp)(v).IsZero()
}

func equalBig(a Big, b Big) bool {
	(*hbls.Fp)(a).IsOne()
	return (*hbls.Fp)(a).IsEqual(b)
}

func subModBigSimple(a Big, b uint8) Big {
	var p hbls.Fp
	hbls.FpSub(&p, a, asBig(uint64(b))) // TODO does it need to mod still?
	return &p
}

func subModBig(a, b Big) Big {
	var p hbls.Fp
	hbls.FpSub(&p, a, b) // TODO does it need to mod still?
	return &p
}

func addModBig(a, b Big) Big {
	var p hbls.Fp
	hbls.FpAdd(&p, a, b) // TODO does it need to mod still?
	return &p
}

func divModBig(a, b Big) Big {
	var p hbls.Fp
	hbls.FpDiv(&p, a, b) // TODO does it need to mod still?
	return &p
}

func mulModBig(a, b Big) Big {
	var p hbls.Fp
	hbls.FpMul(&p, a, b) // TODO does it need to mod still?
	return &p
}

var goMODULUS = _gobigNum("52435875175126190479447740508185965837690552500527637822603658699938581184513")

func _gobigNum(v string) *big.Int {
	var b big.Int
	if err := b.UnmarshalText([]byte(v)); err != nil {
		panic(err)
	}
	return &b
}

// Hacky work around, need to see what's the best way to do a**b in F_p
func powModBig(a, b Big) Big {
	aBytes := (*hbls.Fp)(a).Serialize()
	bBytes := (*hbls.Fp)(b).Serialize()
	var aBig big.Int
	aBig.SetBytes(aBytes)
	var bBig big.Int
	bBig.SetBytes(bBytes)
	var out big.Int
	out.Exp(&aBig, &bBig, goMODULUS)
	var outFp hbls.Fp
	if err := outFp.Deserialize(out.Bytes()); err != nil {
		panic("failed to parse num")
	}
	return &outFp
}
