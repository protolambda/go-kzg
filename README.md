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
BenchmarkFFTSettings_FFT/scale_4-8         	  301101	     39035 ns/op
BenchmarkFFTSettings_FFT/scale_5
BenchmarkFFTSettings_FFT/scale_5-8         	  141576	     85532 ns/op
BenchmarkFFTSettings_FFT/scale_6
BenchmarkFFTSettings_FFT/scale_6-8         	   62918	    197170 ns/op
BenchmarkFFTSettings_FFT/scale_7
BenchmarkFFTSettings_FFT/scale_7-8         	   28546	    414928 ns/op
BenchmarkFFTSettings_FFT/scale_8
BenchmarkFFTSettings_FFT/scale_8-8         	   13293	    911724 ns/op
BenchmarkFFTSettings_FFT/scale_9
BenchmarkFFTSettings_FFT/scale_9-8         	    5558	   2021055 ns/op
BenchmarkFFTSettings_FFT/scale_10
BenchmarkFFTSettings_FFT/scale_10-8        	    2761	   4240929 ns/op
BenchmarkFFTSettings_FFT/scale_11
BenchmarkFFTSettings_FFT/scale_11-8        	    1323	   9137313 ns/op
BenchmarkFFTSettings_FFT/scale_12
BenchmarkFFTSettings_FFT/scale_12-8        	     609	  19611052 ns/op
BenchmarkFFTSettings_FFT/scale_13
BenchmarkFFTSettings_FFT/scale_13-8        	     286	  41877902 ns/op
BenchmarkFFTSettings_FFT/scale_14
BenchmarkFFTSettings_FFT/scale_14-8        	     134	  89225843 ns/op
BenchmarkFFTSettings_FFT/scale_15
BenchmarkFFTSettings_FFT/scale_15-8        	      61	 190170178 ns/op
```

With Go `big.Int`:
```
BenchmarkFFTSettings_FFT
BenchmarkFFTSettings_FFT/scale_4
BenchmarkFFTSettings_FFT/scale_4-8         	  167884	     71236 ns/op
BenchmarkFFTSettings_FFT/scale_5
BenchmarkFFTSettings_FFT/scale_5-8         	   72451	    162501 ns/op
BenchmarkFFTSettings_FFT/scale_6
BenchmarkFFTSettings_FFT/scale_6-8         	   32625	    366703 ns/op
BenchmarkFFTSettings_FFT/scale_7
BenchmarkFFTSettings_FFT/scale_7-8         	   14629	    820555 ns/op
BenchmarkFFTSettings_FFT/scale_8
BenchmarkFFTSettings_FFT/scale_8-8         	    6410	   1810186 ns/op
BenchmarkFFTSettings_FFT/scale_9
BenchmarkFFTSettings_FFT/scale_9-8         	    3060	   3952192 ns/op
BenchmarkFFTSettings_FFT/scale_10
BenchmarkFFTSettings_FFT/scale_10-8        	    1395	   8558782 ns/op
BenchmarkFFTSettings_FFT/scale_11
BenchmarkFFTSettings_FFT/scale_11-8        	     643	  18657967 ns/op
BenchmarkFFTSettings_FFT/scale_12
BenchmarkFFTSettings_FFT/scale_12-8        	     294	  40827669 ns/op
BenchmarkFFTSettings_FFT/scale_13
BenchmarkFFTSettings_FFT/scale_13-8        	     134	  88518052 ns/op
BenchmarkFFTSettings_FFT/scale_14
BenchmarkFFTSettings_FFT/scale_14-8        	      62	 189773925 ns/op
BenchmarkFFTSettings_FFT/scale_15
BenchmarkFFTSettings_FFT/scale_15-8        	      30	 401412330 ns/op
```

## License

MIT, see [`LICENSE`](./LICENSE) file.

