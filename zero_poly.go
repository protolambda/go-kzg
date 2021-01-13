// Original: https://github.com/ethereum/research/blob/master/polynomial_reconstruction/polynomial_reconstruction.py
// Changes:
// - flattened leaf construction,
// - no aggressive poly truncation
// - simplified merges
// - no heap allocations during reduction

package kate

import "fmt"

type ZeroPolyFn func(missingIndices []uint64) ([]Big, []Big)

func (fs *FFTSettings) makeZeroPolyMulLeaf(dst []Big, indices []uint64, domainStride uint64) {
	if len(dst) < len(indices)+1 {
		panic(fmt.Sprintf("expected bigger destination length: %d, got: %d", len(indices)+1, len(dst)))
	}
	// zero out the unused slots
	for i := len(indices) + 1; i < len(dst); i++ {
		CopyBigNum(&dst[i], &ZERO)
	}
	CopyBigNum(&dst[len(indices)], &ONE)
	var negDi Big
	for i, v := range indices {
		subModBig(&negDi, &ZERO, &fs.expandedRootsOfUnity[v*domainStride])
		CopyBigNum(&dst[i], &negDi)
		if i > 0 {
			addModBig(&dst[i], &dst[i], &dst[i-1])
			for j := i - 1; j > 0; j-- {
				mulModBig(&dst[j], &dst[j], &negDi)
				addModBig(&dst[j], &dst[j], &dst[j-1])
			}
			mulModBig(&dst[0], &dst[0], &negDi)
		}
	}
}

func (fs *FFTSettings) reduceLeaves(scratch []Big, dst []Big, ps [][]Big) {
	n := uint64(len(dst))
	if len(ps) == 0 {
		panic("empty leaves")
	}
	if min := uint64(len(ps[0]) * len(ps)); min > n {
		panic(fmt.Sprintf("expected larger destination length: %d, got: %d", min, n))
	}
	if uint64(len(scratch)) < 2*n {
		panic("not enough scratch space")
	}
	// TODO: good to optimize, there's lots of padding
	pPadded := scratch[:n]
	prep := func(pi uint64) {
		p := ps[pi]
		for i := 0; i < len(p); i++ {
			CopyBigNum(&pPadded[i], &p[i])
		}
		for i := uint64(len(p)); i < n; i++ {
			CopyBigNum(&pPadded[i], &ZERO)
		}
	}
	mulEvalPs := scratch[n : 2*n]
	// while we're not using dst, use the space for the intermediate results
	pEval := dst
	prep(0)
	if err := fs.InplaceFFT(pPadded, mulEvalPs, false); err != nil {
		panic(err)
	}
	for i := uint64(1); i < uint64(len(ps)); i++ {
		prep(i)
		if err := fs.InplaceFFT(pPadded, pEval, false); err != nil {
			panic(err)
		}
		for j := uint64(0); j < n; j++ {
			mulModBig(&mulEvalPs[j], &mulEvalPs[j], &pEval[j])
		}
	}
	if err := fs.InplaceFFT(mulEvalPs, dst, true); err != nil {
		panic(err)
	}
	return
}

func (fs *FFTSettings) ZeroPolyViaMultiplication(missingIndices []uint64, length uint64) ([]Big, []Big) {
	if len(missingIndices) == 0 {
		return make([]Big, length, length), make([]Big, length, length)
	}
	if length > fs.maxWidth {
		panic("too many ")
	}
	if !isPowerOfTwo(length) {
		panic("length not a power of two")
	}
	// just under a power of two, since the leaf gets 1 bigger after building a poly for it
	perLeaf := uint64(63)
	perLeafPoly := perLeaf + 1
	leafCount := (uint64(len(missingIndices)) + perLeaf - 1) / perLeaf
	n := nextPowOf2(leafCount * perLeafPoly)

	domainStride := fs.maxWidth / length

	// The assumption here is that if the output is a power of two length, matching the sum of child leaf lengths,
	// then the space can be reused.
	out := make([]Big, n, n)

	// Build the leaves.

	// Just the headers, a leaf re-uses the output space.
	// Combining leaves can be done mostly in-place, using a scratchpad.
	leaves := make([][]Big, leafCount, leafCount)

	offset := uint64(0)
	outOffset := uint64(0)
	max := uint64(len(missingIndices))
	for i := uint64(0); i < leafCount; i++ {
		end := offset + perLeaf
		if end > max {
			end = max
		}
		leaves[i] = out[outOffset : outOffset+perLeafPoly]
		fs.makeZeroPolyMulLeaf(leaves[i], missingIndices[offset:end], domainStride)
		offset += perLeaf
		outOffset += perLeafPoly
	}

	// Now reduce all the leaves to a single poly

	// must be a power of 2
	reductionFactor := 4
	scratch := make([]Big, n*2, n*2)

	// from bottom to top, start reducing leaves.
	for len(leaves) > 1 {
		reducedCount := (len(leaves) + reductionFactor - 1) / reductionFactor
		// all the leaves are the same. Except possibly the last leaf, but that's ok.
		leafSize := len(leaves[0])
		for i := 0; i < reducedCount; i++ {
			start := i * reductionFactor
			end := start + reductionFactor
			if end > int(n)/leafSize {
				end = int(n) / leafSize
			}
			reduced := out[start*leafSize : end*leafSize]
			if end > len(leaves) {
				end = len(leaves)
			}
			if end > start+1 {
				// TODO possible optimization if only 2 values are being reduced, instead of 3 or common 4

				// note: reduced overlaps with leaves[start:end]. Perfect overlap if leaves end-start==reduction_factor
				fs.reduceLeaves(scratch, reduced, leaves[start:end])
			}
			leaves[i] = reduced
		}
		leaves = leaves[:reducedCount]
	}
	zeroPoly := leaves[0]
	// When the input length is really small, we may be dealing with an output larger than we can FFT.
	// Just makes sure it's all zeroes, then truncate it.
	for i := length; i < uint64(len(zeroPoly)); i++ {
		if !equalZero(&zeroPoly[i]) {
			panic(fmt.Sprintf("expected zero coeffs after length %d, got: %s at %d", length, bigStr(&zeroPoly[i]), i))
		}
	}
	if length < uint64(len(zeroPoly)) {
		zeroPoly = zeroPoly[:length]
	}

	zeroEval, err := fs.FFT(zeroPoly, false)
	if err != nil {
		panic(err)
	}

	return zeroEval, zeroPoly
}
