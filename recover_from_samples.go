package kzg

import (
	"fmt"
	"github.com/protolambda/go-kzg/bls"
)

// unshift poly, in-place. Multiplies each coeff with 1/shift_factor**i
func (fs *FFTSettings) ShiftPoly(poly []bls.Big) {
	var shiftFactor bls.Big
	bls.AsBig(&shiftFactor, 5) // primitive root of unity
	var factorPower bls.Big
	bls.CopyBigNum(&factorPower, &bls.ONE)
	var invFactor bls.Big
	bls.InvModBig(&invFactor, &shiftFactor)
	var tmp bls.Big
	for i := 0; i < len(poly); i++ {
		bls.CopyBigNum(&tmp, &poly[i])
		bls.MulModBig(&poly[i], &tmp, &factorPower)
		// TODO: pre-compute all these shift scalars
		bls.CopyBigNum(&tmp, &factorPower)
		bls.MulModBig(&factorPower, &tmp, &invFactor)
	}
}

// unshift poly, in-place. Multiplies each coeff with shift_factor**i
func (fs *FFTSettings) UnshiftPoly(poly []bls.Big) {
	var shiftFactor bls.Big
	bls.AsBig(&shiftFactor, 5) // primitive root of unity
	var factorPower bls.Big
	bls.CopyBigNum(&factorPower, &bls.ONE)
	var tmp bls.Big
	for i := 0; i < len(poly); i++ {
		bls.CopyBigNum(&tmp, &poly[i])
		bls.MulModBig(&poly[i], &tmp, &factorPower)
		// TODO: pre-compute all these shift scalars
		bls.CopyBigNum(&tmp, &factorPower)
		bls.MulModBig(&factorPower, &tmp, &shiftFactor)
	}
}

func (fs *FFTSettings) RecoverPolyFromSamples(samples []*bls.Big, zeroPolyFn ZeroPolyFn) ([]bls.Big, error) {
	// TODO: using a single additional temporary array, all the FFTs can run in-place.

	missingIndices := make([]uint64, 0, len(samples))
	for i, s := range samples {
		if s == nil {
			missingIndices = append(missingIndices, uint64(i))
		}
	}

	zeroEval, zeroPoly := zeroPolyFn(missingIndices, uint64(len(samples)))

	for i, s := range samples {
		if (s == nil) != bls.EqualZero(&zeroEval[i]) {
			panic("bad zero eval")
		}
	}

	polyEvaluationsWithZero := make([]bls.Big, len(samples), len(samples))
	for i, s := range samples {
		if s == nil {
			bls.CopyBigNum(&polyEvaluationsWithZero[i], &bls.ZERO)
		} else {
			bls.MulModBig(&polyEvaluationsWithZero[i], s, &zeroEval[i])
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
		bls.DivModBig(&evalShiftedReconstructedPoly[i], &evalShiftedPolyWithZero[i], &evalShiftedZeroPoly[i])
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
		if s != nil && !bls.EqualBig(&reconstructedData[i], s) {
			return nil, fmt.Errorf("failed to reconstruct data correctly, changed value at index %d. Expected: %s, got: %s", i, bls.BigStr(s), bls.BigStr(&reconstructedData[i]))
		}
	}
	return reconstructedData, nil
}
