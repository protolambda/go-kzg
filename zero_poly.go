// Original: https://github.com/ethereum/research/blob/master/polynomial_reconstruction/polynomial_reconstruction.py
// Changes:
// - flattened leaf construction,
// - no aggressive poly truncation
// - simplified merges
// - no heap allocations during reduction

package kzg

import "fmt"

type ZeroPolyFn func(missingIndices []uint64, length uint64) ([]Big, []Big)

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
	if !isPowerOfTwo(n) {
		panic("destination must be a power of two")
	}
	if len(ps) == 0 {
		panic("empty leaves")
	}
	if min := uint64(len(ps[0]) * len(ps)); min > n {
		panic(fmt.Sprintf("expected larger destination length: %d, got: %d", min, n))
	}
	if uint64(len(scratch)) < 3*n {
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
	pEval := scratch[2*n : 3*n]
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
	domainStride := fs.maxWidth / length
	// just under a power of two, since the leaf gets 1 bigger after building a poly for it
	perLeafPoly := uint64(64)
	perLeaf := perLeafPoly - 1
	if uint64(len(missingIndices)) <= perLeaf {
		zeroPoly := make([]Big, len(missingIndices)+1, length)
		fs.makeZeroPolyMulLeaf(zeroPoly, missingIndices, domainStride)
		// pad with zeroes (capacity is already there)
		zeroPoly = zeroPoly[:length]
		zeroEval, err := fs.FFT(zeroPoly, false)
		if err != nil {
			panic(err)
		}
		return zeroEval, zeroPoly
	}

	leafCount := (uint64(len(missingIndices)) + perLeaf - 1) / perLeaf
	n := nextPowOf2(leafCount * perLeafPoly)

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
	scratch := make([]Big, n*3, n*3)

	// from bottom to top, start reducing leaves.
	for len(leaves) > 1 {
		reducedCount := (len(leaves) + reductionFactor - 1) / reductionFactor
		// all the leaves are the same. Except possibly the last leaf, but that's ok.
		leafSize := len(leaves[0])
		for i := 0; i < reducedCount; i++ {
			start := i * reductionFactor
			end := start + reductionFactor
			// E.g. if we *started* with 2 leaves, we won't have more than that since it is already a power of 2.
			// If we had 3, it would have been rounded up anyway. So just pick the end
			outEnd := end * leafSize
			if outEnd > len(out) {
				outEnd = len(out)
			}
			reduced := out[start*leafSize : outEnd]
			// unlike reduced output, input may be smaller than the amount that aligns with powers of two
			if end > len(leaves) {
				end = len(leaves)
			}
			leavesSlice := leaves[start:end]
			if end > start+1 {
				fs.reduceLeaves(scratch, reduced, leavesSlice)
			}
			leaves[i] = reduced
		}
		leaves = leaves[:reducedCount]
	}
	zeroPoly := leaves[0]
	if zl := uint64(len(zeroPoly)); zl < length {
		zeroPoly = append(zeroPoly, make([]Big, length-zl, length-zl)...)
	} else if zl > length {
		panic("expected output smaller or equal to input length")
	}

	zeroEval, err := fs.FFT(zeroPoly, false)
	if err != nil {
		panic(err)
	}

	return zeroEval, zeroPoly
}
