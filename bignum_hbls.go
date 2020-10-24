// +build bignum_hbls

package go_verkle

import (
	"fmt"
	hbls "github.com/herumi/bls-eth-go-binary/bls"
	"math/big"
)

func init() {
	hbls.Init(hbls.BLS12_381)
	initGlobals()
}

type Big *hbls.Fr

func bigNum(v string) Big {
	var p hbls.Fr
	if err := p.SetString(v, 10); err != nil {
		panic(err)
	}
	return &p
}

func asBig(i uint64) Big {
	var p hbls.Fr
	p.SetInt64(int64(i))
	return &p
}

func bigStr(b Big) string {
	if b == nil {
		return "<nil>"
	}
	return (*hbls.Fr)(b).GetString(10)
}

func equalOne(v Big) bool {
	return (*hbls.Fr)(v).IsOne()
}

func equalZero(v Big) bool {
	return (*hbls.Fr)(v).IsZero()
}

func equalBig(a Big, b Big) bool {
	(*hbls.Fr)(a).IsOne()
	return (*hbls.Fr)(a).IsEqual(b)
}

func subModBigSimple(a Big, b uint8) Big {
	var p hbls.Fr
	hbls.FrSub(&p, a, asBig(uint64(b)))
	return &p
}

func subModBig(a, b Big) Big {
	var p hbls.Fr
	hbls.FrSub(&p, a, b)
	return &p
}

func addModBig(a, b Big) Big {
	var p hbls.Fr
	hbls.FrAdd(&p, a, b)
	return &p
}

func divModBig(a, b Big) Big {
	var p hbls.Fr
	hbls.FrDiv(&p, a, b)
	return &p
}

func mulModBig(a, b Big) Big {
	var p hbls.Fr
	hbls.FrMul(&p, a, b)
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

func invModBig(v Big) Big {
	var p hbls.Fr
	hbls.FrInv(&p, v)
	return &p
}

// Hacky work around, need to see what's the best way to do a**b in F_p
func powModBig(a, b Big) Big {
	aBytes := (*hbls.Fr)(a).Serialize()
	bBytes := (*hbls.Fr)(b).Serialize()
	var aBig big.Int
	aBig.SetBytes(aBytes)
	fmt.Printf("a HBLS: %s\na Big;\n %s\n", bigStr(a), aBig.String())
	var bBig big.Int
	bBig.SetBytes(bBytes)
	var out big.Int
	out.Exp(&aBig, &bBig, goMODULUS)
	var outFr hbls.Fr
	buf := out.Bytes()
	// endianness, super ugly hack
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	if err := outFr.SetLittleEndianMod(buf); err != nil {
		fmt.Printf("out %s\nbig %s\n", bigStr(&outFr), out.String())
		panic(err)
	}
	return &outFr
}
