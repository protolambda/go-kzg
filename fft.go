package go_verkle

import (
	"fmt"
	"math/big"
	"strings"
)

// TODO big int or F_p type
type Big *big.Int

func bigNum(v string) Big {
	var b big.Int
	if err := b.UnmarshalText([]byte(v)); err != nil {
		panic(err)
	}
	return Big(&b)
}

var rootOfUnityCandidates = map[int]Big{
	512: bigNum("12531186154666751577774347439625638674013361494693625348921624593362229945844"),
	256: bigNum("21071158244812412064791010377580296085971058123779034548857891862303448703672"),
	128: bigNum("3535074550574477753284711575859241084625659976293648650204577841347885064712"),
	64:  bigNum("6460039226971164073848821215333189185736442942708452192605981749202491651199"),
	32:  bigNum("32311457133713125762627935188100354218453688428796477340173861531654182464166"),
	16:  bigNum("35811073542294463015946892559272836998938171743018714161809767624935956676211"),
}

const WIDTH = 16

var MODULUS Big = bigNum("52435875175126190479447740508185965837690552500527637822603658699938581184513")
var ROOT_OF_UNITY Big = rootOfUnityCandidates[WIDTH]
var ROOT_OF_UNITY2 Big = rootOfUnityCandidates[WIDTH*2]

var ZERO = Big(big.NewInt(0))
var ONE = Big(big.NewInt(1))

func asBig(i uint64) Big {
	return big.NewInt(int64(i))
}

func bigStr(b Big) string {
	return (*big.Int)(b).String()
}

func cmpBig(a Big, b Big) int {
	return (*big.Int)(a).Cmp(b)
}

func subModBigSimple(a Big, b uint8, mod Big) Big {
	var out big.Int
	out.Sub(a, big.NewInt(int64(b)))
	return out.Mod(&out, mod)
}

func subModBig(a, b Big, mod Big) Big {
	var out big.Int
	out.Sub(a, b)
	return out.Mod(&out, mod)
}

func addModBig(a, b Big, mod Big) Big {
	var out big.Int
	out.Add(a, b)
	return out.Mod(&out, mod)
}

func mulModBig(a, b Big, mod Big) Big {
	var out big.Int
	out.Mul(a, b)
	return out.Mod(&out, mod)
}

func powModBig(a, b Big, mod Big) Big {
	var out big.Int
	return out.Exp(a, b, mod)
}

func debugBigs(msg string, values []Big) {
	var out strings.Builder
	out.WriteString("---")
	out.WriteString(msg)
	out.WriteString("---\n")
	for i, v := range values {
		out.WriteString(fmt.Sprintf("#%4d: %s\n", i, (*big.Int)(v).String()))
	}
	fmt.Println(out.String())
}

func debugBigsOffsetStride(msg string, values []Big, offset uint, stride uint) {
	var out strings.Builder
	out.WriteString("---")
	out.WriteString(msg)
	out.WriteString("---\n")
	j := uint(0)
	for i := offset; i < uint(len(values)); i += stride {
		out.WriteString(fmt.Sprintf("#%4d: %s\n", j, (*big.Int)(values[i]).String()))
		j++
	}
	fmt.Println(out.String())
}

type Config struct {
	WIDTH int
}

func simpleFT(vals []Big, valsOffset uint, valsStride uint, modulus Big, rootsOfUnity []Big, rootsOfUnityStride uint, out []Big) {
	l := uint(len(out))
	for i := uint(0); i < l; i++ {
		last := ZERO
		for j := uint(0); j < l; j++ {
			v := mulModBig(vals[valsOffset+j*valsStride], rootsOfUnity[((i*j)%l)*rootsOfUnityStride], modulus) // TODO lookup could be optimized
			last = addModBig(last, v, modulus)
		}
		out[i] = last
	}
}

func _fft(vals []Big, valsOffset uint, valsStride uint, modulus Big, rootsOfUnity []Big, rootsOfUnityStride uint, out []Big) {
	if len(out) <= 4 { // if the value count is small, run the unoptimized version instead. // TODO tune threshold.
		simpleFT(vals, valsOffset, valsStride, modulus, rootsOfUnity, rootsOfUnityStride, out) // TODO apply stride
		return
	}

	half := uint(len(out)) >> 1
	// L will be the left half of out
	_fft(vals, valsOffset, valsStride<<1, modulus, rootsOfUnity, rootsOfUnityStride<<1, out[:half])
	// R will be the right half of out
	_fft(vals, valsOffset+valsStride, valsStride<<1, modulus, rootsOfUnity, rootsOfUnityStride<<1, out[half:]) // just take even again

	for i := uint(0); i < half; i++ {
		x := out[i]
		y := out[i+half]
		root := rootsOfUnity[i*rootsOfUnityStride]
		yTimesRoot := mulModBig(y, root, modulus)
		out[i] = addModBig(x, yTimesRoot, modulus)
		out[i+half] = subModBig(x, yTimesRoot, modulus)
	}
}

func expandRootOfUnity(rootOfUnity Big, modulus Big) []Big {
	rootz := make([]Big, 2, 10) // TODO initial capacity
	rootz[0] = ONE              // some unused number in py code
	rootz[1] = rootOfUnity
	for i := 1; (*big.Int)(rootz[i]).Cmp(ONE) != 0; i++ {
		rootz = append(rootz, mulModBig(rootz[i], rootOfUnity, modulus))
	}
	return rootz
}

func FFT(vals []Big, modulus Big, rootOfUnity Big, inv bool) []Big {
	rootz := expandRootOfUnity(rootOfUnity, modulus)
	// We make a copy so we can mutate it during the work.
	valsCopy := make([]Big, len(rootz), len(rootz))
	for i, v := range vals {
		valsCopy[i] = v
	}
	// Fill in vals with zeroes if needed
	for i := len(vals); i < len(rootz); i++ {
		valsCopy[i] = ZERO
	}
	if inv {
		exp := subModBigSimple(modulus, 2, modulus)
		invLen := powModBig(asBig(uint64(len(vals))), exp, modulus)
		// reverse roots of unity
		for i, j := 0, len(rootz)-1; i < j; i, j = i+1, j-1 {
			rootz[i], rootz[j] = rootz[j], rootz[i]
		}
		rootz = rootz[:len(rootz)-1]
		//debugBigs("reversed roots of unity", rootz)
		// TODO: currently only FFT regular numbers
		out := make([]Big, len(rootz), len(rootz))
		_fft(vals, 0, 1, modulus, rootz, 1, out)
		for i := 0; i < len(out); i++ {
			out[i] = mulModBig(out[i], invLen, modulus)
		}
		return out
	} else {
		rootz = rootz[:len(rootz)-1]
		out := make([]Big, len(rootz), len(rootz))
		// Regular FFT
		_fft(vals, 0, 1, modulus, rootz, 1, out)
		return out
	}
}
