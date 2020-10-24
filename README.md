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

Benchmarks on a `Intel Core i7-6700HQ @ 8x 3.5GHz` with `-test.benchtime=10s`:

With Herumi BLS `F_p`:
```
BenchmarkFFTSettings_FFT
BenchmarkFFTSettings_FFT/scale_4
BenchmarkFFTSettings_FFT/scale_4-8         	  314210	     36656 ns/op
BenchmarkFFTSettings_FFT/scale_5
BenchmarkFFTSettings_FFT/scale_5-8         	  146878	     80245 ns/op
BenchmarkFFTSettings_FFT/scale_6
BenchmarkFFTSettings_FFT/scale_6-8         	   68470	    175622 ns/op
BenchmarkFFTSettings_FFT/scale_7
BenchmarkFFTSettings_FFT/scale_7-8         	   31270	    388467 ns/op
BenchmarkFFTSettings_FFT/scale_8
BenchmarkFFTSettings_FFT/scale_8-8         	   14319	    826379 ns/op
BenchmarkFFTSettings_FFT/scale_9
BenchmarkFFTSettings_FFT/scale_9-8         	    6648	   1773182 ns/op
BenchmarkFFTSettings_FFT/scale_10
BenchmarkFFTSettings_FFT/scale_10-8        	    3196	   3825345 ns/op
BenchmarkFFTSettings_FFT/scale_11
BenchmarkFFTSettings_FFT/scale_11-8        	    1484	   8079814 ns/op
BenchmarkFFTSettings_FFT/scale_12
BenchmarkFFTSettings_FFT/scale_12-8        	     694	  17279690 ns/op
BenchmarkFFTSettings_FFT/scale_13
BenchmarkFFTSettings_FFT/scale_13-8        	     320	  36955078 ns/op
BenchmarkFFTSettings_FFT/scale_14
BenchmarkFFTSettings_FFT/scale_14-8        	     152	  78407904 ns/op
BenchmarkFFTSettings_FFT/scale_15
BenchmarkFFTSettings_FFT/scale_15-8        	      73	 166327867 ns/op
```

With Go `big.Int`:
```
BenchmarkFFTSettings_FFT
BenchmarkFFTSettings_FFT/scale_4
BenchmarkFFTSettings_FFT/scale_4-8         	  148371	     79721 ns/op
BenchmarkFFTSettings_FFT/scale_5
BenchmarkFFTSettings_FFT/scale_5-8         	   66244	    183889 ns/op
BenchmarkFFTSettings_FFT/scale_6
BenchmarkFFTSettings_FFT/scale_6-8         	   28401	    418033 ns/op
BenchmarkFFTSettings_FFT/scale_7
BenchmarkFFTSettings_FFT/scale_7-8         	   12727	    942062 ns/op
BenchmarkFFTSettings_FFT/scale_8
BenchmarkFFTSettings_FFT/scale_8-8         	    5839	   2062949 ns/op
BenchmarkFFTSettings_FFT/scale_9
BenchmarkFFTSettings_FFT/scale_9-8         	    2634	   4561437 ns/op
BenchmarkFFTSettings_FFT/scale_10
BenchmarkFFTSettings_FFT/scale_10-8        	    1192	  10118087 ns/op
BenchmarkFFTSettings_FFT/scale_11
BenchmarkFFTSettings_FFT/scale_11-8        	     552	  21667105 ns/op
BenchmarkFFTSettings_FFT/scale_12
BenchmarkFFTSettings_FFT/scale_12-8        	     252	  47501651 ns/op
BenchmarkFFTSettings_FFT/scale_13
BenchmarkFFTSettings_FFT/scale_13-8        	     100	 103291006 ns/op
BenchmarkFFTSettings_FFT/scale_14
BenchmarkFFTSettings_FFT/scale_14-8        	      54	 221584268 ns/op
BenchmarkFFTSettings_FFT/scale_15
BenchmarkFFTSettings_FFT/scale_15-8        	      25	 468244617 ns/op
```

## License

MIT, see [`LICENSE`](./LICENSE) file.

