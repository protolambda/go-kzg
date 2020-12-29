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
- `` (default, empty): use Herumi BLS library, using the `F_p` type. Previously used by `bignum_hbls` build tag.
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

#### Roundtrip FFT (old benches)

Operation: Do `FFT` with `2**scale` values, then do the inverse, and compare all results with the inputs.

With Herumi BLS `F_p`:
```
BenchmarkFFTSettings_FFT/scale_4-8         	  361527	     33671 ns/op
BenchmarkFFTSettings_FFT/scale_5-8         	  159710	     73936 ns/op
BenchmarkFFTSettings_FFT/scale_6-8         	   72572	    163174 ns/op
BenchmarkFFTSettings_FFT/scale_7-8         	   33448	    358415 ns/op
BenchmarkFFTSettings_FFT/scale_8-8         	   15434	    785780 ns/op
BenchmarkFFTSettings_FFT/scale_9-8         	    6999	   1696328 ns/op
BenchmarkFFTSettings_FFT/scale_10-8        	    3296	   3622118 ns/op
BenchmarkFFTSettings_FFT/scale_11-8        	    1518	   7761719 ns/op
BenchmarkFFTSettings_FFT/scale_12-8        	     712	  16653418 ns/op
BenchmarkFFTSettings_FFT/scale_13-8        	     334	  35652441 ns/op
BenchmarkFFTSettings_FFT/scale_14-8        	     157	  75955756 ns/op
BenchmarkFFTSettings_FFT/scale_15-8        	      75	 161703563 ns/op
```

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

