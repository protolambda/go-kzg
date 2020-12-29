# Kate and FFT utils

This repo is *super experimental*.

This is an implementation in Go, initially aimed at chunkification and extension of data, 
and building/verifying Kate proofs for the output data.
The Kate proofs are built on top of BLS12-381.

Part of a low-latency data-availability sampling network prototype for Eth2 Phase 1.
See https://github.com/protolambda/eth2-das

Code is based on:
- [KZG Data availability code by Dankrad](https://github.com/ethereum/research/tree/master/kzg_data_availability)
- [Verkle and FFT code by Dankrad and Vitalik](https://github.com/ethereum/research/tree/master/verkle)
- [Reed solomon erasure code recovery with FFTs by Vitalik](https://ethresear.ch/t/reed-solomon-erasure-code-recovery-in-n-log-2-n-time-with-ffts/3039)
- [FFT explainer by Vitalik](https://vitalik.ca/general/2019/05/12/fft.html)
- [Kate explainer by Dankrad](https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html)
- [Kate amortized paper by Dankrad and Dmitry](https://github.com/khovratovich/Kate/blob/master/Kate_amortized.pdf)

Features:
- (I)FFT on `F_r`
- (I)FFT on `G1`
- Specialized FFT for extension of `F_r` data
- Kate
  - commitments
  - generate/verify proof for single point
  - generate/verify proofs for multiple points
  - generate/verify proofs for all points, using FK20
  - generate/verify proofs for ranges (cosets) of points, using FK20
- Data recovery: given an arbitrary subset of data (at least half), recover the rest
- Optimized for Data-availability usage
- Change bignum / BLS with build tags.

## BLS

TODO: working with Herumi BLS currently as it exposes more functionality in Go API than BLST does. Still very limited compared to python.
The BLS functionality is generalized, and simple `G1` and `G2` types are exposed to use it. In the future different BLS libraries could be supported.

## Bignums

The BLS curve order is used for the modulo math, different libraries could be used to provide this functionality.
Note: some of these libraries do not have full BLS functionality, only big-num / uint256. The Kate code will be excluded when compiling with a non-BLS build tag.

Build tag options:
- ` ` (default, empty): use Herumi BLS library, using the `F_p` type. Previously used by `bignum_hbls` build tag.
- `-tags bignum_hol256`: Use the uint256 code that Geth uses, [`holiman/uint256`](https://github.com/holiman/uint256)
- `-tags bignum_pure`: Use the native Go bignum implementation.


## Benchmarks

### FFT benchmarks

Benchmarks on a `Intel Core i7-6700HQ @ 8x 3.5GHz` with `-test.benchtime=10s`:

#### FFT with F_r elements

Using Herumi BLS. A single FFT per operation, of `2**scale` elements. Random numbers of `0...modulus` as inputs.

```
BenchmarkFFTSettings_FFT/scale_4-8         	  777571	     14872 ns/op
BenchmarkFFTSettings_FFT/scale_5-8         	  365899	     35068 ns/op
BenchmarkFFTSettings_FFT/scale_6-8         	  153195	     78093 ns/op
BenchmarkFFTSettings_FFT/scale_7-8         	   69822	    173340 ns/op
BenchmarkFFTSettings_FFT/scale_8-8         	   32451	    372674 ns/op
BenchmarkFFTSettings_FFT/scale_9-8         	   14539	    814604 ns/op
BenchmarkFFTSettings_FFT/scale_10-8        	    6512	   1804612 ns/op
BenchmarkFFTSettings_FFT/scale_11-8        	    3088	   3739216 ns/op
BenchmarkFFTSettings_FFT/scale_12-8        	    1549	   8132423 ns/op
BenchmarkFFTSettings_FFT/scale_13-8        	     687	  17562389 ns/op
BenchmarkFFTSettings_FFT/scale_14-8        	     319	  38035214 ns/op
BenchmarkFFTSettings_FFT/scale_15-8        	     146	  79236893 ns/op
```

Note: about 1/3 of the computation time is CGO call overhead to the native Herumi BLS code.
When moving the Kate code into the native library, or reducing the overhead otherwise, the performance can improve significantly.

#### FFT with G1 points

Using Herumi BLS. A single FFT per operation, of `2**scale` elements. Random full G1 points, generator times random scalar.
```
BenchmarkFFTSettings_FFTG1/scale_4-8               3612     3214052 ns/op
BenchmarkFFTSettings_FFTG1/scale_5-8               1524     7814302 ns/op
BenchmarkFFTSettings_FFTG1/scale_6-8                652    18545635 ns/op
BenchmarkFFTSettings_FFTG1/scale_7-8                282    43225687 ns/op
BenchmarkFFTSettings_FFTG1/scale_8-8                121    97727803 ns/op
BenchmarkFFTSettings_FFTG1/scale_9-8                 52   220462276 ns/op
BenchmarkFFTSettings_FFTG1/scale_10-8                24   486038684 ns/op
BenchmarkFFTSettings_FFTG1/scale_11-8                10  1069652806 ns/op
BenchmarkFFTSettings_FFTG1/scale_12-8                 5  2337882518 ns/op
BenchmarkFFTSettings_FFTG1/scale_13-8                 2  5058462998 ns/op
BenchmarkFFTSettings_FFTG1/scale_14-8                 1  10894399107 ns/op
BenchmarkFFTSettings_FFTG1/scale_15-8                 1  23526792397 ns/op
```

#### Alternative FFT

Operation: A single FFT per operation, of `2**scale` elements. Random numbers of `0...modulus` as inputs.

With Go `big.Int`:
```
BenchmarkFFTSettings_FFT/scale_4-8         	  150244	     81521 ns/op
BenchmarkFFTSettings_FFT/scale_5-8         	   59360	    205111 ns/op
BenchmarkFFTSettings_FFT/scale_6-8         	   26980	    436960 ns/op
BenchmarkFFTSettings_FFT/scale_7-8         	   12319	    972553 ns/op
BenchmarkFFTSettings_FFT/scale_8-8         	    5661	   2128780 ns/op
BenchmarkFFTSettings_FFT/scale_9-8         	    2554	   4695178 ns/op
BenchmarkFFTSettings_FFT/scale_10-8        	    1174	  10349451 ns/op
BenchmarkFFTSettings_FFT/scale_11-8        	     535	  22469941 ns/op
BenchmarkFFTSettings_FFT/scale_12-8        	     243	  49311291 ns/op
BenchmarkFFTSettings_FFT/scale_13-8        	     100	 108367278 ns/op
BenchmarkFFTSettings_FFT/scale_14-8        	      51	 231137653 ns/op
BenchmarkFFTSettings_FFT/scale_15-8        	      24	 490071871 ns/op
```

With Holiman u256:

```
BenchmarkFFTSettings_FFT/scale_4-8         	  723585	     16249 ns/op
BenchmarkFFTSettings_FFT/scale_5-8         	  328123	     37372 ns/op
BenchmarkFFTSettings_FFT/scale_6-8         	  147979	     81936 ns/op
BenchmarkFFTSettings_FFT/scale_7-8         	   65496	    184643 ns/op
BenchmarkFFTSettings_FFT/scale_8-8         	   29562	    407406 ns/op
BenchmarkFFTSettings_FFT/scale_9-8         	   13485	    912428 ns/op
BenchmarkFFTSettings_FFT/scale_10-8        	    5881	   1986363 ns/op
BenchmarkFFTSettings_FFT/scale_11-8        	    2884	   4259827 ns/op
BenchmarkFFTSettings_FFT/scale_12-8        	    1320	   9258334 ns/op
BenchmarkFFTSettings_FFT/scale_13-8        	     558	  19882835 ns/op
BenchmarkFFTSettings_FFT/scale_14-8        	     288	  43007504 ns/op
BenchmarkFFTSettings_FFT/scale_15-8        	     130	  89666704 ns/op
```

And with Go native big numbers:
```
BenchmarkFFTSettings_FFT/scale_4-8         	  276931	     42424 ns/op
BenchmarkFFTSettings_FFT/scale_5-8         	  116278	     96975 ns/op
BenchmarkFFTSettings_FFT/scale_6-8         	   54937	    229615 ns/op
BenchmarkFFTSettings_FFT/scale_7-8         	   23569	    494532 ns/op
BenchmarkFFTSettings_FFT/scale_8-8         	   10000	   1125187 ns/op
BenchmarkFFTSettings_FFT/scale_9-8         	    5050	   2468875 ns/op
BenchmarkFFTSettings_FFT/scale_10-8        	    2275	   5260237 ns/op
BenchmarkFFTSettings_FFT/scale_11-8        	    1048	  11334248 ns/op
BenchmarkFFTSettings_FFT/scale_12-8        	     480	  24877020 ns/op
BenchmarkFFTSettings_FFT/scale_13-8        	     223	  53423275 ns/op
BenchmarkFFTSettings_FFT/scale_14-8        	      99	 115540193 ns/op
BenchmarkFFTSettings_FFT/scale_15-8        	      48	 244955527 ns/op
```

#### Extension

Operation: do an extension of even values to odd values (even values are half the domain of the total).
Then next round applies the same function again, but to the output of the previous round.

The scale is the *extended* width: `2**scale`.

With Herumi BLS `F_p`:
```
BenchmarkFFTExtension/scale_4-8         	 2901286	      8197 ns/op
BenchmarkFFTExtension/scale_5-8         	 1231710	     19712 ns/op
BenchmarkFFTExtension/scale_6-8         	  532922	     46486 ns/op
BenchmarkFFTExtension/scale_7-8         	  227506	    104641 ns/op
BenchmarkFFTExtension/scale_8-8         	  100636	    246322 ns/op
BenchmarkFFTExtension/scale_9-8         	   43652	    557844 ns/op
BenchmarkFFTExtension/scale_10-8        	   19508	   1250852 ns/op
BenchmarkFFTExtension/scale_11-8        	    8983	   2725149 ns/op
BenchmarkFFTExtension/scale_12-8        	    4099	   6055065 ns/op
BenchmarkFFTExtension/scale_13-8        	    1818	  12882515 ns/op
BenchmarkFFTExtension/scale_14-8        	     838	  28595864 ns/op
BenchmarkFFTExtension/scale_15-8        	     385	  61239177 ns/op
```

**Note**: extending using regular FFTs costs more than a single FFT (1 to convert to coeffs, then pad with zeroes, then one 2x the size inverse).
To extend to `2**15` with normal FFTs: `79236893*1.5 = ~ 0.118 seconds`. With specialized extension: `61239177 = ~ 0.0612 seconds`, cutting 48% of the cost.

## License

MIT, see [`LICENSE`](./LICENSE) file.

