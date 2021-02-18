package kzg

import (
	"fmt"
	"github.com/protolambda/go-kzg/bls"
	"math/rand"
	"testing"
)

func TestErasureCodeRecoverSimple(t *testing.T) {
	// Create some random data, with padding...
	fs := NewFFTSettings(5)
	poly := make([]bls.Fr, fs.maxWidth, fs.maxWidth)
	for i := uint64(0); i < fs.maxWidth/2; i++ {
		bls.AsFr(&poly[i], i)
	}
	for i := fs.maxWidth / 2; i < fs.maxWidth; i++ {
		poly[i] = bls.ZERO
	}
	debugFrs("poly", poly)
	// Get data for polynomial SLOW_INDICES
	data, err := fs.FFT(poly, false)
	if err != nil {
		t.Fatal(err)
	}
	debugFrs("data", data)

	// copy over the 2nd half, leave the first half as nils
	subset := make([]*bls.Fr, fs.maxWidth, fs.maxWidth)
	half := fs.maxWidth / 2
	for i := half; i < fs.maxWidth; i++ {
		subset[i] = &data[i]
	}

	debugFrPtrs("subset", subset)
	recovered, err := fs.ErasureCodeRecover(subset)
	if err != nil {
		t.Fatal(err)
	}
	debugFrs("recovered", recovered)
	for i := range recovered {
		if got := &recovered[i]; !bls.EqualFr(got, &data[i]) {
			t.Errorf("recovery at index %d got %s but expected %s", i, bls.FrStr(got), bls.FrStr(&data[i]))
		}
	}
	// And recover the original coeffs for good measure
	back, err := fs.FFT(recovered, true)
	if err != nil {
		t.Fatal(err)
	}
	debugFrs("back", back)
	for i := uint64(0); i < half; i++ {
		if got := &back[i]; !bls.EqualFr(got, &poly[i]) {
			t.Errorf("coeff at index %d got %s but expected %s", i, bls.FrStr(got), bls.FrStr(&poly[i]))
		}
	}
	for i := half; i < fs.maxWidth; i++ {
		if got := &back[i]; !bls.EqualZero(got) {
			t.Errorf("expected zero padding in index %d", i)
		}
	}
}

func TestErasureCodeRecover(t *testing.T) {
	// Create some random poly, with padding so we get redundant data
	fs := NewFFTSettings(7)
	poly := make([]bls.Fr, fs.maxWidth, fs.maxWidth)
	for i := uint64(0); i < fs.maxWidth/2; i++ {
		bls.AsFr(&poly[i], i)
	}
	for i := fs.maxWidth / 2; i < fs.maxWidth; i++ {
		poly[i] = bls.ZERO
	}
	debugFrs("poly", poly)
	// Get coefficients for polynomial SLOW_INDICES
	data, err := fs.FFT(poly, false)
	if err != nil {
		t.Fatal(err)
	}
	debugFrs("data", data)

	// Util to pick a random subnet of the values
	randomSubset := func(known uint64, rngSeed uint64) []*bls.Fr {
		withMissingValues := make([]*bls.Fr, fs.maxWidth, fs.maxWidth)
		for i := range data {
			withMissingValues[i] = &data[i]
		}
		rng := rand.New(rand.NewSource(int64(rngSeed)))
		missing := fs.maxWidth - known
		pruned := rng.Perm(int(fs.maxWidth))[:missing]
		for _, i := range pruned {
			withMissingValues[i] = nil
		}
		return withMissingValues
	}

	// Try different amounts of known indices, and try it in multiple random ways
	var lastKnown uint64 = 0
	for knownRatio := 0.7; knownRatio < 1.0; knownRatio += 0.05 {
		known := uint64(float64(fs.maxWidth) * knownRatio)
		if known == lastKnown {
			continue
		}
		lastKnown = known
		for i := 0; i < 3; i++ {
			t.Run(fmt.Sprintf("random_subset_%d_known_%d", i, known), func(t *testing.T) {
				subset := randomSubset(known, uint64(i))

				debugFrPtrs("subset", subset)
				recovered, err := fs.ErasureCodeRecover(subset)
				if err != nil {
					t.Fatal(err)
				}
				debugFrs("recovered", recovered)
				for i := range recovered {
					if got := &recovered[i]; !bls.EqualFr(got, &data[i]) {
						t.Errorf("recovery at index %d got %s but expected %s", i, bls.FrStr(got), bls.FrStr(&data[i]))
					}
				}
				// And recover the original coeffs for good measure
				back, err := fs.FFT(recovered, true)
				if err != nil {
					t.Fatal(err)
				}
				debugFrs("back", back)
				half := uint64(len(back)) / 2
				for i := uint64(0); i < half; i++ {
					if got := &back[i]; !bls.EqualFr(got, &poly[i]) {
						t.Errorf("coeff at index %d got %s but expected %s", i, bls.FrStr(got), bls.FrStr(&poly[i]))
					}
				}
				for i := half; i < fs.maxWidth; i++ {
					if got := &back[i]; !bls.EqualZero(got) {
						t.Errorf("expected zero padding in index %d", i)
					}
				}
			})
		}
	}
}
