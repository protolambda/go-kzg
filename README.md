# Verkle utils

Based on research implementation by @dankrad and @vbuterin here: https://github.com/ethereum/research/tree/master/verkle

This repo is *super experimental*.

This is an implementation in Go, initially aimed at chunkification and extension of data, 
and building proofs for the output data. 

Part of a low-latency data-availability sampling network prototype for Phase1.
See https://github.com/protolambda/eth2-das

Also see:
- https://ethresear.ch/t/reed-solomon-erasure-code-recovery-in-n-log-2-n-time-with-ffts/3039
- https://vitalik.ca/general/2019/05/12/fft.html

## Bignums

The BLS curve order is used for the modulo math, the Herumi BLS `F_p` type is can be used as `Big` with the `bignum_hbls` build tag.
By default, Go `big.Int` are used instead.

### FFT benchmarks

#### Roundtrip

Operation: Do `FFT` with `2**scale` values, then do the inverse, and compare all results with the inputs.

With Herumi BLS `F_p`:

Go `big.Int`:
```
BenchmarkFFTSettings_FFT
BenchmarkFFTSettings_FFT/scale_4
BenchmarkFFTSettings_FFT/scale_4-8         	    8472	    141630 ns/op
BenchmarkFFTSettings_FFT/scale_5
BenchmarkFFTSettings_FFT/scale_5-8         	    4375	    309654 ns/op
BenchmarkFFTSettings_FFT/scale_6
BenchmarkFFTSettings_FFT/scale_6-8         	    1976	    622838 ns/op
BenchmarkFFTSettings_FFT/scale_7
BenchmarkFFTSettings_FFT/scale_7-8         	     928	   1299832 ns/op
BenchmarkFFTSettings_FFT/scale_8
BenchmarkFFTSettings_FFT/scale_8-8         	     427	   2842466 ns/op
BenchmarkFFTSettings_FFT/scale_9
BenchmarkFFTSettings_FFT/scale_9-8         	     189	   6689181 ns/op
```

## License

MIT, see [`LICENSE`](./LICENSE) file.

