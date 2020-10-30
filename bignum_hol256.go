// +build bignum_hol256

package go_verkle

import (
	u256 "github.com/holiman/uint256"
	"math/big"
)

var _modulus u256.Int

type Big u256.Int

func init() {
	bigNum((*Big)(&_modulus), "52435875175126190479447740508185965837690552500527637822603658699938581184513")
	initGlobals()
}

func bigNum(dst *Big, v string) {
	var b big.Int
	if err := b.UnmarshalText([]byte(v)); err != nil {
		panic(err)
	}
	if overflow := (*u256.Int)(dst).SetFromBig(&b); overflow {
		panic("overflow")
	}
}

func copyBigNum(dst *Big, v *Big) {
	*dst = *v
}

func asBig(dst *Big, i uint64) {
	(*u256.Int)(dst).SetUint64(i)
}

func bigStr(b *Big) string {
	if b == nil {
		return "<nil>"
	}
	return (*u256.Int)(b).ToBig().String()
}

func equalOne(v *Big) bool {
	return *(v) == [4]uint64{0: 1}
}

func equalZero(v *Big) bool {
	return (*u256.Int)(v).IsZero()
}

func equalBig(a *Big, b *Big) bool {
	return (*u256.Int)(a).Eq((*u256.Int)(b))
}

func subModBig(dst *Big, a, b *Big) {
	if (*u256.Int)(dst).SubOverflow((*u256.Int)(a), (*u256.Int)(b)) {
		var tmp u256.Int // hacky
		tmp.Sub(new(u256.Int), (*u256.Int)(dst))
		(*u256.Int)(dst).Sub(&_modulus, &tmp)
	}
}

func addModBig(dst *Big, a, b *Big) {
	(*u256.Int)(dst).AddMod((*u256.Int)(a), (*u256.Int)(b), &_modulus)
}

func divModBig(dst *Big, a, b *Big) {
	(*u256.Int)(dst).Div((*u256.Int)(a), (*u256.Int)(b))
	(*u256.Int)(dst).Mod((*u256.Int)(dst), &_modulus)
}

func mulModBig(dst *Big, a, b *Big) {
	(*u256.Int)(dst).MulMod((*u256.Int)(a), (*u256.Int)(b), &_modulus)
}

// TODO not optimized, but also not used as much
func invModBig(dst *Big, v *Big) {
	// pow(x, n - 2, n)
	var tmp big.Int
	tmp.ModInverse((*u256.Int)(v).ToBig(), (&_modulus).ToBig())
	(*u256.Int)(dst).SetFromBig(&tmp)
}

//func sqrModBig(dst *Big, v *Big) {
//
//}
