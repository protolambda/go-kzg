// +build !bignum_pure,!bignum_hol256

package kate

import "fmt"

func (fs *FFTSettings) simpleFTG1(vals []G1, valsOffset uint64, valsStride uint64, rootsOfUnity []Big, rootsOfUnityStride uint64, out []G1) {
	l := uint64(len(out))
	var v G1
	var tmp G1
	var last G1
	for i := uint64(0); i < l; i++ {
		jv := &vals[valsOffset]
		r := &rootsOfUnity[0]
		mulG1(&v, jv, r)
		CopyG1(&last, &v)

		for j := uint64(1); j < l; j++ {
			jv := &vals[valsOffset+j*valsStride]
			r := &rootsOfUnity[((i*j)%l)*rootsOfUnityStride]
			mulG1(&v, jv, r)
			CopyG1(&tmp, &last)
			addG1(&last, &tmp, &v)
		}
		CopyG1(&out[i], &last)
	}
}

func (fs *FFTSettings) _fftG1(vals []G1, valsOffset uint64, valsStride uint64, rootsOfUnity []Big, rootsOfUnityStride uint64, out []G1) {
	if len(out) <= 4 { // if the value count is small, run the unoptimized version instead. // TODO tune threshold. (can be different for G1)
		fs.simpleFTG1(vals, valsOffset, valsStride, rootsOfUnity, rootsOfUnityStride, out)
		return
	}

	half := uint64(len(out)) >> 1
	// L will be the left half of out
	fs._fftG1(vals, valsOffset, valsStride<<1, rootsOfUnity, rootsOfUnityStride<<1, out[:half])
	// R will be the right half of out
	fs._fftG1(vals, valsOffset+valsStride, valsStride<<1, rootsOfUnity, rootsOfUnityStride<<1, out[half:]) // just take even again

	var yTimesRoot G1
	var x, y G1
	for i := uint64(0); i < half; i++ {
		// temporary copies, so that writing to output doesn't conflict with input
		CopyG1(&x, &out[i])
		CopyG1(&y, &out[i+half])
		root := &rootsOfUnity[i*rootsOfUnityStride]
		mulG1(&yTimesRoot, &y, root)
		addG1(&out[i], &x, &yTimesRoot)
		subG1(&out[i+half], &x, &yTimesRoot)
	}
}

func (fs *FFTSettings) FFTG1(vals []G1, inv bool) ([]G1, error) {
	n := uint64(len(vals))
	if n > fs.maxWidth {
		return nil, fmt.Errorf("got %d values but only have %d roots of unity", n, fs.maxWidth)
	}
	if !isPowerOfTwo(n) {
		return nil, fmt.Errorf("got %d values but not a power of two", n)
	}
	// We make a copy so we can mutate it during the work.
	valsCopy := make([]G1, n, n)
	for i := 0; i < len(vals); i++ { // TODO: maybe optimize this away, and write back to original input array?
		CopyG1(&valsCopy[i], &vals[i])
	}
	if inv {
		var invLen Big
		asBig(&invLen, n)
		invModBig(&invLen, &invLen)
		rootz := fs.reverseRootsOfUnity[:fs.maxWidth]
		stride := fs.maxWidth / n

		out := make([]G1, n, n)
		fs._fftG1(valsCopy, 0, 1, rootz, stride, out)
		var tmp G1
		for i := 0; i < len(out); i++ {
			mulG1(&tmp, &out[i], &invLen)
			CopyG1(&out[i], &tmp)
		}
		return out, nil
	} else {
		out := make([]G1, n, n)
		rootz := fs.expandedRootsOfUnity[:fs.maxWidth]
		stride := fs.maxWidth / n
		// Regular FFT
		fs._fftG1(valsCopy, 0, 1, rootz, stride, out)
		return out, nil
	}
}

// rearrange G1 elements in reverse bit order. Supports 2**31 max element count.
func reverseBitOrderG1(values []G1) {
	if len(values) > (1 << 31) {
		panic("list too large")
	}
	var tmp G1
	reverseBitOrder(uint32(len(values)), func(i, j uint32) {
		CopyG1(&tmp, &values[i])
		CopyG1(&values[i], &values[j])
		CopyG1(&values[j], &tmp)
	})
}
