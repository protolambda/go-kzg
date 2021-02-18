package kzg

import (
	"fmt"
	"github.com/protolambda/go-kzg/bls"
)

func (fs *FFTSettings) simpleFT(vals []bls.Big, valsOffset uint64, valsStride uint64, rootsOfUnity []bls.Big, rootsOfUnityStride uint64, out []bls.Big) {
	l := uint64(len(out))
	var v bls.Big
	var tmp bls.Big
	var last bls.Big
	for i := uint64(0); i < l; i++ {
		jv := &vals[valsOffset]
		r := &rootsOfUnity[0]
		bls.MulModBig(&v, jv, r)
		bls.CopyBigNum(&last, &v)

		for j := uint64(1); j < l; j++ {
			jv := &vals[valsOffset+j*valsStride]
			r := &rootsOfUnity[((i*j)%l)*rootsOfUnityStride]
			bls.MulModBig(&v, jv, r)
			bls.CopyBigNum(&tmp, &last)
			bls.AddModBig(&last, &tmp, &v)
		}
		bls.CopyBigNum(&out[i], &last)
	}
}

func (fs *FFTSettings) _fft(vals []bls.Big, valsOffset uint64, valsStride uint64, rootsOfUnity []bls.Big, rootsOfUnityStride uint64, out []bls.Big) {
	if len(out) <= 4 { // if the value count is small, run the unoptimized version instead. // TODO tune threshold.
		fs.simpleFT(vals, valsOffset, valsStride, rootsOfUnity, rootsOfUnityStride, out)
		return
	}

	half := uint64(len(out)) >> 1
	// L will be the left half of out
	fs._fft(vals, valsOffset, valsStride<<1, rootsOfUnity, rootsOfUnityStride<<1, out[:half])
	// R will be the right half of out
	fs._fft(vals, valsOffset+valsStride, valsStride<<1, rootsOfUnity, rootsOfUnityStride<<1, out[half:]) // just take even again

	var yTimesRoot bls.Big
	var x, y bls.Big
	for i := uint64(0); i < half; i++ {
		// temporary copies, so that writing to output doesn't conflict with input
		bls.CopyBigNum(&x, &out[i])
		bls.CopyBigNum(&y, &out[i+half])
		root := &rootsOfUnity[i*rootsOfUnityStride]
		bls.MulModBig(&yTimesRoot, &y, root)
		bls.AddModBig(&out[i], &x, &yTimesRoot)
		bls.SubModBig(&out[i+half], &x, &yTimesRoot)
	}
}

func (fs *FFTSettings) FFT(vals []bls.Big, inv bool) ([]bls.Big, error) {
	n := uint64(len(vals))
	if n > fs.maxWidth {
		return nil, fmt.Errorf("got %d values but only have %d roots of unity", n, fs.maxWidth)
	}
	n = nextPowOf2(n)
	// We make a copy so we can mutate it during the work.
	valsCopy := make([]bls.Big, n, n)
	for i := 0; i < len(vals); i++ {
		bls.CopyBigNum(&valsCopy[i], &vals[i])
	}
	for i := uint64(len(vals)); i < n; i++ {
		bls.CopyBigNum(&valsCopy[i], &bls.ZERO)
	}
	out := make([]bls.Big, n, n)
	if err := fs.InplaceFFT(valsCopy, out, inv); err != nil {
		return nil, err
	}
	return out, nil
}

func (fs *FFTSettings) InplaceFFT(vals []bls.Big, out []bls.Big, inv bool) error {
	n := uint64(len(vals))
	if n > fs.maxWidth {
		return fmt.Errorf("got %d values but only have %d roots of unity", n, fs.maxWidth)
	}
	if !bls.IsPowerOfTwo(n) {
		return fmt.Errorf("got %d values but not a power of two", n)
	}
	if inv {
		var invLen bls.Big
		bls.AsBig(&invLen, n)
		bls.InvModBig(&invLen, &invLen)
		rootz := fs.reverseRootsOfUnity[:fs.maxWidth]
		stride := fs.maxWidth / n

		fs._fft(vals, 0, 1, rootz, stride, out)
		var tmp bls.Big
		for i := 0; i < len(out); i++ {
			bls.MulModBig(&tmp, &out[i], &invLen)
			bls.CopyBigNum(&out[i], &tmp) // TODO: depending on bignum implementation, allow to directly write back to an input
		}
		return nil
	} else {
		rootz := fs.expandedRootsOfUnity[:fs.maxWidth]
		stride := fs.maxWidth / n
		// Regular FFT
		fs._fft(vals, 0, 1, rootz, stride, out)
		return nil
	}
}

// rearrange Big elements in reverse bit order. Supports 2**31 max element count.
func reverseBitOrderBig(values []bls.Big) {
	if len(values) > (1 << 31) {
		panic("list too large")
	}
	var tmp bls.Big
	reverseBitOrder(uint32(len(values)), func(i, j uint32) {
		bls.CopyBigNum(&tmp, &values[i])
		bls.CopyBigNum(&values[i], &values[j])
		bls.CopyBigNum(&values[j], &tmp)
	})
}

// rearrange Big ptr elements in reverse bit order. Supports 2**31 max element count.
func reverseBitOrderBigPtr(values []*bls.Big) {
	if len(values) > (1 << 31) {
		panic("list too large")
	}
	reverseBitOrder(uint32(len(values)), func(i, j uint32) {
		values[i], values[j] = values[j], values[i]
	})
}
