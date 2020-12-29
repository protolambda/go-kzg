// Original: https://github.com/ethereum/research/blob/master/mimc_stark/fft.py

package kate

import (
	"fmt"
)

// Expands the power circle for a given root of unity to WIDTH+1 values.
// The first entry will be 1, the last entry will also be 1,
// for convenience when reversing the array (useful for inverses)
func expandRootOfUnity(rootOfUnity *Big) []Big {
	rootz := make([]Big, 2)
	rootz[0] = ONE // some unused number in py code
	rootz[1] = *rootOfUnity
	for i := 1; !equalOne(&rootz[i]); {
		rootz = append(rootz, Big{})
		this := &rootz[i]
		i++
		mulModBig(&rootz[i], this, rootOfUnity)
	}
	return rootz
}

type FFTSettings struct {
	maxWidth uint64
	// the generator used to get all roots of unity
	rootOfUnity *Big
	// domain, starting and ending with 1 (duplicate!)
	expandedRootsOfUnity []Big
	// reverse domain, same as inverse values of domain. Also starting and ending with 1.
	reverseRootsOfUnity []Big
}

// TODO: generate some setup G1, G2 for testing purposes
// Secret point to evaluate polynomials at.
// Setup values are defined as [g * s**i for i in range(m)]  (correct?)

func NewFFTSettings(maxScale uint8) *FFTSettings {
	width := uint64(1) << maxScale
	root := &scale2RootOfUnity[maxScale]
	rootz := expandRootOfUnity(&scale2RootOfUnity[maxScale])
	// reverse roots of unity
	rootzReverse := make([]Big, len(rootz), len(rootz))
	copy(rootzReverse, rootz)
	for i, j := uint64(0), uint64(len(rootz)-1); i < j; i, j = i+1, j-1 {
		rootzReverse[i], rootzReverse[j] = rootzReverse[j], rootzReverse[i]
	}

	return &FFTSettings{
		maxWidth:             width,
		rootOfUnity:          root,
		expandedRootsOfUnity: rootz,
		reverseRootsOfUnity:  rootzReverse,
	}
}

func (fs *FFTSettings) simpleFT(vals []Big, valsOffset uint64, valsStride uint64, rootsOfUnity []Big, rootsOfUnityStride uint64, out []Big) {
	l := uint64(len(out))
	var v Big
	var tmp Big
	var last Big
	for i := uint64(0); i < l; i++ {
		jv := &vals[valsOffset]
		r := &rootsOfUnity[0]
		mulModBig(&v, jv, r)
		CopyBigNum(&last, &v)

		for j := uint64(1); j < l; j++ {
			jv := &vals[valsOffset+j*valsStride]
			r := &rootsOfUnity[((i*j)%l)*rootsOfUnityStride]
			mulModBig(&v, jv, r)
			CopyBigNum(&tmp, &last)
			addModBig(&last, &tmp, &v)
		}
		CopyBigNum(&out[i], &last)
	}
}

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

func (fs *FFTSettings) _fft(vals []Big, valsOffset uint64, valsStride uint64, rootsOfUnity []Big, rootsOfUnityStride uint64, out []Big) {
	if len(out) <= 4 { // if the value count is small, run the unoptimized version instead. // TODO tune threshold.
		fs.simpleFT(vals, valsOffset, valsStride, rootsOfUnity, rootsOfUnityStride, out)
		return
	}

	half := uint64(len(out)) >> 1
	// L will be the left half of out
	fs._fft(vals, valsOffset, valsStride<<1, rootsOfUnity, rootsOfUnityStride<<1, out[:half])
	// R will be the right half of out
	fs._fft(vals, valsOffset+valsStride, valsStride<<1, rootsOfUnity, rootsOfUnityStride<<1, out[half:]) // just take even again

	var yTimesRoot Big
	var x, y Big
	for i := uint64(0); i < half; i++ {
		// temporary copies, so that writing to output doesn't conflict with input
		CopyBigNum(&x, &out[i])
		CopyBigNum(&y, &out[i+half])
		root := &rootsOfUnity[i*rootsOfUnityStride]
		mulModBig(&yTimesRoot, &y, root)
		addModBig(&out[i], &x, &yTimesRoot)
		subModBig(&out[i+half], &x, &yTimesRoot)
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

func (fs *FFTSettings) FFT(vals []Big, inv bool) ([]Big, error) {
	n := uint64(len(vals))
	if n > fs.maxWidth {
		return nil, fmt.Errorf("got %d values but only have %d roots of unity", n, fs.maxWidth)
	}
	if !isPowerOfTwo(n) {
		return nil, fmt.Errorf("got %d values but not a power of two", n)
	}
	// We make a copy so we can mutate it during the work.
	valsCopy := make([]Big, n, n) // TODO: maybe optimize this away, and write back to original input array?
	for i := uint64(0); i < n; i++ {
		CopyBigNum(&valsCopy[i], &vals[i])
	}
	if inv {
		var invLen Big
		asBig(&invLen, n)
		invModBig(&invLen, &invLen)
		rootz := fs.reverseRootsOfUnity[:fs.maxWidth]
		stride := fs.maxWidth / n

		out := make([]Big, n, n)
		fs._fft(valsCopy, 0, 1, rootz, stride, out)
		var tmp Big
		for i := 0; i < len(out); i++ {
			mulModBig(&tmp, &out[i], &invLen)
			CopyBigNum(&out[i], &tmp) // TODO: depending on bignum implementation, allow to directly write back to an input
		}
		return out, nil
	} else {
		out := make([]Big, n, n)
		rootz := fs.expandedRootsOfUnity[:fs.maxWidth]
		stride := fs.maxWidth / n
		// Regular FFT
		fs._fft(valsCopy, 0, 1, rootz, stride, out)
		return out, nil
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
