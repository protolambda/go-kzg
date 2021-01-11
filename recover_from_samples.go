package kate

import (
	"fmt"
	"math/bits"
)

//
//func (fs *FFTSettings) mulPolysDirect(dst []Big, a []Big, b []Big) {
//	if len(b) == 0 || len(a) == 0 || len(dst) != len(b) + len(a) - 1 {
//		panic("invalid usage, inputs must not be zero length, and destination needs to match")
//	}
//	for i := 0; i < len(dst); i++ {
//		CopyBigNum(&dst[i], &ZERO)
//	}
//	var tmp Big
//	for i := 0; i < len(b); i++ {
//		if equalZero(&b[i]) {
//			continue
//		} else if equalOne(&b[i]) {
//			for j := 0; j < len(a); j++ {
//				addModBig(&dst[i+j], &dst[i+j], &a[j])
//			}
//		} else {
//			for j := 0; j < len(a); j++ {
//				mulModBig(&tmp, &b[i], &a[j])
//				addModBig(&dst[i+j], &dst[i+j], &tmp)
//			}
//		}
//	}
//}
//
//func (fs *FFTSettings) mulPolys(dst []Big, a []Big, b []Big) {
//	if len(a) <= 64 || len(b) <= 64 {
//		fs.mulPolysDirect(dst, a, b)
//	} else {
//		n := nextPowOf2(uint64(len(a) + len(b)))
//		stride := fs.maxWidth / n
//		out := truncatePoly(fs.mulPolysWithFFT(a, b, stride))
//		for i := 0; i < len(out); i++ {
//			CopyBigNum(&dst[i], &out[i])
//		}
//		for i := len(out); i < len(dst); i++ {
//			CopyBigNum(&dst[i], &ZERO)
//		}
//	}
//}
//
//func (fs *FFTSettings) mulManyPolys(polys... []Big) []Big {
//	n := uint64(0)
//	for _, p := range polys {
//		l := uint64(len(p))
//		n += l
//	}
//	n = nextPowOf2(n)
//	if n > fs.maxWidth {
//		panic("poly output too large")
//	}
//	prods := make([]Big, n, n)
//	for i := uint64(0); i < n; i++ {
//		CopyBigNum(&prods[i], &ONE)
//	}
//	tmp := make([]Big, n, n)
//	tmpOut := make([]Big, n, n)
//	for _, p := range polys {
//		for i := 0; i < len(p); i++ {
//			CopyBigNum(&tmp[i], &p[i])
//		}
//		for i := uint64(len(p)); i < n; i++ {
//			CopyBigNum(&tmp[i], &ZERO)
//		}
//		if err := fs.InplaceFFT(tmp, tmpOut, false); err != nil {
//			panic(err)
//		}
//		for j := uint64(0); j < n; j++ {
//			mulModBig(&prods[j], &prods[j], &tmpOut[j])
//		}
//	}
//	if err := fs.InplaceFFT(prods, tmpOut, true); err != nil {
//		panic(err)
//	}
//	return tmpOut
//}

// if not already a power of 2, return the next power of 2
func nextPowOf2(v uint64) uint64 {
	if v == 0 {
		return 1
	}
	return uint64(1) << bits.Len64(v-1)
}

//
//func truncatePoly(poly []Big) []Big {
//	deg := polyDegree(poly)
//	return poly[:deg+1]
//}
//
//func polyDegree(poly []Big) int {
//	for i := len(poly)-1; i >= 0; i-- {
//		if !equalZero(&poly[i]) {
//			return i
//		}
//	}
//	return -1
//}

// unshift poly, in-place. Multiplies each coeff with 1/shift_factor**i
func (fs *FFTSettings) ShiftPoly(poly []Big) {
	var shiftFactor Big
	asBig(&shiftFactor, 5) // primitive root of unity
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
	asBig(&shiftFactor, 5) // primitive root of unity
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

	missingIndices := make([]uint64, 0, len(samples))
	for i, s := range samples {
		if s == nil {
			missingIndices = append(missingIndices, uint64(i))
		}
	}

	zeroEval, zeroPoly := zeroPolyFn(missingIndices)

	for i, s := range samples {
		if (s == nil) != equalZero(&zeroEval[i]) {
			panic("bad zero eval")
		}
	}

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
