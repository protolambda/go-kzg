// +build !bignum_hbls

package go_verkle

import (
	"math/big"
)

var MODULUS Big = bigNum("52435875175126190479447740508185965837690552500527637822603658699938581184513")

type Big *big.Int

func bigNum(v string) Big {
	var b big.Int
	if err := b.UnmarshalText([]byte(v)); err != nil {
		panic(err)
	}
	return &b
}

func asBig(i uint64) Big {
	return big.NewInt(int64(i))
}

func bigStr(b Big) string {
	return (*big.Int)(b).String()
}

func equalOne(v Big) bool {
	return (*big.Int)(v).Cmp(ONE) == 0
}

func equalZero(v Big) bool {
	return (*big.Int)(v).Cmp(ZERO) == 0
}

func equalBig(a Big, b Big) bool {
	return (*big.Int)(a).Cmp(b) == 0
}

func subModBigSimple(a Big, b uint8) Big {
	var out big.Int
	out.Sub(a, big.NewInt(int64(b)))
	return out.Mod(&out, MODULUS)
}

func subModBig(a, b Big) Big {
	var out big.Int
	out.Sub(a, b)
	return out.Mod(&out, MODULUS)
}

func addModBig(a, b Big) Big {
	var out big.Int
	out.Add(a, b)
	return out.Mod(&out, MODULUS)
}

func divModBig(a, b Big) Big {
	var out big.Int
	out.Div(a, b)
	return out.Mod(&out, MODULUS)
}

func mulModBig(a, b Big) Big {
	var out big.Int
	out.Mul(a, b)
	return out.Mod(&out, MODULUS)
}

func powModBig(a, b Big) Big {
	var out big.Int
	return out.Exp(a, b, MODULUS)
}
