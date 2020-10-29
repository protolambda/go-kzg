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
BenchmarkFFTSettings_FFT/scale_4-8         	  361527	     33671 ns/op
BenchmarkFFTSettings_FFT/scale_5
BenchmarkFFTSettings_FFT/scale_5-8         	  159710	     73936 ns/op
BenchmarkFFTSettings_FFT/scale_6
BenchmarkFFTSettings_FFT/scale_6-8         	   72572	    163174 ns/op
BenchmarkFFTSettings_FFT/scale_7
BenchmarkFFTSettings_FFT/scale_7-8         	   33448	    358415 ns/op
BenchmarkFFTSettings_FFT/scale_8
BenchmarkFFTSettings_FFT/scale_8-8         	   15434	    785780 ns/op
BenchmarkFFTSettings_FFT/scale_9
BenchmarkFFTSettings_FFT/scale_9-8         	    6999	   1696328 ns/op
BenchmarkFFTSettings_FFT/scale_10
BenchmarkFFTSettings_FFT/scale_10-8        	    3296	   3622118 ns/op
BenchmarkFFTSettings_FFT/scale_11
BenchmarkFFTSettings_FFT/scale_11-8        	    1518	   7761719 ns/op
BenchmarkFFTSettings_FFT/scale_12
BenchmarkFFTSettings_FFT/scale_12-8        	     712	  16653418 ns/op
BenchmarkFFTSettings_FFT/scale_13
BenchmarkFFTSettings_FFT/scale_13-8        	     334	  35652441 ns/op
BenchmarkFFTSettings_FFT/scale_14
BenchmarkFFTSettings_FFT/scale_14-8        	     157	  75955756 ns/op
BenchmarkFFTSettings_FFT/scale_15
BenchmarkFFTSettings_FFT/scale_15-8        	      75	 161703563 ns/op
```

With Go `big.Int`:
```
BenchmarkFFTSettings_FFT
BenchmarkFFTSettings_FFT/scale_4
BenchmarkFFTSettings_FFT/scale_4-8         	  150244	     81521 ns/op
BenchmarkFFTSettings_FFT/scale_5
BenchmarkFFTSettings_FFT/scale_5-8         	   59360	    205111 ns/op
BenchmarkFFTSettings_FFT/scale_6
BenchmarkFFTSettings_FFT/scale_6-8         	   26980	    436960 ns/op
BenchmarkFFTSettings_FFT/scale_7
BenchmarkFFTSettings_FFT/scale_7-8         	   12319	    972553 ns/op
BenchmarkFFTSettings_FFT/scale_8
BenchmarkFFTSettings_FFT/scale_8-8         	    5661	   2128780 ns/op
BenchmarkFFTSettings_FFT/scale_9
BenchmarkFFTSettings_FFT/scale_9-8         	    2554	   4695178 ns/op
BenchmarkFFTSettings_FFT/scale_10
BenchmarkFFTSettings_FFT/scale_10-8        	    1174	  10349451 ns/op
BenchmarkFFTSettings_FFT/scale_11
BenchmarkFFTSettings_FFT/scale_11-8        	     535	  22469941 ns/op
BenchmarkFFTSettings_FFT/scale_12
BenchmarkFFTSettings_FFT/scale_12-8        	     243	  49311291 ns/op
BenchmarkFFTSettings_FFT/scale_13
BenchmarkFFTSettings_FFT/scale_13-8        	     100	 108367278 ns/op
BenchmarkFFTSettings_FFT/scale_14
BenchmarkFFTSettings_FFT/scale_14-8        	      51	 231137653 ns/op
BenchmarkFFTSettings_FFT/scale_15
BenchmarkFFTSettings_FFT/scale_15-8        	      24	 490071871 ns/op
```

And a quick naive benchmark of the unoptimized python code:
```
scale_4            200 ops          276574 ns/op
scale_5            200 ops          413917 ns/op
scale_6            200 ops          772979 ns/op
scale_7            200 ops         1701269 ns/op
scale_8            200 ops         3290780 ns/op
scale_9            200 ops         7027640 ns/op
scale_10           200 ops        15122420 ns/op
scale_11           200 ops        32552731 ns/op
``` 

For scale 11 (i.e. width `2**11=2048 bignums`), the difference is: `32552731 / 7761719 = ~ 4`.
So HBLS is about 4 times faster than the Python code.

#### Extension

Operation: do an extension of even values to odd values (even values are half the domain of the total).
Then next round apply the same function again, but to the odd values.

With Herumi BLS `F_p`:
```
BenchmarkFFTExtension/scale_4
BenchmarkFFTExtension/scale_4-8         	 1646263	      7287 ns/op
BenchmarkFFTExtension/scale_5
BenchmarkFFTExtension/scale_5-8         	  588529	     19247 ns/op
BenchmarkFFTExtension/scale_6
BenchmarkFFTExtension/scale_6-8         	  250412	     47733 ns/op
BenchmarkFFTExtension/scale_7
BenchmarkFFTExtension/scale_7-8         	  105211	    114526 ns/op
BenchmarkFFTExtension/scale_8
BenchmarkFFTExtension/scale_8-8         	   45291	    267046 ns/op
BenchmarkFFTExtension/scale_9
BenchmarkFFTExtension/scale_9-8         	   19873	    608291 ns/op
BenchmarkFFTExtension/scale_10
BenchmarkFFTExtension/scale_10-8        	    8538	   1373925 ns/op
BenchmarkFFTExtension/scale_11
BenchmarkFFTExtension/scale_11-8        	    3877	   3020615 ns/op
BenchmarkFFTExtension/scale_12
BenchmarkFFTExtension/scale_12-8        	    1816	   6736797 ns/op
BenchmarkFFTExtension/scale_13
BenchmarkFFTExtension/scale_13-8        	     806	  14644598 ns/op
BenchmarkFFTExtension/scale_14
BenchmarkFFTExtension/scale_14-8        	     376	  31621359 ns/op
BenchmarkFFTExtension/scale_15
BenchmarkFFTExtension/scale_15-8        	     176	  67786045 ns/op
```

With Go `big.Int`:
```
BenchmarkFFTExtension
BenchmarkFFTExtension/scale_4
BenchmarkFFTExtension/scale_4-8         	  575293	     20007 ns/op
BenchmarkFFTExtension/scale_5
BenchmarkFFTExtension/scale_5-8         	  210736	     56757 ns/op
BenchmarkFFTExtension/scale_6
BenchmarkFFTExtension/scale_6-8         	   81622	    147382 ns/op
BenchmarkFFTExtension/scale_7
BenchmarkFFTExtension/scale_7-8         	   32914	    365113 ns/op
BenchmarkFFTExtension/scale_8
BenchmarkFFTExtension/scale_8-8         	   13926	    861853 ns/op
BenchmarkFFTExtension/scale_9
BenchmarkFFTExtension/scale_9-8         	    5916	   1997615 ns/op
BenchmarkFFTExtension/scale_10
BenchmarkFFTExtension/scale_10-8        	    2451	   4555595 ns/op
BenchmarkFFTExtension/scale_11
BenchmarkFFTExtension/scale_11-8        	    1162	  10188049 ns/op
BenchmarkFFTExtension/scale_12
BenchmarkFFTExtension/scale_12-8        	     524	  22792625 ns/op
BenchmarkFFTExtension/scale_13
BenchmarkFFTExtension/scale_13-8        	     232	  51342911 ns/op
BenchmarkFFTExtension/scale_14
BenchmarkFFTExtension/scale_14-8        	     100	 111619251 ns/op
BenchmarkFFTExtension/scale_15
BenchmarkFFTExtension/scale_15-8        	      50	 235745595 ns/op
```

## License

MIT, see [`LICENSE`](./LICENSE) file.

