package go_verkle
import "math/big"

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

type Config struct {
	WIDTH int
}

func simpleFT(vals []Big, modulus Big, rootsOfUnity []Big, out []Big) {
	l := len(rootsOfUnity)
	for i := 0; i < l; i++ {
		last := ZERO
		for j := 0; j < l; j++ {
			v := mulModBig(vals[j], rootsOfUnity[(i*j)%l], modulus)  // TODO lookup could be optimized
			last = addModBig(last, v, modulus)
		}
		out[i] = last
	}
}

// in-place move all even values to the left, all odd values to the right
func reorgValues(vals []Big) (l, r []Big) {
	// all even indices to L
	// all odd indices to R
	l = vals[:len(vals)/2]
	r = vals[:len(vals)-len(l)]
	for i := 0; i < len(vals); {
		l[i >> 1] = vals[i]
		i++
		if i < len(vals) {
			r[i>>1] = vals[i]
			i++
		} else {
			panic("unexpected odd count of vals")
		}
	}
	return l, r
}

func _fft(vals []Big, modulus Big, rootsOfUnity []Big, out []Big) {
	if len(vals) <= 4 {  // if the value count is small, run the unoptimized version instead. // TODO tune threshold.
		simpleFT(vals, modulus, rootsOfUnity, out)
		return
	}

	oldRoots := make([]Big, len(rootsOfUnity), len(rootsOfUnity))
	copy(oldRoots, rootsOfUnity)

	lVals, rVals := reorgValues(vals)
	lROU, rROU := reorgValues(rootsOfUnity)

	half := len(lROU)

	// L will be the left half of out
	_fft(lVals, modulus, lROU, out[:half])
	// R will be the right half of out
	_fft(rVals, modulus, rROU, out[half:]) // TODO: this is lROU in python code?

	for i := 0; i < half; i++ {
		x := out[i]
		y := out[i+half]
		root := oldRoots[i] // TODO how does this work, it only accesses half?
		yTimesRoot := mulModBig(y, root, modulus)
		out[i] = addModBig(x, yTimesRoot, modulus)
		out[i+half] = subModBig(x, yTimesRoot, modulus)
	}
}

func expandRootOfUnity(rootOfUnity Big, modulus Big) []Big {
	rootz := make([]Big, 2, 10) // TODO initial capacity
	rootz[0] = ONE  // some unused number in py code
	rootz[1] = rootOfUnity
	for i := 1; rootz[i] != ONE; i++ {
		rootz = append(rootz, mulModBig(rootz[i], rootOfUnity, modulus))
	}
	return rootz
}


func fft(vals []Big, modulus Big, rootOfUnity Big, inv bool) []Big {
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
	// reverse roots
	for i, j := 0, len(rootz)-1; i < j; i, j = i+1, j-1 {
		rootz[i], rootz[j] = rootz[j], rootz[i]
	}
	if inv {
		exp := subModBigSimple(modulus, 2, modulus)
		invLen := powModBig(asBig(uint64(len(vals))), exp, modulus)
		// Everything except last. The py code does rootz[:0:-1], excluding the last index of reversed slice
		// TODO: currently only FFT regular numbers
		out := make([]Big, len(rootz), len(rootz))
		_fft(vals, modulus, rootz[:len(rootz)-1], out)
		for i := 0; i < len(out); i++ {
			out[i] = mulModBig(out[i], invLen, modulus)
		}
		return out
	} else {
		out := make([]Big, len(rootz), len(rootz))
		// Regular FFT
		_fft(vals, modulus, rootz, out)
		return out
	}
}
