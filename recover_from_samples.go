package kate

import (
	"fmt"
	"math/bits"
)

//func (fs *FFTSettings) PolyQuotientRemainder(a []Big, b []Big) []Big {
//
//}
//
//func (fs *FFTSettings) FastExtendedEuclideanAlgo(a []Big, b []Big) []Big {
//	degreeA := polyDegree(a)
//	degreeB := polyDegree(b)
//	if degreeA == degreeB {
//		remainder := fs.PolyQuotientRemainder(a, b)
//		return fs.FastExtendedEuclideanAlgo(b, remainder)
//	} else if degreeB == 0 || degreeA > degreeB {
//		m := fs.Mgcd(a, b)
//
//	} else {
//		return fs.FastExtendedEuclideanAlgo(b, a)
//	}
//}
//
//type ZeroPolyFn func(zeroVector []bool) ([]Big, []Big)
//
//// zeroVector input: true for 1, false for 0
//func (fs *FFTSettings) ZeroPolynomialViaGCD(zeroVector []bool) ([]Big, []Big) {
//	// TODO: change to in-place, and re-order steps, can alloc much less that way
//	e1 := make([]Big, len(zeroVector), len(zeroVector))
//	e2 := make([]Big, len(zeroVector), len(zeroVector))
//	for i, v := range zeroVector {
//		if !v {
//			CopyBigNum(&e1[i], &ZERO)
//			CopyBigNum(&e2[i], &ZERO)
//		} else {
//			if i&1 == 0 {
//				CopyBigNum(&e1[i], &ZERO)
//				CopyBigNum(&e2[i], &ONE)
//			} else {
//				CopyBigNum(&e1[i], &ONE)
//				CopyBigNum(&e2[i], &ZERO)
//			}
//		}
//	}
//	p1, err := fs.FFT(e1, true)
//	if err != nil {
//		panic(err)
//	}
//	p2, err := fs.FFT(e2, true)
//	if err != nil {
//		panic(err)
//	}
//	zeroPoly := fs.FastEuclideanAlgo(p1, p2)
//
//	// quickly check the zero poly degree
//	{
//		zeroCount := 0
//		for _, v := range zeroVector {
//			if !v {
//				zeroCount += 1
//			}
//		}
//		degree := polyDegree(zeroPoly)
//		if degree != zeroCount {
//			panic(fmt.Sprintf("unexpected difference between zero poly degree %d and zero count %d", degree, zeroCount))
//		}
//	}
//
//	r, err := fs.FFT(zeroPoly, false)
//	for i, a := range zeroVector {
//		b := &r[i]
//		if !((!a && equalZero(b)) || (a && !equalZero(b))) {
//			panic(fmt.Sprintf("recovery output mismatch: index: %d, zeroVector[i]: %v, result: %d", i, a, bigStr(b)))
//		}
//	}
//	return r, zeroPoly
//}


type ZeroPolyFn func (zeroVector []bool) ([]Big, []Big)

func (fs *FFTSettings) mulPolysDirect(dst []Big, a []Big, b []Big) {
	if len(b) == 0 || len(a) == 0 || len(dst) != len(b) + len(a) - 1 {
		panic("invalid usage, inputs must not be zero length, and destination needs to match")
	}
	for i := 0; i < len(dst); i++ {
		CopyBigNum(&dst[i], &ZERO)
	}
	var tmp Big
	for i := 0; i < len(b); i++ {
		if equalZero(&b[i]) {
			continue
		} else if equalOne(&b[i]) {
			for j := 0; j < len(a); j++ {
				addModBig(&dst[i+j], &dst[i+j], &a[j])
			}
		} else {
			for j := 0; j < len(a); j++ {
				mulModBig(&tmp, &b[i], &a[j])
				addModBig(&dst[i+j], &dst[i+j], &tmp)
			}
		}
	}
}

func (fs *FFTSettings) mulPolys(dst []Big, a []Big, b []Big) {
	if len(a) <= 64 || len(b) <= 64 {
		fs.mulPolysDirect(dst, a, b)
	} else {
		n := nextPowOf2(uint64(len(a) + len(b)))
		stride := fs.maxWidth / n
		out := truncatePoly(fs.mulPolysWithFFT(a, b, stride))
		for i := 0; i < len(out); i++ {
			CopyBigNum(&dst[i], &out[i])
		}
		for i := len(out); i < len(dst); i++ {
			CopyBigNum(&dst[i], &ZERO)
		}
	}
}

func (fs *FFTSettings) mulManyPolys(polys... []Big) []Big {
	n := uint64(0)
	for _, p := range polys {
		l := uint64(len(p))
		n += l
	}
	n = nextPowOf2(n)
	if n > fs.maxWidth {
		panic("poly output too large")
	}
	prods := make([]Big, n, n)
	for i := uint64(0); i < n; i++ {
		CopyBigNum(&prods[i], &ONE)
	}
	tmp := make([]Big, n, n)
	tmpOut := make([]Big, n, n)
	for _, p := range polys {
		for i := 0; i < len(p); i++ {
			CopyBigNum(&tmp[i], &p[i])
		}
		for i := uint64(len(p)); i < n; i++ {
			CopyBigNum(&tmp[i], &ZERO)
		}
		if err := fs.InplaceFFT(tmp, tmpOut, false); err != nil {
			panic(err)
		}
		for j := uint64(0); j < n; j++ {
			mulModBig(&prods[j], &prods[j], &tmpOut[j])
		}
	}
	if err := fs.InplaceFFT(prods, tmpOut, true); err != nil {
		panic(err)
	}
	return tmpOut
}

