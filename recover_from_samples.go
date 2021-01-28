package kzg

import (
	"fmt"
)

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

	zeroEval, zeroPoly := zeroPolyFn(missingIndices, uint64(len(samples)))

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
	for i := 0; i < len(evalShiftedReconstructedPoly); i++ {
		divModBig(&evalShiftedReconstructedPoly[i], &evalShiftedPolyWithZero[i], &evalShiftedZeroPoly[i])
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
