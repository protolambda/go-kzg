// Experimental translation to Go.
// Original:
// - https://github.com/ethereum/research/blob/master/mimc_stark/fft.py
// - https://github.com/ethereum/research/blob/master/mimc_stark/recovery.py

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

func incrementBig(v Big) Big {
	x := (*big.Int)(v)
	return x.Add(x, asBig(1))
}

func addModBig(a, b Big, mod Big) Big {
	var out big.Int
	out.Add(a, b)
	return out.Mod(&out, mod)
}

func rshBig(v Big, sh uint) Big {
	var out big.Int
	return out.Rsh(v, sh)
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
			jv := vals[valsOffset+j*valsStride]
			r := rootsOfUnity[((i*j)%l)*rootsOfUnityStride]
			v := mulModBig(jv, r, modulus) // TODO lookup could be optimized
			last = addModBig(last, v, modulus)
		}
		out[i] = last
	}
}

func _fft(vals []Big, valsOffset uint, valsStride uint, modulus Big, rootsOfUnity []Big, rootsOfUnityStride uint, out []Big) {
	if len(out) <= 4 { // if the value count is small, run the unoptimized version instead. // TODO tune threshold.
		simpleFT(vals, valsOffset, valsStride, modulus, rootsOfUnity, rootsOfUnityStride, out)
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
	valsCopy := make([]Big, len(rootz)-1, len(rootz)-1)
	copy(valsCopy, vals)
	// Fill in vals with zeroes if needed
	for i := len(vals); i < len(rootz)-1; i++ {
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
		_fft(valsCopy, 0, 1, modulus, rootz, 1, out)
		for i := 0; i < len(out); i++ {
			out[i] = mulModBig(out[i], invLen, modulus)
		}
		return out
	} else {
		rootz = rootz[:len(rootz)-1]
		out := make([]Big, len(rootz), len(rootz))
		// Regular FFT
		_fft(valsCopy, 0, 1, modulus, rootz, 1, out)
		return out
	}
}

func mulPolys(a []Big, b []Big, modulus Big, rootOfUnity Big) []Big {
	rootz := expandRootOfUnity(rootOfUnity, modulus)
	rootz = rootz[:len(rootz)-1]
	// pad a and b to match roots of unity
	aVals := make([]Big, len(rootz), len(rootz))
	bVals := make([]Big, len(rootz), len(rootz))
	for i := 0; i < len(a); i++ {
		aVals[i] = a[i]
	}
	for i := len(a); i < len(aVals); i++ {
		aVals[i] = ZERO
	}
	for i := 0; i < len(b); i++ {
		bVals[i] = b[i]
	}
	for i := len(b); i < len(bVals); i++ {
		bVals[i] = ZERO
	}
	// Get FFT of a and b
	x1 := make([]Big, len(aVals), len(aVals))
	_fft(aVals, 0, 1, modulus, rootz, 1, x1)
	x2 := make([]Big, len(bVals), len(bVals))
	_fft(bVals, 0, 1, modulus, rootz, 1, x2)
	// multiply the two. Hack: store results in x1
	for i := 0; i < len(x1); i++ {
		x1[i] = mulModBig(x1[i], x2[i], modulus)
	}
	// compute the FFT of the multiplied values. Hack: store results in x2
	_fft(x1, 0, 1, modulus, rootz, 1, x2)
	return x2
}

// Calculates modular inverses [1/values[0], 1/values[1] ...]
func multiInv(values []Big, modulus Big) []Big {
	partials := make([]Big, len(values)+1, len(values)+1)
	partials[0] = values[0]
	for i := 0; i < len(values); i++ {
		partials[i+1] = mulModBig(partials[i], values[i], modulus)
	}
	exp := subModBigSimple(modulus, 2, modulus)
	inv := powModBig(partials[len(partials)-1], exp, modulus)
	outputs := make([]Big, len(values), len(values))
	for i := len(values); i > 0; i-- {
		outputs[i-1] = mulModBig(partials[i-1], inv, modulus)
		inv = mulModBig(inv, values[i-1], modulus)
	}
	return outputs
}

// Generates q(x) = poly(k * x)
func pOfKX(poly []Big, modulus Big, k Big) []Big {
	out := make([]Big, len(poly), len(poly))
	powerOfK := ONE
	for i, x := range poly {
		out[i] = mulModBig(x, powerOfK, modulus)
		powerOfK = mulModBig(powerOfK, k, modulus)
	}
	return out
}

func inefficientOddEvenDiv2(positions []uint) (even []uint, odd []uint) { // TODO optimize away
	for _, p := range positions {
		if p&1 == 0 {
			even = append(even, p>>1)
		} else {
			odd = append(odd, p>>1)
		}
	}
	return
}

// Return (x - root**positions[0]) * (x - root**positions[1]) * ...
// possibly with a constant factor offset
func _zPoly(positions []uint, modulus Big, rootsOfUnity []Big, rootsOfUnityStride uint) []Big {
	// If there are not more than 4 positions, use the naive
	// O(n^2) algorithm as it is faster
	if len(positions) <= 4 {
		/*
		   root = [1]
		   for pos in positions:
		       x = roots_of_unity[pos]
		       root.insert(0, 0)
		       for j in range(len(root)-1):
		           root[j] -= root[j+1] * x
		   return [x % modulus for x in root]
		*/
		root := make([]Big, len(positions)+1, len(positions)+1)
		root[0] = ONE
		i := 1
		for _, pos := range positions {
			x := rootsOfUnity[pos*rootsOfUnityStride]
			root[i] = ZERO
			for j := i; j >= 1; j-- {
				v := mulModBig(root[j-1], x, modulus)
				root[j] = subModBig(root[j], v, modulus)
			}
			i++
		}
		// We did the reverse representation of 'root' as the python code, to not insert at the start all the time.
		// Now turn it back around.
		for i, j := 0, len(root)-1; i < j; i, j = i+1, j-1 {
			root[i], root[j] = root[j], root[i]
		}
		return root
	}
	// Recursively find the zpoly for even indices and odd
	// indices, operating over a half-size subgroup in each case
	evenPositions, oddPositions := inefficientOddEvenDiv2(positions)
	left := _zPoly(evenPositions, modulus, rootsOfUnity, rootsOfUnityStride<<1)
	right := _zPoly(oddPositions, modulus, rootsOfUnity, rootsOfUnityStride<<1)
	invRoot := rootsOfUnity[uint(len(rootsOfUnity))-rootsOfUnityStride]
	// Offset the result for the odd indices, and combine the two
	out := mulPolys(left, pOfKX(right, modulus, invRoot), modulus, rootsOfUnity[1])
	// Deal with the special case where mul_polys returns zero
	// when it should return x ^ (2 ** k) - 1
	isZero := true
	for _, o := range out {
		if cmpBig(o, ZERO) != 0 {
			isZero = false
			break
		}
	}
	if isZero {
		// TODO: it's [1] + [0] * (len(out) - 1) + [modulus - 1] in python, but strange it's 1 larger than out
		out[0] = ONE
		for i := 1; i < len(out); i++ {
			out[i] = ZERO
		}
		last := subModBigSimple(modulus, 1, modulus)
		out = append(out, last)
		return out
	} else {
		return out
	}
}

func zPoly(positions []uint, modulus Big, rootOfUnity Big) []Big {
	// Precompute roots of unity
	rootz := expandRootOfUnity(rootOfUnity, modulus)
	rootz = rootz[:len(rootz)-1]
	return _zPoly(positions, modulus, rootz, 1)
}

// TODO test unhappy case
const maxRecoverAttempts = 10

func ErasureCodeRecover(vals []Big, modulus Big, rootOfUnity Big) []Big {
	// Generate the polynomial that is zero at the roots of unity
	// corresponding to the indices where vals[i] is None
	positions := make([]uint, 0, len(vals))
	for i := uint(0); i < uint(len(vals)); i++ {
		if vals[i] == nil {
			positions = append(positions, i)
		}
	}
	z := zPoly(positions, modulus, rootOfUnity)
	zVals := FFT(z, modulus, rootOfUnity, false)

	// Pointwise-multiply (vals filling in zero at missing spots) * z
	// By construction, this equals vals * z
	pTimesZVals := make([]Big, len(vals), len(vals))
	for i := uint(0); i < uint(len(vals)); i++ {
		if vals[i] == nil {
			// 0 * zVals[i] == 0
			pTimesZVals[i] = ZERO
		} else {
			pTimesZVals[i] = mulModBig(vals[i], zVals[i], modulus)
		}
	}
	pTimesZ := FFT(pTimesZVals, modulus, rootOfUnity, true)

	// Keep choosing k values until the algorithm does not fail
	// Check only with primitive roots of unity

	expMin1 := subModBigSimple(modulus, 1, modulus)
	expMin1Div2 := rshBig(expMin1, 1)

	expMin2 := subModBigSimple(modulus, 2, modulus)

	attempts := 0
	for k := asBig(2); cmpBig(k, modulus) < 0; k = incrementBig(k) {
		if cmpBig(powModBig(k, expMin1Div2, modulus), ONE) == 0 {
			continue
		}
		invk := powModBig(k, expMin2, modulus)
		// Convert p_times_z(x) and z(x) into new polynomials
		// q1(x) = p_times_z(k*x) and q2(x) = z(k*x)
		// These are likely to not be 0 at any of the evaluation points.
		pTimesZOfKX := pOfKX(pTimesZ, modulus, k)
		pTimesZOfKXVals := FFT(pTimesZOfKX, modulus, rootOfUnity, false)
		zOfKX := pOfKX(z, modulus, k)
		zOfKXVals := FFT(zOfKX, modulus, rootOfUnity, false)
		// Compute q1(x) / q2(x) = p(k*x)
		invZOfKXVals := multiInv(zOfKXVals, modulus)
		pOfKxVals := make([]Big, len(pTimesZOfKXVals), len(pTimesZOfKXVals))
		for i := 0; i < len(pOfKxVals); i++ {
			pOfKxVals[i] = mulModBig(pTimesZOfKXVals[i], invZOfKXVals[i], modulus)
		}
		pOfKx := FFT(pOfKxVals, modulus, rootOfUnity, true)

		// Given q3(x) = p(k*x), recover p(x)
		pOfX := make([]Big, len(pOfKx), len(pOfKx))
		for i, x := range pOfKx {
			pOfX[i] = mulModBig(x, powModBig(invk, asBig(uint64(i)), modulus), modulus)
		}
		output := FFT(pOfX, modulus, rootOfUnity, false)

		// Check that the output matches the input
		success := true
		for i, inpd := range vals {
			if inpd == nil {
				continue
			}
			if cmpBig(inpd, output[i]) != 0 {
				success = false
				break
			}
		}

		if !success {
			attempts += 1
			if attempts >= maxRecoverAttempts {
				panic("bad inputs")
			}
			continue
		}
		// Output the evaluations if all good
		return output
	}
	panic("unreachable")
}