// if not already a power of 2, return the next power of 2
func nextPowOf2(v uint64) uint64 {
	if v == 0 {
		return 1
	}
	return uint64(1) << bits.Len64(v-1)
}

func (fs *FFTSettings) ZeroPolyViaMultiplication(zeroVector []bool) ([]Big, []Big) {
	//n := nextPowOf2(uint64(len(zeroVector)))
	//domainStride := fs.maxWidth / n

	//ps := [][]Big{make([]Big, 0, n / 64)}
	//
	//for i, x := range zeroVector {
	//	if !x {
	//		term := []Big{ZERO, ONE}
	//		subModBig(&term[0], &ZERO, &fs.expandedRootsOfUnity[uint64(i)*domainStride])
	//		if len(ps) > 0 && polyDegree(ps[len(ps)-1]) < 63 {
	//			ps[len(ps)-1] = fs.mulPolys()ps[len(ps)-1]
	//		}
	//	}
	//}
	//


	return nil, nil // TODO
}

func truncatePoly(poly []Big) []Big {
	deg := polyDegree(poly)
	return poly[:deg+1]
}

func polyDegree(poly []Big) int {
	for i := len(poly)-1; i >= 0; i-- {
		if !equalZero(&poly[i]) {
			return i
		}
	}
	return -1
}

// unshift poly, in-place. Multiplies each coeff with 1/shift_factor**i
func (fs *FFTSettings) ShiftPoly(poly []Big) {
	var shiftFactor Big
	asBig(&shiftFactor, 5)  // primitive root of unity
	var factorPower Big
	CopyBigNum(&factorPower, &ONE)
	var invFactor Big
	invModBig(&invFactor, &shiftFactor)
	var tmp Big
	for i := 0; i < len(poly); i++ {
		CopyBigNum(&tmp, &poly[i])
		mulModBig(&poly[i], &tmp, &factorPower)
		// TODO: pre-compute all these shift scalars
		CopyBigNum(&tmp, &factorPower)
		mulModBig(&factorPower, &tmp, &invFactor)
	}
}

// unshift poly, in-place. Multiplies each coeff with shift_factor**i
func (fs *FFTSettings) UnshiftPoly(poly []Big) {
	var shiftFactor Big
	asBig(&shiftFactor, 5)  // primitive root of unity
	var factorPower Big
	CopyBigNum(&factorPower, &ONE)
	var tmp Big
	for i := 0; i < len(poly); i++ {
		CopyBigNum(&tmp, &poly[i])
		mulModBig(&poly[i], &tmp, &factorPower)
		// TODO: pre-compute all these shift scalars
		CopyBigNum(&tmp, &factorPower)
		mulModBig(&factorPower, &tmp, &shiftFactor)
	}
}

func (fs *FFTSettings) RecoverPolyFromSamples(samples []*Big, zeroPolyFn ZeroPolyFn) ([]Big, error) {
	// TODO: using a single additional temporary array, all the FFTs can run in-place.

	zeroVector := make([]bool, len(samples), len(samples))
	for i, s := range samples {
		zeroVector[i] = s != nil
	}

	zeroEval, zeroPoly := zeroPolyFn(zeroVector)
	polyEvaluationsWithZero := make([]Big, len(samples), len(samples))
	for i, s := range samples {
		if s == nil {
			CopyBigNum(&polyEvaluationsWithZero[i], &ZERO)
		} else {
			mulModBig(&polyEvaluationsWithZero[i], s, &zeroEval[i])
		}
	}
	polyWithZero, err := fs.FFT(polyEvaluationsWithZero, true)
	if err != nil {
		return nil, err
	}
	// shift in-place
	fs.ShiftPoly(polyWithZero)
	shiftedPolyWithZero := polyWithZero

	fs.ShiftPoly(zeroPoly)
	shiftedZeroPoly := zeroPoly

	evalShiftedPolyWithZero, err := fs.FFT(shiftedPolyWithZero, false)
	if err != nil {
		return nil, err
	}
	evalShiftedZeroPoly, err := fs.FFT(shiftedZeroPoly, false)
	if err != nil {
		return nil, err
	}

	evalShiftedReconstructedPoly := evalShiftedPolyWithZero
	var tmp Big
	for i := 0; i < len(evalShiftedReconstructedPoly); i++ {
		divModBig(&tmp, &evalShiftedPolyWithZero[i], &evalShiftedZeroPoly[i])
	}
	shiftedReconstructedPoly, err := fs.FFT(evalShiftedReconstructedPoly, true)
	if err != nil {
		return nil, err
	}
	fs.UnshiftPoly(shiftedReconstructedPoly)
	reconstructedPoly := shiftedReconstructedPoly

	reconstructedData, err := fs.FFT(reconstructedPoly, false)
	if err != nil {
		return nil, err
	}
	for i, s := range samples {
		if s != nil && !equalBig(&reconstructedData[i], s) {
			return nil, fmt.Errorf("failed to reconstruct data correctly, changed value at index %d. Expected: %s, got: %s", i, bigStr(s), bigStr(&reconstructedData[i]))
		}
	}
	return reconstructedData, nil
}

